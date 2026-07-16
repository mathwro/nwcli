package cmd

import (
	"fmt"
	"io"

	"github.com/mathwro/nwcli/internal/selfupdate"
	"github.com/spf13/cobra"
)

var updateInstall = selfupdate.InstallLatest

type updateResult struct {
	Target string `json:"target"`
}

func newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "update",
		Short: "Update nwcli to the latest version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			format, err := cmd.Root().PersistentFlags().GetString("output")
			if err != nil {
				return err
			}
			installStdout := cmd.OutOrStdout()
			if format == "json" {
				installStdout = io.Discard
			}
			runErr := updateInstall(cmd.Context(), installStdout, cmd.ErrOrStderr())
			if runErr != nil {
				runErr = newCommandError(exitUnreachable, runErr)
			}
			return emitCommandResult(cmd, "update", updateResult{Target: selfupdate.Target}, runErr, func(w io.Writer) error {
				_, err := fmt.Fprintln(w, "nwcli updated to the latest available version")
				return err
			})
		},
	}
}
