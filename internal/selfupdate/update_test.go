package selfupdate

import (
	"bytes"
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestInstallLatestRunsGoInstallForLatestModule(t *testing.T) {
	var gotName string
	var gotArgs []string
	var gotStdout io.Writer
	var gotStderr io.Writer
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	err := installLatest(context.Background(), stdout, stderr, func(_ context.Context, name string, args []string, commandStdout, commandStderr io.Writer) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		gotStdout = commandStdout
		gotStderr = commandStderr
		return nil
	})
	if err != nil {
		t.Fatalf("installLatest returned error: %v", err)
	}
	if gotName != "go" {
		t.Fatalf("command = %q, want go", gotName)
	}
	wantArgs := []string{"install", "github.com/mathwro/nwcli@latest"}
	if !reflect.DeepEqual(gotArgs, wantArgs) {
		t.Fatalf("args = %#v, want %#v", gotArgs, wantArgs)
	}
	if gotStdout != stdout || gotStderr != stderr {
		t.Fatal("installLatest did not forward output writers")
	}
}

func TestInstallLatestWrapsGoInstallFailure(t *testing.T) {
	sentinel := errors.New("install failed")
	err := installLatest(context.Background(), io.Discard, io.Discard, func(context.Context, string, []string, io.Writer, io.Writer) error {
		return sentinel
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want wrapped install failure", err)
	}
	if !strings.Contains(err.Error(), "update nwcli") {
		t.Fatalf("error = %q, want update context", err)
	}
}

func TestInstallLatestRejectsHomebrewAndScoopManagedBinaries(t *testing.T) {
	for _, tc := range []struct {
		exe  string
		hint string
	}{
		{exe: "/opt/homebrew/Cellar/nwcli/1.0.0/bin/nwcli", hint: "brew upgrade"},
		{exe: `C:\Users\alice\scoop\apps\nwcli\current\nwcli.exe`, hint: "scoop update"},
	} {
		err := installLatestWithEnv(context.Background(), io.Discard, io.Discard,
			func() (string, error) { return tc.exe, nil },
			func(string) (string, error) { return "/usr/bin/go", nil },
			func(context.Context, string, []string, io.Writer, io.Writer) error {
				t.Fatalf("go install should not run for %s", tc.exe)
				return nil
			},
		)
		if err == nil || !strings.Contains(err.Error(), tc.hint) {
			t.Fatalf("exe %s: err = %v, want hint %q", tc.exe, err, tc.hint)
		}
	}
}

func TestInstallLatestExplainsMissingGoToolchain(t *testing.T) {
	err := installLatestWithEnv(context.Background(), io.Discard, io.Discard,
		func() (string, error) { return "/usr/local/bin/nwcli", nil },
		func(string) (string, error) { return "", errors.New("executable file not found") },
		func(context.Context, string, []string, io.Writer, io.Writer) error { return nil },
	)
	if err == nil || !strings.Contains(err.Error(), "Go toolchain not found") {
		t.Fatalf("err = %v, want missing Go toolchain explanation", err)
	}
}

func TestInstallLatestRejectsPrebuiltReleaseBinaryOutsidePackageManagerPaths(t *testing.T) {
	oldIsPrebuilt := isPrebuiltRelease
	t.Cleanup(func() { isPrebuiltRelease = oldIsPrebuilt })
	isPrebuiltRelease = func() bool { return true }

	err := installLatestWithEnv(context.Background(), io.Discard, io.Discard,
		func() (string, error) { return "/usr/local/bin/nwcli", nil },
		func(string) (string, error) { return "/usr/bin/go", nil },
		func(context.Context, string, []string, io.Writer, io.Writer) error {
			t.Fatal("go install should not run for prebuilt release binary")
			return nil
		},
	)
	if err == nil || !strings.Contains(err.Error(), "prebuilt release") || !strings.Contains(err.Error(), "https://github.com/mathwro/nwcli/releases/latest") {
		t.Fatalf("err = %v, want prebuilt release guidance", err)
	}
}
