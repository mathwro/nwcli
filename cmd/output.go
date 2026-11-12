package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"
)

const (
	exitUsage       = 2
	exitUnhealthy   = 3
	exitUnreachable = 4
	exitPermission  = 5
)

type commandError struct {
	code        int
	err         error
	emittedJSON bool
}

func newCommandError(code int, err error) error {
	if err == nil {
		return nil
	}
	return commandError{code: code, err: err}
}

func markEmittedJSON(err error) error {
	if err == nil {
		return nil
	}
	var target commandError
	if errors.As(err, &target) {
		target.emittedJSON = true
		return target
	}
	return commandError{code: ExitCode(err), err: err, emittedJSON: true}
}

func isEmittedJSONError(err error) bool {
	var target commandError
	return errors.As(err, &target) && target.emittedJSON
}

func (e commandError) Error() string { return e.err.Error() }
func (e commandError) Unwrap() error { return e.err }

func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var target commandError
	if errors.As(err, &target) {
		return target.code
	}
	return exitUsage
}

type outputError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type outputEnvelope struct {
	Version int          `json:"version"`
	Command string       `json:"command"`
	OK      bool         `json:"ok"`
	Result  any          `json:"result,omitempty"`
	Error   *outputError `json:"error,omitempty"`
}

func validateOutputFormat(format string) error {
	if format != "text" && format != "json" {
		return newCommandError(exitUsage, fmt.Errorf("unsupported output format %q: use text or json", format))
	}
	return nil
}

func writeJSONOutput(w io.Writer, command string, result any, runErr error) error {
	envelope := outputEnvelope{Version: 1, Command: command, OK: runErr == nil, Result: result}
	if runErr != nil {
		envelope.Error = &outputError{Code: ExitCode(runErr), Message: runErr.Error()}
	}
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	return encoder.Encode(envelope)
}

func emitCommandResult(cmd *cobra.Command, command string, result any, runErr error, renderText func(io.Writer) error) error {
	format, err := cmd.Root().PersistentFlags().GetString("output")
	if err != nil {
		return err
	}
	if format == "json" {
		if err := writeJSONOutput(cmd.OutOrStdout(), command, result, runErr); err != nil {
			return err
		}
		return markEmittedJSON(runErr)
	}
	if renderText != nil {
		if err := renderText(cmd.OutOrStdout()); err != nil {
			return err
		}
	}
	return runErr
}
