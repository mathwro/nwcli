package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const Target = "github.com/mathwro/nwcli@latest"

type runner func(context.Context, string, []string, io.Writer, io.Writer) error

func InstallLatest(ctx context.Context, stdout, stderr io.Writer) error {
	return installLatestWithEnv(ctx, stdout, stderr, os.Executable, exec.LookPath, runCommand)
}

func installLatest(ctx context.Context, stdout, stderr io.Writer, run runner) error {
	return installLatestWithEnv(ctx, stdout, stderr, func() (string, error) { return "", nil }, func(string) (string, error) { return "go", nil }, run)
}

func installLatestWithEnv(
	ctx context.Context,
	stdout, stderr io.Writer,
	executable func() (string, error),
	lookPath func(string) (string, error),
	run runner,
) error {
	if exe, err := executable(); err == nil && exe != "" {
		if resolved, evalErr := filepath.EvalSymlinks(exe); evalErr == nil && resolved != "" {
			exe = resolved
		}
		normalized := strings.ToLower(strings.ReplaceAll(exe, `\`, `/`))
		switch {
		case strings.Contains(normalized, "/cellar/nwcli/") || strings.Contains(normalized, "/linuxbrew/"):
			return fmt.Errorf("nwcli is managed by Homebrew (%s): run 'brew upgrade mathwro/tools/nwcli' instead", exe)
		case strings.Contains(normalized, "/scoop/apps/nwcli/"):
			return fmt.Errorf("nwcli is managed by Scoop (%s): run 'scoop update nwcli' instead", exe)
		}
	}
	if _, err := lookPath("go"); err != nil {
		return fmt.Errorf("Go toolchain not found on PATH: install Go or update nwcli with your package manager (brew/scoop)")
	}
	if err := run(ctx, "go", []string{"install", Target}, stdout, stderr); err != nil {
		return fmt.Errorf("update nwcli: %w", err)
	}
	return nil
}

func runCommand(ctx context.Context, name string, args []string, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
