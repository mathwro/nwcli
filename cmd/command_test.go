package cmd

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func executeForTest(args ...string) (string, string, error) {
	root := NewRootCommand(func(context.Context) error { return nil })
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&stderr)
	root.SetArgs(args)
	err := root.Execute()
	return stdout.String(), stderr.String(), err
}

func TestSubnetCarveCommandPrintsAllocation(t *testing.T) {
	stdout, _, err := executeForTest("subnet", "carve", "--cidr", "10.0.0.0/24", "--sizes", "27,27,28,28")
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	for _, fragment := range []string{"10.0.0.0/27", "10.0.0.32/27", "10.0.0.64/28", "10.0.0.80/28", "Unused: 10.0.0.96 - 10.0.0.255"} {
		if !strings.Contains(stdout, fragment) {
			t.Fatalf("stdout missing %q:\n%s", fragment, stdout)
		}
	}
}

func TestSubnetCarveRequiresCIDR(t *testing.T) {
	_, _, err := executeForTest("subnet", "carve", "--sizes", "27")
	if err == nil {
		t.Fatal("Execute returned nil error, want required cidr error")
	}
	if !strings.Contains(err.Error(), "required flag") || !strings.Contains(err.Error(), "cidr") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRootNoArgsRunsInjectedMenu(t *testing.T) {
	run := false
	root := NewRootCommand(func(context.Context) error {
		run = true
		return nil
	})
	root.SetArgs([]string{})
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if !run {
		t.Fatal("menu runner was not called")
	}
}

func TestHTTPCommandRejectsMissingHost(t *testing.T) {
	_, _, err := executeForTest("http")
	if err == nil {
		t.Fatal("Execute returned nil error, want host arg error")
	}
}
