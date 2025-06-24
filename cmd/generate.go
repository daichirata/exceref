package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/daichirata/exceref/internal/exceref"
)

var generateCmd = &cobra.Command{
	Use:  "generate",
	RunE: generateFunc,
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringP("out", "o", "", "Set output directory")
	generateCmd.Flags().StringP("lang", "l", "go", "Set output format")
	generateCmd.Flags().StringP("prefix", "p", "", "Set output model name prefix")
	generateCmd.Flags().StringP("template", "t", "", "Set template path")

	generateCmd.MarkFlagRequired("out")
	generateCmd.MarkFlagRequired("template")
}

func generateFunc(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("FILE needs to be provided")
	}

	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}
	lang, err := cmd.Flags().GetString("lang")
	if err != nil {
		return err
	}
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return err
	}
	templatePath, err := cmd.Flags().GetString("template")
	if err != nil {
		return err
	}

	option := exceref.GenerateOption{
		Prefix:       prefix,
		OutDir:       outDir,
		TemplatePath: templatePath,
	}

	file, err := exceref.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()

	return file.Generate(exceref.BuildGenerator(lang, option))
}
