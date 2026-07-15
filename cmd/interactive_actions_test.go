package cmd

import (
	"testing"

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

func fieldValidator(action tui.Action, name string) func(string) error {
	for _, field := range action.Fields {
		if field.Name == name {
			return field.Validate
		}
	}
	return func(string) error { return nil }
}
