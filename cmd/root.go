package cmd

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/mathwro/nwcli/internal/subnet"
	"github.com/mathwro/nwcli/internal/tui"
	"github.com/spf13/cobra"
)

func NewRootCommand(menuRunner func(context.Context) error) *cobra.Command {
	if menuRunner == nil {
		menuRunner = runInteractiveMenu
	}
	root := &cobra.Command{
		Use:   "nwcli",
		Short: "Network toolbox for subnet planning and connectivity checks",
		Long:  "nwcli is a network toolbox for carving IPv4 subnets, running ICMP ping checks, and testing HTTP/S endpoints.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return menuRunner(cmd.Context())
		},
	}
	root.AddCommand(newSubnetCommand(), newPingCommand(), newHTTPCommand())
	return root
}

func Execute() {
	if err := NewRootCommand(runInteractiveMenu).Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runInteractiveMenu(ctx context.Context) error {
	return tui.RunApp(ctx, interactiveActions())
}

func interactiveActions() []tui.Action {
	return []tui.Action{
		{
			Title: "Subnet carve",
			Fields: []tui.Field{
				{Name: "cidr", Label: "Parent CIDR", Default: "10.0.0.0/24"},
				{Name: "sizes", Label: "Child prefixes", Default: "27,27,28,28"},
			},
			Run: runInteractiveSubnet,
		},
		{
			Title: "Ping test",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Default: "127.0.0.1"},
				{Name: "count", Label: "Count", Default: "4"},
				{Name: "timeout", Label: "Timeout", Default: "1s"},
			},
			RunProgress: runInteractivePing,
		},
		{
			Title: "HTTP/S test",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Default: "example.com"},
				{Name: "port", Label: "Port", Default: "443"},
				{Name: "tls", Label: "TLS true/false", Default: "true"},
				{Name: "path", Label: "Path", Default: "/"},
				{Name: "timeout", Label: "Timeout", Default: "2s"},
			},
			Run: runInteractiveHTTP,
		},
		{Title: "Quit"},
	}
}

func runInteractiveSubnet(_ context.Context, values map[string]string) (string, error) {
	prefixes, err := subnet.ParsePrefixes(values["sizes"])
	if err != nil {
		return "", err
	}
	plan, err := subnet.Plan(values["cidr"], prefixes)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := subnet.Render(&out, plan); err != nil {
		return "", err
	}
	return out.String(), nil
}

func runInteractivePing(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (string, error) {
	count, err := strconv.Atoi(values["count"])
	if err != nil {
		return "", fmt.Errorf("invalid count %q: %w", values["count"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	options := checks.PingOptions{
		Host:    values["host"],
		Count:   count,
		Timeout: timeout,
		OnProgress: func(current, total int) {
			if report != nil {
				report(tui.ProgressEvent{
					Label:   fmt.Sprintf("Ping %d of %d", current, total),
					Current: current,
					Total:   total,
				})
			}
		},
	}
	attempts, summary, err := checks.CheckPing(ctx, options)
	var out bytes.Buffer
	renderPing(&out, attempts, summary)
	return out.String(), err
}

func runInteractiveHTTP(ctx context.Context, values map[string]string) (string, error) {
	port, err := strconv.Atoi(values["port"])
	if err != nil {
		return "", fmt.Errorf("invalid port %q: %w", values["port"], err)
	}
	useTLS, err := strconv.ParseBool(values["tls"])
	if err != nil {
		return "", fmt.Errorf("invalid tls value %q: %w", values["tls"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	result := checks.CheckHTTP(ctx, nil, checks.HTTPOptions{Host: values["host"], Port: port, TLS: useTLS, Path: values["path"], Timeout: timeout})
	var out bytes.Buffer
	renderHTTP(&out, result)
	return out.String(), httpResultError(result)
}
