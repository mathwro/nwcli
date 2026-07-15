package cmd

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSubnetInspectCommand(t *testing.T) {
	stdout, _, err := executeForTest("subnet", "inspect", "10.0.0.37/24")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, want := range []string{"CIDR: 10.0.0.0/24", "Family: IPv4", "First: 10.0.0.0", "Last: 10.0.0.255", "Addresses: 256"} {
		if !strings.Contains(stdout, want+"\n") {
			t.Fatalf("stdout missing %q:\n%s", want, stdout)
		}
	}
}

func TestSubnetContainsAndOverlapCommands(t *testing.T) {
	stdout, _, err := executeForTest("subnet", "contains", "10.0.0.0/24", "10.0.0.42")
	if err != nil || stdout != "true\n" {
		t.Fatalf("contains = %q, %v", stdout, err)
	}
	stdout, _, err = executeForTest("subnet", "overlap", "10.0.0.0/24", "10.0.1.0/24")
	if err != nil || stdout != "false\n" {
		t.Fatalf("overlap = %q, %v", stdout, err)
	}
}

func TestSubnetSplitAndSummarizeCommands(t *testing.T) {
	stdout, _, err := executeForTest("subnet", "split", "10.0.0.0/24", "--prefix", "26")
	if err != nil {
		t.Fatalf("split returned error: %v", err)
	}
	for _, prefix := range []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"} {
		if !strings.Contains(stdout, prefix+"\n") {
			t.Fatalf("split output missing %q:\n%s", prefix, stdout)
		}
	}
	stdout, _, err = executeForTest("subnet", "summarize", "10.0.0.0/25", "10.0.0.128/25")
	if err != nil || stdout != "10.0.0.0/24\n" {
		t.Fatalf("summarize = %q, %v", stdout, err)
	}
}

func TestSubnetInspectJSONOutput(t *testing.T) {
	stdout, _, err := executeForTest("--output", "json", "subnet", "inspect", "2001:db8::1/126")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	var envelope struct {
		Command string `json:"command"`
		Result  struct {
			CIDR string `json:"cidr"`
		} `json:"result"`
	}
	if err := json.Unmarshal([]byte(stdout), &envelope); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if envelope.Command != "subnet inspect" || envelope.Result.CIDR != "2001:db8::/126" {
		t.Fatalf("envelope = %#v", envelope)
	}
}
