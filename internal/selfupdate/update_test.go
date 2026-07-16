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
