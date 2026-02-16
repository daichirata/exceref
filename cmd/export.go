package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/daichirata/exceref/internal/errs"
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
	exportCmd.Flags().StringP("prefix", "p", "", "Set output file name prefix")

	exportCmd.MarkFlagRequired("out")
}

func exportFunc(cmd *cobra.Command, args []string) error {
	if len(args) < 1 || args[0] == "" {
		return errors.New("FILE needs to be provided")
	}

	outDir, err := cmd.Flags().GetString("out")
	if err != nil {
		return errs.Wrap(err, "get out flag")
	}
	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return errs.Wrap(err, "get format flag")
	}
	prefix, err := cmd.Flags().GetString("prefix")
	if err != nil {
		return errs.Wrap(err, "get prefix flag")
	}

	file, err := exceref.Open(args[0])
	if err != nil {
		return errs.Wrap(err, "open file")
	}
	defer file.Close()

	return errs.Wrap(file.Export(exceref.BuildExporter(format, outDir, prefix)), "export sheets")
}
