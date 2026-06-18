// Package cli wires the cobra command tree and translates outcomes into process
// exit codes: 0 = clean, 2 = drift detected, 1 = operational error.
package cli

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// Version is overridden at build time via -ldflags.
var Version = "dev"

// exitError carries a specific process exit code up to Execute. A drift gate
// returns one with an empty message (drift is not an error to print).
type exitError struct {
	code int
	msg  string
}

func (e *exitError) Error() string { return e.msg }

// Execute runs the root command and returns the process exit code.
func Execute() int {
	root := newRootCmd()
	err := root.Execute()
	if err == nil {
		return 0
	}
	var ee *exitError
	if errors.As(err, &ee) {
		if ee.msg != "" {
			fmt.Fprintln(os.Stderr, "error:", ee.msg)
		}
		return ee.code
	}
	fmt.Fprintln(os.Stderr, "error:", err)
	return 1
}

func newRootCmd() *cobra.Command {
	var logLevel string
	root := &cobra.Command{
		Use:           "tfdrift",
		Short:         "Agentless Terraform drift detection",
		Long:          "tfdrift compares Terraform state against live cloud infrastructure and reports drift,\nwithout running terraform plan or apply.",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRun: func(_ *cobra.Command, _ []string) {
			setupLogging(logLevel)
		},
	}
	root.PersistentFlags().StringVar(&logLevel, "log-level", "info", "log level: debug|info|warn|error")
	root.AddCommand(newScanCmd(), newVersionCmd())
	return root
}

// setupLogging sends structured logs to stderr so stdout stays pure for
// --output json.
func setupLogging(level string) {
	var l slog.Level
	switch strings.ToLower(level) {
	case "debug":
		l = slog.LevelDebug
	case "warn":
		l = slog.LevelWarn
	case "error":
		l = slog.LevelError
	default:
		l = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: l})))
}

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the tfdrift version",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "tfdrift %s\n", Version)
		},
	}
}
