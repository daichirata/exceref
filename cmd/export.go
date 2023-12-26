package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/daichirata/exceref/internal/exceref"
)

var exportCmd = &cobra.Command{
	Use:  "export",
	RunE: exportFunc,
}

func init() {
	rootCmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("out", "o", "", "Set output directory")
	exportCmd.Flags().StringP("format", "f", "csv", "Set output format")

	exportCmd.MarkFlagRequired("out")
}

func exportFunc(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("FILE needs to be provided")
	}

	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return err
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}

	file, err := exceref.Open(args[0])
	if err != nil {
		return err
	}
	defer file.Close()

	return file.Export(exceref.BuildExporter(format, outDir))
}
