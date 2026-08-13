package version

import "testing"

func TestInjectedBuildMetadata(t *testing.T) {
	oldVersion, oldCommit, oldDate := semanticVersion, commit, buildDate
	t.Cleanup(func() {
		semanticVersion, commit, buildDate = oldVersion, oldCommit, oldDate
	})

	semanticVersion = "1.2.3"
	commit = "0123456789abcdef0123456789abcdef01234567"
	buildDate = "2026-08-13"

	if got, want := String(), "nwcli 1.2.3 (commit 0123456789ab, built 2026-08-13)"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestDevelopmentMetadataUsesStableFallbacks(t *testing.T) {
	oldVersion, oldCommit, oldDate := semanticVersion, commit, buildDate
	t.Cleanup(func() {
		semanticVersion, commit, buildDate = oldVersion, oldCommit, oldDate
	})

	semanticVersion = ""
	commit = ""
	buildDate = ""

	version, revision, date := buildMetadata()
	if version != "0.0.0-dev" {
		t.Fatalf("version = %q, want development version", version)
	}
	if revision == "" || date == "" {
		t.Fatalf("metadata = (%q, %q, %q), want non-empty fallbacks", version, revision, date)
	}
}
