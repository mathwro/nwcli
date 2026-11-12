package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestUpdateCommandInstallsLatestVersion(t *testing.T) {
	original := updateInstall
	t.Cleanup(func() { updateInstall = original })
	called := false
	updateInstall = func(_ context.Context, stdout, stderr io.Writer) error {
		called = true
		if stdout == nil || stderr == nil {
			t.Fatal("update command did not forward output writers")
		}
		return nil
	}

	stdout, _, err := executeForTest("update")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !called {
		t.Fatal("update installer was not called")
	}
	if !strings.Contains(stdout, "nwcli updated to the latest available version") {
		t.Fatalf("stdout = %q, want success confirmation", stdout)
	}
}

func TestUpdateCommandWritesJSONEnvelope(t *testing.T) {
	original := updateInstall
	t.Cleanup(func() { updateInstall = original })
	updateInstall = func(context.Context, io.Writer, io.Writer) error { return nil }

	stdout, _, err := executeForTest("--output", "json", "update")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope outputEnvelope
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal returned error: %v\nstdout: %s", err, stdout)
	}
	if envelope.Command != "update" || !envelope.OK {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestUpdateCommandRejectsArguments(t *testing.T) {
	original := updateInstall
	t.Cleanup(func() { updateInstall = original })
	called := false
	updateInstall = func(context.Context, io.Writer, io.Writer) error {
		called = true
		return nil
	}

	_, _, err := executeForTest("update", "unexpected")
	if err == nil {
		t.Fatal("Execute returned nil error, want argument error")
	}
	if called {
		t.Fatal("installer ran for invalid arguments")
	}
}

func TestUpdateCommandIsNotAnInteractiveAction(t *testing.T) {
	for _, action := range interactiveActions() {
		if strings.EqualFold(action.Title, "update") {
			t.Fatal("update command unexpectedly appears in TUI actions")
		}
	}
}

func TestUpdateCommandUsesUsageExitCodeForPrebuiltReleaseError(t *testing.T) {
	original := updateInstall
	t.Cleanup(func() { updateInstall = original })
	updateInstall = func(context.Context, io.Writer, io.Writer) error {
		return errors.New("nwcli was installed from a prebuilt release archive")
	}

	_, _, err := executeForTest("update")
	if got := ExitCode(err); got != exitUsage {
		t.Fatalf("ExitCode(err) = %d, want %d", got, exitUsage)
	}
}
