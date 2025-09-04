package cmd

import (
	"github.com/daichirata/exceref/cmd/meta"
)

func init() {
	rootCmd.AddCommand(meta.Cmd)
}
