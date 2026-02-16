package meta

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
	Cmd.AddCommand(exportCmd)

	exportCmd.Flags().StringP("out", "o", "", "Set output directory")

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

	file, err := exceref.Open(args[0])
	if err != nil {
		return errs.Wrap(err, "open file")
	}
	defer file.Close()

	return errs.Wrap(file.ExportMetadata(outDir), "export metadata")
}
