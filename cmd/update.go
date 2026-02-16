package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/daichirata/exceref/internal/errs"
	"github.com/daichirata/exceref/internal/exceref"
)

var updateCmd = &cobra.Command{
	Use:  "update",
	RunE: updateFunc,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func updateFunc(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return errors.New("FILENAME needs to be provided")
	}

	file, err := exceref.Open(args[0])
	if err != nil {
		return errs.Wrap(err, "open file")
	}
	defer file.Close()

	if err := file.UpdateReferenceData(); err != nil {
		return errs.Wrap(err, "update reference data")
	}
	if err := file.UpdateDataValidations(); err != nil {
		return errs.Wrap(err, "update data validations")
	}
	return errs.Wrap(file.Save(), "save file")
}
