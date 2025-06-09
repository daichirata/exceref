package cmd

import (
	"errors"
	"strings"

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

	// language specific
	generateCmd.Flags().String("package-name", "model", "[Go] Set output package name")
	generateCmd.Flags().String("tag-name", "json", "[Go] Set output struct tag name")

	generateCmd.MarkFlagRequired("out")
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

	packageName, err := cmd.Flags().GetString("package-name")
	if err != nil {
		return err
	}
	tagName, err := cmd.Flags().GetString("tag-name")
	if err != nil {
		return err
	}

	option := exceref.GenerateOption{
		GoPackageName: packageName,
		GoTagNames:    strings.Split(tagName, ","),
	}

	file, err := exceref.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()

	return file.Generate(exceref.BuildGenerator(lang, outDir, prefix, option))
}
