package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/mathwro/nwcli/internal/checks"
)

func TestSubnetCarveJSONOutputUsesSharedEnvelope(t *testing.T) {
	stdout, _, err := executeForTest("--output", "json", "subnet", "carve", "--cidr", "10.0.0.0/24", "--sizes", "27,27")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Version int    `json:"version"`
		Command string `json:"command"`
		OK      bool   `json:"ok"`
		Result  struct {
			Parent string `json:"parent"`
			Blocks []struct {
				CIDR string `json:"cidr"`
			} `json:"blocks"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("stdout is not JSON: %v\n%s", err, stdout)
	}
	if envelope.Version != 1 || envelope.Command != "subnet carve" || !envelope.OK {
		t.Fatalf("envelope = %#v", envelope)
	}
	if envelope.Result.Parent != "10.0.0.0/24" || len(envelope.Result.Blocks) != 2 || envelope.Result.Blocks[1].CIDR != "10.0.0.32/27" {
		t.Fatalf("result = %#v", envelope.Result)
	}
}

func TestRootRejectsUnknownOutputFormat(t *testing.T) {
	_, _, err := executeForTest("--output", "yaml", "subnet", "carve", "--cidr", "10.0.0.0/24", "--sizes", "27")
	if err == nil {
		t.Fatal("Execute returned nil error")
	}
}

func TestExitCodeUsesTypedFailuresAndUsageFallback(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Fatalf("ExitCode(nil) = %d", got)
	}
	if got := ExitCode(newCommandError(exitUnhealthy, errors.New("unhealthy"))); got != exitUnhealthy {
		t.Fatalf("ExitCode(typed) = %d", got)
	}
	if got := ExitCode(errors.New("bad arguments")); got != exitUsage {
		t.Fatalf("ExitCode(usage) = %d", got)
	}
}

func TestFailedCheckSuppressesCobraUsageAndStderrInJSONMode(t *testing.T) {
	original := httpCheck
	t.Cleanup(func() { httpCheck = original })
	httpCheck = func(_ context.Context, _ *http.Client, opts checks.HTTPOptions) checks.HTTPResult {
		return checks.HTTPResult{URL: "http://" + opts.Host + ":80/", StatusCode: 503, Reachable: true, Healthy: false}
	}

	var stdout, stderr bytes.Buffer
	code := runCLI(context.Background(), []string{"--output", "json", "http", "example.com"}, &stdout, &stderr, func(context.Context) error { return nil })
	if code != exitUnhealthy {
		t.Fatalf("code = %d, want %d", code, exitUnhealthy)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty when JSON error envelope emitted", stderr.String())
	}
	var envelope outputEnvelope
	if err := json.Unmarshal(stdout.Bytes(), &envelope); err != nil {
		t.Fatalf("invalid JSON %q: %v", stdout.String(), err)
	}
	if envelope.OK || envelope.Error == nil || envelope.Error.Code != exitUnhealthy {
		t.Fatalf("envelope = %#v", envelope)
	}
}

func TestFailedCheckPrintsErrorOnceWithoutUsageInTextMode(t *testing.T) {
	original := httpCheck
	t.Cleanup(func() { httpCheck = original })
	httpCheck = func(_ context.Context, _ *http.Client, opts checks.HTTPOptions) checks.HTTPResult {
		return checks.HTTPResult{URL: "http://" + opts.Host + ":80/", StatusCode: 503, Reachable: true, Healthy: false}
	}

	var stdout, stderr bytes.Buffer
	code := runCLI(context.Background(), []string{"http", "example.com"}, &stdout, &stderr, func(context.Context) error { return nil })
	if code != exitUnhealthy {
		t.Fatalf("code = %d, want %d", code, exitUnhealthy)
	}
	if got := stderr.String(); got != "endpoint reachable but unhealthy: status 503\n" {
		t.Fatalf("stderr = %q, want single error line without usage", got)
	}
}
