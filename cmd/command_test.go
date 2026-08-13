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

func TestRootVersionDoesNotRunMenu(t *testing.T) {
	root := NewRootCommand(func(context.Context) error {
		t.Fatal("version should not run the interactive menu")
		return nil
	})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--version"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := stdout.String(); !strings.HasPrefix(got, "nwcli ") || strings.Count(got, "\n") != 1 {
		t.Fatalf("expected one concise version line, got %q", got)
	}
}

func TestRootHelpDoesNotRunMenu(t *testing.T) {
	root := NewRootCommand(func(context.Context) error {
		t.Fatal("help should not run the interactive menu")
		return nil
	})
	var stdout bytes.Buffer
	root.SetOut(&stdout)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"--help"})

	if err := root.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if got := stdout.String(); !strings.Contains(got, "nwcli is a network toolbox for subnet planning, DNS diagnostics, and connectivity checks.") {
		t.Fatalf("help missing command description:\n%s", got)
	}
}

func TestHTTPCommandRejectsMissingHost(t *testing.T) {
	_, _, err := executeForTest("http")
	if err == nil {
		t.Fatal("Execute returned nil error, want host arg error")
	}
}

func TestInteractiveActionsUseTUIAppFields(t *testing.T) {
	actions := interactiveActions()
	if len(actions) < 8 {
		t.Fatalf("len(actions) = %d, want expanded toolbox actions", len(actions))
	}
	if actions[0].Title != "Subnet carve" {
		t.Fatalf("first action = %q, want Subnet carve", actions[0].Title)
	}
	if len(actions[0].Fields) != 2 || actions[0].Fields[0].Label != "Parent CIDR" || actions[0].Fields[1].Label != "Child prefixes" {
		t.Fatalf("subnet fields = %#v", actions[0].Fields)
	}
	for _, action := range actions {
		if action.Title == "Ping test" {
			if action.RunProgress == nil {
				t.Fatalf("ping action = %#v, want progress-enabled Ping test", action)
			}
			return
		}
	}
	t.Fatal("Ping test action not found")
}
