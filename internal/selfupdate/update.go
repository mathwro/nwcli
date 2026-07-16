package selfupdate

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

const Target = "github.com/mathwro/nwcli@latest"

type runner func(context.Context, string, []string, io.Writer, io.Writer) error

func InstallLatest(ctx context.Context, stdout, stderr io.Writer) error {
	return installLatest(ctx, stdout, stderr, runCommand)
}

func installLatest(ctx context.Context, stdout, stderr io.Writer, run runner) error {
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
