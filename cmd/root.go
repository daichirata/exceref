package cmd

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:          "exceref",
	SilenceUsage: true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if debug {
			slog.SetDefault(slog.New(
				slog.NewTextHandler(
					os.Stdout,
					&slog.HandlerOptions{Level: slog.LevelDebug},
				)),
			)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "debug")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
