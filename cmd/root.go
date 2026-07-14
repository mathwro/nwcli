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
	items := []tui.MenuItem{
		{Title: "Subnet carve", Run: runInteractiveSubnet},
		{Title: "Ping test", Run: runInteractivePing},
		{Title: "HTTP/S test", Run: runInteractiveHTTP},
		{Title: "Quit"},
	}
	return tui.RunMenu(ctx, items)
}

func runInteractiveSubnet(ctx context.Context) (string, error) {
	values, err := tui.Ask(ctx, "Subnet carve", []tui.Field{
		{Name: "cidr", Prompt: "Parent CIDR", Default: "10.0.0.0/24"},
		{Name: "sizes", Prompt: "Child prefixes", Default: "27,27,28,28"},
	})
	if err != nil {
		return "", err
	}
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

func runInteractivePing(ctx context.Context) (string, error) {
	values, err := tui.Ask(ctx, "Ping test", []tui.Field{
		{Name: "host", Prompt: "Host", Default: "127.0.0.1"},
		{Name: "count", Prompt: "Count", Default: "4"},
		{Name: "timeout", Prompt: "Timeout", Default: "1s"},
	})
	if err != nil {
		return "", err
	}
	count, err := strconv.Atoi(values["count"])
	if err != nil {
		return "", fmt.Errorf("invalid count %q: %w", values["count"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return "", fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	attempts, summary, err := checks.CheckPing(ctx, checks.PingOptions{Host: values["host"], Count: count, Timeout: timeout})
	var out bytes.Buffer
	renderPing(&out, attempts, summary)
	return out.String(), err
}

func runInteractiveHTTP(ctx context.Context) (string, error) {
	values, err := tui.Ask(ctx, "HTTP/S test", []tui.Field{
		{Name: "host", Prompt: "Host", Default: "example.com"},
		{Name: "port", Prompt: "Port", Default: "443"},
		{Name: "tls", Prompt: "TLS true/false", Default: "true"},
		{Name: "path", Prompt: "Path", Default: "/"},
		{Name: "timeout", Prompt: "Timeout", Default: "2s"},
	})
	if err != nil {
		return "", err
	}
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
