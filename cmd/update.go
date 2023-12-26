package cmd

import (
	"errors"

	"github.com/spf13/cobra"

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
		return err
	}
	defer file.Close()

	if err := file.UpdateReferenceData(); err != nil {
		return err
	}
	if err := file.UpdateDataValidations(); err != nil {
		return err
	}
	return file.Save()
}
