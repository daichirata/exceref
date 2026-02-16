package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"
)

var debug bool

var rootCmd = &cobra.Command{
	Use:           "exceref",
	SilenceUsage:  true,
	SilenceErrors: true,
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
		printErrorChain(err)
		os.Exit(1)
	}
}

func printErrorChain(err error) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", err)

	cause := errors.Unwrap(err)
	for cause != nil {
		fmt.Fprintf(os.Stderr, "Caused by: %s\n", cause)
		cause = errors.Unwrap(cause)
	}
}
