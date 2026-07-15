package cmd

import (
	"encoding/json"
	"errors"
	"testing"
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
