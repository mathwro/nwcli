package cmd

import (
	"context"
	"testing"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/mathwro/nwcli/internal/tui"
)

func TestInteractiveActionsExposeNewDiagnosticsWithDescriptions(t *testing.T) {
	actions := interactiveActions()
	byTitle := make(map[string]tui.Action, len(actions))
	for _, action := range actions {
		byTitle[action.Title] = action
		if action.Title != "Quit" && action.Description == "" {
			t.Fatalf("action %q has no description", action.Title)
		}
	}
	for _, title := range []string{"Subnet carve", "CIDR inspect", "Ping test", "DNS lookup", "TCP connect", "TLS inspect", "HTTP/S test", "Quit"} {
		if _, ok := byTitle[title]; !ok {
			t.Fatalf("missing action %q", title)
		}
	}
	for _, title := range []string{"Ping test", "DNS lookup", "TCP connect", "TLS inspect", "HTTP/S test"} {
		if byTitle[title].RunProgress == nil {
			t.Fatalf("action %q does not use progress UI", title)
		}
	}
	foundContinuous := false
	for _, field := range byTitle["Ping test"].Fields {
		if field.Name == "continuous" && field.Default == "false" && field.Validate != nil {
			foundContinuous = true
		}
	}
	if !foundContinuous {
		t.Fatal("Ping test action missing validated continuous field")
	}
}

func TestInteractiveActionFieldsValidateBeforeRunning(t *testing.T) {
	actions := interactiveActions()
	for _, action := range actions {
		switch action.Title {
		case "DNS lookup":
			if err := fieldValidator(action, "type")("SRV"); err == nil {
				t.Fatal("DNS type validator accepted SRV")
			}
		case "TCP connect":
			if err := fieldValidator(action, "address")("example.com"); err == nil {
				t.Fatal("TCP address validator accepted missing port")
			}
		case "CIDR inspect":
			if err := fieldValidator(action, "cidr")("not-a-cidr"); err == nil {
				t.Fatal("CIDR validator accepted invalid input")
			}
		}
	}
}

func TestInteractiveSubnetBuildsStructuredAllocationResult(t *testing.T) {
	result, err := runInteractiveSubnet(context.Background(), map[string]string{
		"cidr":  "10.0.0.0/24",
		"sizes": "27,27,28,28",
	})
	if err != nil {
		t.Fatalf("runInteractiveSubnet returned error: %v", err)
	}
	if result.Headline != "10.0.0.0/24" || result.Summary != "96 of 256 addresses allocated" {
		t.Fatalf("result heading = %q, summary = %q", result.Headline, result.Summary)
	}
	if result.Allocation == nil {
		t.Fatal("Allocation = nil")
	}
	if len(result.Allocation.Segments) != 4 {
		t.Fatalf("len(Segments) = %d, want 4", len(result.Allocation.Segments))
	}
	if result.Allocation.Available == nil || result.Allocation.Available.Size != 160 {
		t.Fatalf("Available = %#v, want 160 addresses", result.Allocation.Available)
	}
}

func TestInteractiveContinuousPingReportsLiveLoss(t *testing.T) {
	original := continuousPingCheck
	t.Cleanup(func() { continuousPingCheck = original })
	continuousPingCheck = func(_ context.Context, opts checks.PingOptions) (checks.PingSummary, error) {
		summary := checks.PingSummary{Sent: 4, Received: 3, LossPercent: 25}
		opts.OnUpdate(checks.PingUpdate{
			Attempt: checks.PingAttempt{Seq: 4, Err: context.DeadlineExceeded},
			Summary: summary,
		})
		return summary, nil
	}
	var events []tui.ProgressEvent
	_, err := runInteractivePing(context.Background(), map[string]string{
		"host":       "127.0.0.1",
		"count":      "4",
		"timeout":    "1s",
		"continuous": "true",
	}, func(event tui.ProgressEvent) {
		events = append(events, event)
	})
	if err != nil {
		t.Fatalf("runInteractivePing returned error: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("len(events) = %d, want 1", len(events))
	}
	event := events[0]
	if event.Total != 0 || event.Current != 4 || event.Tone != tui.ToneDanger {
		t.Fatalf("event = %#v", event)
	}
	if event.Label == "" || len(event.Fields) != 4 || event.Fields[2].Label != "Dropped" || event.Fields[2].Value != "1" || event.Fields[3].Value != "25.0%" {
		t.Fatalf("event fields = %#v", event)
	}
}

func fieldValidator(action tui.Action, name string) func(string) error {
	for _, field := range action.Fields {
		if field.Name == name {
			return field.Validate
		}
	}
	return func(string) error { return nil }
}
