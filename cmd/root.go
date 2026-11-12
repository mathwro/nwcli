package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/mathwro/nwcli/internal/subnet"
	"github.com/mathwro/nwcli/internal/tui"
	"github.com/mathwro/nwcli/internal/version"
	"github.com/spf13/cobra"
)

func NewRootCommand(menuRunner func(context.Context) error) *cobra.Command {
	if menuRunner == nil {
		menuRunner = runInteractiveMenu
	}
	var outputFormat string
	root := &cobra.Command{
		Use:           "nwcli",
		Short:         "Network toolbox for subnet planning and connectivity checks",
		Long:          "nwcli is a network toolbox for subnet planning, DNS diagnostics, and connectivity checks.",
		Version:       version.String(),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return validateOutputFormat(outputFormat)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return menuRunner(cmd.Context())
		},
	}
	root.SetVersionTemplate("{{.Version}}\n")
	root.PersistentFlags().StringVar(&outputFormat, "output", "text", "output format: text or json")
	_ = root.RegisterFlagCompletionFunc("output", func(*cobra.Command, []string, string) ([]string, cobra.ShellCompDirective) {
		return []string{"text", "json"}, cobra.ShellCompDirectiveNoFileComp
	})
	root.AddCommand(
		newSubnetCommand(),
		newPingCommand(),
		newTraceCommand(),
		newHTTPCommand(),
		newDNSCommand(),
		newTCPCommand(),
		newTLSCommand(),
		newInterfacesCommand(),
		newBatchCommand(),
		newUpdateCommand(),
		newCompletionCommand(),
	)
	return root
}

func Execute() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if code := runCLI(ctx, os.Args[1:], os.Stdout, os.Stderr, runInteractiveMenu); code != 0 {
		os.Exit(code)
	}
}

func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer, menuRunner func(context.Context) error) int {
	root := NewRootCommand(menuRunner)
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)
	err := root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	if isEmittedJSONError(err) {
		return ExitCode(err)
	}
	format, flagErr := root.PersistentFlags().GetString("output")
	if flagErr == nil && format == "json" {
		if writeErr := writeJSONOutput(stdout, root.Name(), nil, err); writeErr == nil {
			return ExitCode(err)
		}
	}
	fmt.Fprintln(stderr, err)
	return ExitCode(err)
}

func runInteractiveMenu(ctx context.Context) error {
	return tui.RunApp(ctx, interactiveActions())
}

func interactiveActions() []tui.Action {
	return []tui.Action{
		{
			Title:       "Subnet carve",
			Description: "Allocate ordered IPv4 child networks and see remaining space",
			Fields: []tui.Field{
				{Name: "cidr", Label: "Parent CIDR", Placeholder: "10.0.0.0/24", Validate: validateCIDR},
				{Name: "sizes", Label: "Child prefixes", Placeholder: "web:27,app:27,db:28,mgmt:28", Validate: validatePrefixes},
				{Name: "sort", Label: "Sort largest-first (VLSM)", Default: "true", Toggle: true, Validate: validateBoolean},
			},
			Run: runInteractiveSubnet,
		},
		{
			Title:       "CIDR inspect",
			Description: "Inspect, test containment/overlap, split, or summarize CIDRs",
			Fields: []tui.Field{
				{Name: "mode", Label: "Operation", Default: "inspect", Options: []string{"inspect", "contains", "overlap", "split", "summarize"}},
				{
					Name: "cidr", Label: "CIDR", Placeholder: "10.0.0.0/24",
					Visible:  func(v map[string]string) bool { return v["mode"] != "summarize" },
					Validate: validateCIDR,
				},
				{
					Name: "address", Label: "IP address", Placeholder: "10.0.0.42",
					Visible:  func(v map[string]string) bool { return v["mode"] == "contains" },
					Validate: validateIPAddress,
				},
				{
					Name: "cidr_b", Label: "Second CIDR", Placeholder: "10.0.0.128/25",
					Visible:  func(v map[string]string) bool { return v["mode"] == "overlap" },
					Validate: validateCIDR,
				},
				{
					Name: "prefix", Label: "Child prefix", Placeholder: "26",
					Visible:  func(v map[string]string) bool { return v["mode"] == "split" },
					Validate: validatePositiveInteger,
				},
				{
					Name: "cidrs", Label: "CIDR list", Placeholder: "10.0.0.0/25, 10.0.0.128/25",
					Visible:  func(v map[string]string) bool { return v["mode"] == "summarize" },
					Validate: validateCIDRList,
				},
			},
			Run: runInteractiveCIDR,
		},
		{
			Title:       "Ping test",
			Description: "Measure ICMP reachability, loss, and latency",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Placeholder: "127.0.0.1", Validate: validateRequired},
				{Name: "continuous", Label: "Continuous", Default: "false", Toggle: true, Validate: validateBoolean},
				{
					Name: "count", Label: "Count", Placeholder: "4",
					Visible:  func(values map[string]string) bool { return values["continuous"] != "true" },
					Validate: validatePositiveInteger,
				},
				{Name: "timeout", Label: "Timeout", Placeholder: "1s", Validate: validateDuration},
			},
			RunProgress: runInteractivePing,
		},
		{
			Title:       "Trace route",
			Description: "Trace hop-by-hop ICMP path and latency to a host",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Placeholder: "1.1.1.1", Validate: validateRequired},
				{Name: "max_hops", Label: "Max hops", Placeholder: "20", Validate: validateHopLimit},
				{Name: "timeout", Label: "Hop timeout", Placeholder: "1s", Validate: validateDuration},
			},
			RunProgress: runInteractiveTrace,
		},
		{
			Title:       "DNS lookup",
			Description: "Query A, AAAA, CNAME, MX, NS, TXT, or PTR records",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Placeholder: "example.com", Validate: validateRequired},
				{Name: "type", Label: "Record type", Placeholder: "A", Validate: validateDNSRecordType},
				{Name: "server", Label: "Resolver (optional)", Placeholder: "1.1.1.1"},
				{Name: "timeout", Label: "Timeout", Placeholder: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveDNS,
		},
		{
			Title:       "TCP connect",
			Description: "Measure transport reachability to a port, list, or range",
			Fields: []tui.Field{
				{Name: "address", Label: "Address / ports", Placeholder: "example.com:80,443", Validate: validateTCPTarget},
				{Name: "timeout", Label: "Timeout", Placeholder: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveTCP,
		},
		{
			Title:       "TLS inspect",
			Description: "Verify TLS and inspect certificate validity and negotiation",
			Fields: []tui.Field{
				{Name: "address", Label: "Address", Placeholder: "example.com:443", Validate: validateHostPort},
				{Name: "server_name", Label: "Server name (optional)", Placeholder: "example.com"},
				{Name: "timeout", Label: "Timeout", Placeholder: "3s", Validate: validateDuration},
				{Name: "warn_before", Label: "Expiry warning", Placeholder: "30d", Validate: validateExtendedDuration},
			},
			RunProgress: runInteractiveTLS,
		},
		{
			Title:       "HTTP/S test",
			Description: "Check endpoint status, reachability, health, and latency",
			Fields: []tui.Field{
				{Name: "host", Label: "Host", Placeholder: "example.com", Validate: validateRequired},
				{Name: "port", Label: "Port", Placeholder: "443", Validate: validatePort},
				{Name: "tls", Label: "TLS", Default: "true", Toggle: true, Validate: validateBoolean},
				{Name: "path", Label: "Path", Placeholder: "/"},
				{Name: "follow_redirects", Label: "Follow redirects", Default: "false", Toggle: true, Validate: validateBoolean},
				{Name: "timeout", Label: "Timeout", Placeholder: "2s", Validate: validateDuration},
			},
			RunProgress: runInteractiveHTTP,
		},
		{
			Title:       "Local interfaces",
			Description: "Inspect local NICs, MTU, MAC, flags, and assigned CIDRs",
			Fields: []tui.Field{
				{Name: "all", Label: "Include down", Default: "false", Toggle: true, Validate: validateBoolean},
			},
			Run: runInteractiveInterfaces,
		},
		{Title: "Quit", Description: "Exit nwcli"},
	}
}

func runInteractiveSubnet(_ context.Context, values map[string]string) (tui.Result, error) {
	specs, err := subnet.ParsePrefixSpecs(values["sizes"])
	if err != nil {
		return tui.Result{}, err
	}
	sortVLSM := false
	if values["sort"] != "" {
		sortVLSM, err = strconv.ParseBool(values["sort"])
		if err != nil {
			return tui.Result{}, fmt.Errorf("invalid sort value %q: %w", values["sort"], err)
		}
	}
	plan, err := subnet.PlanSpecs(values["cidr"], specs, sortVLSM)
	if err != nil {
		return tui.Result{}, err
	}
	var out bytes.Buffer
	if err := subnet.Render(&out, plan); err != nil {
		return tui.Result{}, err
	}
	segments := make([]tui.AllocationSegment, len(plan.Blocks))
	var allocated uint64
	for i, block := range plan.Blocks {
		label := block.CIDR.String()
		if block.Name != "" {
			label = block.Name + "  " + label
		}
		segments[i] = tui.AllocationSegment{
			Index:  block.Index,
			Label:  label,
			Start:  block.First.String(),
			End:    block.Last.String(),
			Offset: uint64(block.StartOffset),
			Size:   uint64(block.Size),
		}
		allocated += uint64(block.Size)
	}
	var available *tui.AllocationSegment
	if plan.Unused != nil {
		available = &tui.AllocationSegment{
			Start:  plan.Unused.First.String(),
			End:    plan.Unused.Last.String(),
			Offset: uint64(plan.Unused.StartOffset),
			Size:   uint64(plan.Unused.Size),
		}
	}
	return tui.Result{
		Headline: plan.Parent.String(),
		Summary:  fmt.Sprintf("%d of %d addresses allocated", allocated, plan.ParentSize),
		Text:     out.String(),
		Allocation: &tui.AllocationMap{
			Start:     plan.ParentFirst.String(),
			End:       plan.ParentLast.String(),
			Total:     uint64(plan.ParentSize),
			Segments:  segments,
			Available: available,
		},
	}, nil
}

func runInteractiveCIDR(_ context.Context, values map[string]string) (tui.Result, error) {
	mode := values["mode"]
	if mode == "" {
		mode = "inspect"
	}
	switch mode {
	case "contains":
		contains, err := subnet.ContainsCIDR(values["cidr"], values["address"])
		if err != nil {
			return tui.Result{}, err
		}
		verdict := "Address is outside CIDR"
		if contains {
			verdict = "Address is inside CIDR"
		}
		return tui.Result{
			Headline: fmt.Sprintf("%s  ∈  %s", values["address"], values["cidr"]),
			Summary:  verdict,
			Text:     fmt.Sprintf("%t\n", contains),
			Fields: []tui.ResultField{
				{Label: "CIDR", Value: values["cidr"]},
				{Label: "Address", Value: values["address"]},
				boolResultField("Contains", contains),
			},
		}, nil
	case "overlap":
		overlap, err := subnet.OverlapsCIDR(values["cidr"], values["cidr_b"])
		if err != nil {
			return tui.Result{}, err
		}
		summary := "Ranges are disjoint"
		if overlap {
			summary = "Ranges overlap"
		}
		return tui.Result{
			Headline: fmt.Sprintf("%s  ∩  %s", values["cidr"], values["cidr_b"]),
			Summary:  summary,
			Text:     fmt.Sprintf("%t\n", overlap),
			Fields: []tui.ResultField{
				{Label: "First CIDR", Value: values["cidr"]},
				{Label: "Second CIDR", Value: values["cidr_b"]},
				{Label: "Overlaps", Value: strconv.FormatBool(overlap), Tone: map[bool]tui.ResultTone{true: tui.ToneWarning, false: tui.ToneSuccess}[overlap]},
			},
		}, nil
	case "split":
		childPrefix, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(values["prefix"]), "/"))
		if err != nil {
			return tui.Result{}, fmt.Errorf("invalid prefix %q: %w", values["prefix"], err)
		}
		children, err := subnet.SplitCIDR(values["cidr"], childPrefix)
		if err != nil {
			return tui.Result{}, err
		}
		return tui.Result{
			Headline: values["cidr"],
			Summary:  fmt.Sprintf("Split into %d /%d prefixes", len(children), childPrefix),
			Text:     strings.Join(children, "\n") + "\n",
			Fields: []tui.ResultField{
				{Label: "Parent", Value: values["cidr"]},
				{Label: "Child prefix", Value: fmt.Sprintf("/%d", childPrefix)},
				{Label: "Count", Value: strconv.Itoa(len(children)), Tone: tui.ToneSuccess},
			},
			Sections: []tui.ResultSection{{Title: "Child prefixes", Lines: children}},
		}, nil
	case "summarize":
		inputs := splitCIDRList(values["cidrs"])
		prefixes, err := subnet.SummarizeCIDRs(inputs)
		if err != nil {
			return tui.Result{}, err
		}
		return tui.Result{
			Headline: fmt.Sprintf("%d → %d prefixes", len(inputs), len(prefixes)),
			Summary:  "Collapsed adjacent and contained ranges",
			Text:     strings.Join(prefixes, "\n") + "\n",
			Fields: []tui.ResultField{
				{Label: "Input count", Value: strconv.Itoa(len(inputs))},
				{Label: "Summary count", Value: strconv.Itoa(len(prefixes)), Tone: tui.ToneSuccess},
			},
			Sections: []tui.ResultSection{{Title: "Summary prefixes", Lines: prefixes}},
		}, nil
	default:
		info, err := subnet.InspectCIDR(values["cidr"])
		if err != nil {
			return tui.Result{}, err
		}
		var out bytes.Buffer
		_ = renderCIDRInfo(&out, info)
		fields := []tui.ResultField{
			{Label: "Family", Value: info.Family},
			{Label: "Prefix", Value: fmt.Sprintf("/%d", info.Prefix)},
			{Label: "Scope", Value: info.Scope},
		}
		if info.Netmask != "" {
			fields = append(fields,
				tui.ResultField{Label: "Netmask", Value: info.Netmask},
				tui.ResultField{Label: "Wildcard", Value: info.Wildcard},
			)
		}
		fields = append(fields,
			tui.ResultField{Label: "First address", Value: info.First},
			tui.ResultField{Label: "Last address", Value: info.Last},
			tui.ResultField{Label: "Usable range", Value: fmt.Sprintf("%s → %s (%s)", info.UsableFirst, info.UsableLast, info.UsableAddresses)},
			tui.ResultField{Label: "Addresses", Value: info.Addresses, Tone: tui.ToneSuccess},
		)
		return tui.Result{
			Headline: info.CIDR,
			Summary:  fmt.Sprintf("%s · %s", info.Family, info.Scope),
			Text:     out.String(),
			Fields:   fields,
		}, nil
	}
}

func runInteractiveTrace(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	maxHops, err := strconv.Atoi(values["max_hops"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid max_hops %q: %w", values["max_hops"], err)
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	result, checkErr := traceCheck(ctx, checks.TraceOptions{
		Host:    values["host"],
		MaxHops: maxHops,
		Timeout: timeout,
		OnUpdate: func(hop checks.TraceHop) {
			if report == nil {
				return
			}
			label := fmt.Sprintf("Hop %d · %s (%s)", hop.TTL, hop.Address, hop.Latency.Round(time.Millisecond))
			tone := tui.ToneSuccess
			if hop.Timeout {
				label = fmt.Sprintf("Hop %d · * (timed out)", hop.TTL)
				tone = tui.ToneWarning
			} else if hop.Err != nil {
				label = fmt.Sprintf("Hop %d · %v", hop.TTL, hop.Err)
				tone = tui.ToneDanger
			}
			report(tui.ProgressEvent{Label: label, Current: hop.TTL, Total: maxHops, Tone: tone})
		},
	})
	var out bytes.Buffer
	renderTrace(&out, result)
	hopLines := make([]string, len(result.Hops))
	for i, hop := range result.Hops {
		if hop.Timeout {
			hopLines[i] = fmt.Sprintf("%2d  *", hop.TTL)
		} else if hop.Err != nil {
			hopLines[i] = fmt.Sprintf("%2d  %v", hop.TTL, hop.Err)
		} else {
			hopLines[i] = fmt.Sprintf("%2d  %-36s %s", hop.TTL, hop.Address, hop.Latency.Round(time.Microsecond))
		}
	}
	summary := fmt.Sprintf("Stopped after %d hops", len(result.Hops))
	if result.Reached {
		summary = fmt.Sprintf("Destination reached in %d hops", len(result.Hops))
	}
	headline := values["host"]
	if result.Target != "" && result.Target != values["host"] {
		headline = fmt.Sprintf("%s (%s)", values["host"], result.Target)
	}
	return tui.Result{
		Headline: headline,
		Summary:  summary,
		Text:     out.String(),
		Fields: []tui.ResultField{
			boolResultField("Reached", result.Reached),
			{Label: "Hops", Value: strconv.Itoa(len(result.Hops))},
			{Label: "Target IP", Value: result.Target},
		},
		Sections: []tui.ResultSection{{Title: "Route hops", Lines: hopLines}},
	}, pingResultError(checkErr)
}

func runInteractiveInterfaces(_ context.Context, values map[string]string) (tui.Result, error) {
	includeAll := values["all"] == "true"
	items, err := listInterfaces(includeAll)
	if err != nil {
		return tui.Result{}, newCommandError(exitUnreachable, err)
	}
	var out bytes.Buffer
	_ = renderInterfaces(&out, items)
	sections := make([]tui.ResultSection, 0, len(items))
	totalAddrs := 0
	for _, item := range items {
		totalAddrs += len(item.Addresses)
		mac := item.HardwareAddr
		if mac == "" {
			mac = "none"
		}
		lines := []string{fmt.Sprintf("mtu %d · mac %s · %s", item.MTU, mac, strings.Join(item.Flags, ","))}
		if len(item.Addresses) == 0 {
			lines = append(lines, "(no addresses)")
		} else {
			lines = append(lines, item.Addresses...)
		}
		sections = append(sections, tui.ResultSection{Title: item.Name, Lines: lines})
	}
	return tui.Result{
		Headline: fmt.Sprintf("%d interfaces", len(items)),
		Summary:  fmt.Sprintf("%d assigned addresses", totalAddrs),
		Text:     out.String(),
		Fields: []tui.ResultField{
			{Label: "Interfaces", Value: strconv.Itoa(len(items)), Tone: tui.ToneSuccess},
			{Label: "Addresses", Value: strconv.Itoa(totalAddrs)},
		},
		Sections: sections,
	}, nil
}

func runInteractivePing(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	continuous := false
	var err error
	if values["continuous"] != "" {
		continuous, err = strconv.ParseBool(values["continuous"])
		if err != nil {
			return tui.Result{}, fmt.Errorf("invalid continuous value %q: %w", values["continuous"], err)
		}
	}
	count := 0
	if !continuous {
		count, err = strconv.Atoi(values["count"])
		if err != nil {
			return tui.Result{}, fmt.Errorf("invalid count %q: %w", values["count"], err)
		}
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	var recorded []checks.PingAttempt
	const maxRecordedAttempts = 200
	options := checks.PingOptions{
		Host:    values["host"],
		Count:   count,
		Timeout: timeout,
		OnUpdate: func(update checks.PingUpdate) {
			if continuous {
				if len(recorded) >= maxRecordedAttempts {
					copy(recorded, recorded[1:])
					recorded[len(recorded)-1] = update.Attempt
				} else {
					recorded = append(recorded, update.Attempt)
				}
			}
			if report != nil {
				report(pingProgressEvent(update, count, continuous))
			}
		},
	}
	var attempts []checks.PingAttempt
	var summary checks.PingSummary
	var checkErr error
	if continuous {
		summary, checkErr = continuousPingCheck(ctx, options)
		attempts = recorded
	} else {
		attempts, summary, checkErr = pingCheck(ctx, options)
	}
	if continuous && errors.Is(checkErr, context.Canceled) {
		checkErr = nil
	}
	var out bytes.Buffer
	renderPing(&out, attempts, summary)
	attemptLines := make([]string, len(attempts))
	for i, attempt := range attempts {
		if attempt.Err != nil {
			attemptLines[i] = fmt.Sprintf("#%d  %s", attempt.Seq, attempt.Err)
		} else {
			attemptLines[i] = fmt.Sprintf("#%d  %s", attempt.Seq, attempt.Latency.Round(time.Millisecond))
		}
	}
	lossTone := tui.ToneSuccess
	if summary.LossPercent > 0 {
		lossTone = tui.ToneWarning
	}
	if summary.Received == 0 {
		lossTone = tui.ToneDanger
	}
	result := tui.Result{
		Headline: values["host"],
		Summary:  fmt.Sprintf("%d of %d replies · %.1f%% loss", summary.Received, summary.Sent, summary.LossPercent),
		Text:     out.String(),
		Fields: []tui.ResultField{
			{Label: "Sent", Value: strconv.Itoa(summary.Sent)},
			{Label: "Received", Value: strconv.Itoa(summary.Received), Tone: toneForBool(summary.Received > 0)},
			{Label: "Packet loss", Value: fmt.Sprintf("%.1f%%", summary.LossPercent), Tone: lossTone},
			{Label: "Minimum", Value: summary.Min.Round(time.Millisecond).String()},
			{Label: "Average", Value: summary.Avg.Round(time.Millisecond).String()},
			{Label: "Maximum", Value: summary.Max.Round(time.Millisecond).String()},
		},
	}
	if len(attemptLines) > 0 {
		result.Sections = []tui.ResultSection{{Title: "Attempts", Lines: attemptLines}}
	}
	return result, pingResultError(checkErr)
}

func pingProgressEvent(update checks.PingUpdate, count int, continuous bool) tui.ProgressEvent {
	dropped := update.Summary.Sent - update.Summary.Received
	tone := tui.ToneSuccess
	label := fmt.Sprintf("Reply #%d · %s", update.Attempt.Seq, update.Attempt.Latency.Round(time.Millisecond))
	if update.Attempt.Err != nil {
		tone = tui.ToneDanger
		label = fmt.Sprintf("DROP #%d · %v", update.Attempt.Seq, update.Attempt.Err)
	}
	lossTone := tui.ToneSuccess
	if dropped > 0 {
		lossTone = tui.ToneDanger
	}
	total := count
	if continuous {
		total = 0
	}
	return tui.ProgressEvent{
		Label:   label,
		Current: update.Summary.Sent,
		Total:   total,
		Tone:    tone,
		Fields: []tui.ResultField{
			{Label: "Sent", Value: strconv.Itoa(update.Summary.Sent)},
			{Label: "Received", Value: strconv.Itoa(update.Summary.Received), Tone: tui.ToneSuccess},
			{Label: "Dropped", Value: strconv.Itoa(dropped), Tone: lossTone},
			{Label: "Loss", Value: fmt.Sprintf("%.1f%%", update.Summary.LossPercent), Tone: lossTone},
		},
	}
}

func runInteractiveHTTP(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	port, err := strconv.Atoi(values["port"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid port %q: %w", values["port"], err)
	}
	useTLS, err := strconv.ParseBool(values["tls"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid tls value %q: %w", values["tls"], err)
	}
	followRedirects := false
	if values["follow_redirects"] != "" {
		followRedirects, err = strconv.ParseBool(values["follow_redirects"])
		if err != nil {
			return tui.Result{}, fmt.Errorf("invalid follow_redirects value %q: %w", values["follow_redirects"], err)
		}
	}
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, fmt.Errorf("invalid timeout %q: %w", values["timeout"], err)
	}
	reportSingle(report, "Requesting endpoint", false)
	checkResult := httpCheck(ctx, nil, checks.HTTPOptions{Host: values["host"], Port: port, TLS: useTLS, Path: values["path"], FollowRedirects: followRedirects, Timeout: timeout})
	reportSingle(report, "HTTP check complete", true)
	var out bytes.Buffer
	renderHTTP(&out, checkResult)
	statusTone := tui.ToneDanger
	if checkResult.Healthy {
		statusTone = tui.ToneSuccess
	} else if checkResult.Reachable {
		statusTone = tui.ToneWarning
	}
	summary := "Endpoint could not be reached"
	if checkResult.Reachable {
		summary = fmt.Sprintf("Endpoint responded with HTTP %d", checkResult.StatusCode)
	}
	fields := []tui.ResultField{
		{Label: "Status", Value: strconv.Itoa(checkResult.StatusCode), Tone: statusTone},
		{Label: "Latency", Value: checkResult.Latency.Round(time.Millisecond).String()},
		boolResultField("Reachable", checkResult.Reachable),
		boolResultField("Healthy", checkResult.Healthy),
	}
	if checkResult.Location != "" {
		fields = append(fields, tui.ResultField{Label: "Location", Value: checkResult.Location})
	}
	var sections []tui.ResultSection
	if timingLines := formatHTTPTimingLines(checkResult.Timings); len(timingLines) > 0 {
		sections = append(sections, tui.ResultSection{Title: "Phase timings", Lines: timingLines})
	}
	return tui.Result{
		Headline: checkResult.URL,
		Summary:  summary,
		Text:     out.String(),
		Fields:   fields,
		Sections: sections,
	}, httpResultError(checkResult)
}

func formatHTTPTimingLines(timings checks.HTTPTimings) []string {
	total := timings.Total
	if total <= 0 {
		return nil
	}
	phases := []struct {
		label string
		d     time.Duration
	}{
		{"DNS lookup", timings.DNSLookup},
		{"TCP connect", timings.TCPConnect},
		{"TLS handshake", timings.TLSHandshake},
		{"Server TTFB", timings.TTFB},
	}
	var lines []string
	const barWidth = 20
	for _, phase := range phases {
		if phase.d <= 0 {
			continue
		}
		filled := min(barWidth, max(1, int((phase.d*barWidth)/total)))
		bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
		lines = append(lines, fmt.Sprintf("%-14s %s  %s", phase.label, bar, phase.d.Round(time.Microsecond)))
	}
	lines = append(lines, fmt.Sprintf("%-14s %s  %s", "Total", strings.Repeat("─", barWidth), total.Round(time.Microsecond)))
	return lines
}

func runInteractiveDNS(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, err
	}
	reportSingle(report, "Querying DNS", false)
	checkResult := dnsCheck(ctx, nil, checks.DNSOptions{Host: values["host"], Type: strings.ToUpper(values["type"]), Server: values["server"], Timeout: timeout})
	reportSingle(report, "DNS query complete", true)
	var out bytes.Buffer
	renderDNS(&out, checkResult)
	resolver := checkResult.Server
	if resolver == "" {
		resolver = "system default"
	}
	return tui.Result{
		Headline: fmt.Sprintf("%s  %s", checkResult.Type, checkResult.Query),
		Summary:  fmt.Sprintf("%d records returned", len(checkResult.Records)),
		Text:     out.String(),
		Fields: []tui.ResultField{
			{Label: "Resolver", Value: resolver},
			{Label: "Latency", Value: checkResult.Latency.Round(time.Millisecond).String()},
			boolResultField("Reachable", checkResult.Reachable),
			boolResultField("Healthy", checkResult.Healthy),
		},
		Sections: []tui.ResultSection{{Title: "Records", Lines: checkResult.Records}},
	}, dnsResultError(checkResult)
}

func runInteractiveTCP(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, err
	}
	targets, err := checks.ExpandTCPTarget(values["address"])
	if err != nil {
		return tui.Result{}, err
	}
	if len(targets) == 1 {
		reportSingle(report, "Opening TCP connection", false)
		checkResult := tcpCheck(ctx, nil, checks.TCPOptions{Address: targets[0], Timeout: timeout})
		reportSingle(report, "TCP check complete", true)
		var out bytes.Buffer
		renderTCP(&out, checkResult)
		summary := "Connection failed"
		if checkResult.Reachable {
			summary = "Connection established"
		}
		return tui.Result{
			Headline: checkResult.Address,
			Summary:  summary,
			Text:     out.String(),
			Fields: []tui.ResultField{
				boolResultField("Reachable", checkResult.Reachable),
				{Label: "Latency", Value: checkResult.Latency.Round(time.Millisecond).String()},
				{Label: "Remote", Value: checkResult.RemoteAddress},
				{Label: "Local", Value: checkResult.LocalAddress},
			},
		}, tcpResultError(checkResult)
	}
	reportSingle(report, fmt.Sprintf("Scanning %d TCP ports", len(targets)), false)
	batch := checks.RunBatch(ctx, targets, 8, func(runCtx context.Context, target string) checks.TCPResult {
		return tcpCheck(runCtx, nil, checks.TCPOptions{Address: target, Timeout: timeout})
	})
	reportSingle(report, "TCP scan complete", true)
	var out bytes.Buffer
	lines := make([]string, len(batch))
	openCount := 0
	for i, item := range batch {
		if item.Result.Reachable {
			openCount++
			fmt.Fprintf(&out, "%s reachable=true latency=%s\n", item.Item, item.Result.Latency.Round(time.Millisecond))
			lines[i] = fmt.Sprintf("OPEN    %-22s %s", item.Item, item.Result.Latency.Round(time.Millisecond))
		} else {
			fmt.Fprintf(&out, "%s reachable=false latency=%s error=%v\n", item.Item, item.Result.Latency.Round(time.Millisecond), item.Result.Err)
			lines[i] = fmt.Sprintf("CLOSED  %-22s %v", item.Item, item.Result.Err)
		}
	}
	var runErr error
	tone := tui.ToneSuccess
	if openCount == 0 {
		tone = tui.ToneDanger
		runErr = newCommandError(exitUnreachable, fmt.Errorf("all %d TCP ports unreachable", len(batch)))
	} else if openCount < len(batch) {
		tone = tui.ToneWarning
		runErr = newCommandError(exitUnhealthy, fmt.Errorf("%d of %d TCP ports unreachable", len(batch)-openCount, len(batch)))
	}
	return tui.Result{
		Headline: values["address"],
		Summary:  fmt.Sprintf("%d of %d ports reachable", openCount, len(batch)),
		Text:     out.String(),
		Fields: []tui.ResultField{
			{Label: "Checked", Value: strconv.Itoa(len(batch))},
			{Label: "Open", Value: strconv.Itoa(openCount), Tone: tone},
			{Label: "Closed", Value: strconv.Itoa(len(batch) - openCount)},
		},
		Sections: []tui.ResultSection{{Title: "Port results", Lines: lines}},
	}, runErr
}

func runInteractiveTLS(ctx context.Context, values map[string]string, report func(tui.ProgressEvent)) (tui.Result, error) {
	timeout, err := time.ParseDuration(values["timeout"])
	if err != nil {
		return tui.Result{}, err
	}
	warnBefore, err := parseExtendedDuration(values["warn_before"])
	if err != nil {
		return tui.Result{}, err
	}
	reportSingle(report, "Negotiating TLS", false)
	checkResult := tlsCheck(ctx, checks.TLSOptions{Address: values["address"], ServerName: values["server_name"], Timeout: timeout, WarnBefore: warnBefore})
	reportSingle(report, "TLS inspection complete", true)
	var out bytes.Buffer
	renderTLS(&out, checkResult)
	summary := "TLS negotiation failed"
	if checkResult.Verified {
		summary = "Certificate chain verified"
	}
	if checkResult.ExpiringSoon {
		summary = "Certificate expires within the warning window"
	}
	certificateLines := make([]string, 0, len(checkResult.Certificates)*3)
	for i, certificate := range checkResult.Certificates {
		certificateLines = append(certificateLines,
			fmt.Sprintf("%d. %s", i+1, certificate.Subject),
			"   Issuer  "+certificate.Issuer,
			"   Expires "+certificate.NotAfter.UTC().Format(time.RFC3339),
		)
	}
	verifiedField := boolResultField("Verified", checkResult.Verified)
	if checkResult.ExpiringSoon {
		verifiedField.Tone = tui.ToneWarning
	}
	return tui.Result{
		Headline: checkResult.Address,
		Summary:  summary,
		Text:     out.String(),
		Fields: []tui.ResultField{
			verifiedField,
			{Label: "Server name", Value: checkResult.ServerName},
			{Label: "Protocol", Value: checkResult.Version},
			{Label: "Cipher", Value: checkResult.CipherSuite},
			{Label: "Latency", Value: checkResult.Latency.Round(time.Millisecond).String()},
		},
		Sections: []tui.ResultSection{{Title: "Certificates", Lines: certificateLines}},
	}, tlsResultError(checkResult)
}

func boolResultField(label string, value bool) tui.ResultField {
	text := "no"
	if value {
		text = "yes"
	}
	return tui.ResultField{Label: label, Value: text, Tone: toneForBool(value)}
}

func toneForBool(value bool) tui.ResultTone {
	if value {
		return tui.ToneSuccess
	}
	return tui.ToneDanger
}

func reportSingle(report func(tui.ProgressEvent), label string, complete bool) {
	if report == nil {
		return
	}
	current := 0
	if complete {
		current = 1
	}
	report(tui.ProgressEvent{Label: label, Current: current, Total: 1})
}

func validateRequired(value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func validateCIDR(value string) error {
	_, err := subnet.InspectCIDR(value)
	return err
}

func validatePrefixes(value string) error {
	_, err := subnet.ParsePrefixSpecs(value)
	return err
}

func validatePositiveInteger(value string) error {
	number, err := strconv.Atoi(value)
	if err != nil || number <= 0 {
		return fmt.Errorf("must be a positive integer")
	}
	return nil
}

func validateDuration(value string) error {
	duration, err := time.ParseDuration(value)
	if err != nil || duration <= 0 {
		return fmt.Errorf("must be a positive duration")
	}
	return nil
}

func validateExtendedDuration(value string) error {
	_, err := parseExtendedDuration(value)
	return err
}

func validateDNSRecordType(value string) error {
	if !supportedDNSType(strings.ToUpper(strings.TrimSpace(value))) {
		return fmt.Errorf("use A, AAAA, CNAME, MX, NS, TXT, or PTR")
	}
	return nil
}

func validateHostPort(value string) error {
	host, port, err := net.SplitHostPort(value)
	if err != nil || host == "" || port == "" {
		return fmt.Errorf("expected HOST:PORT")
	}
	return nil
}

func validateTCPTarget(value string) error {
	_, err := checks.ExpandTCPTarget(value)
	return err
}

func validatePort(value string) error {
	port, err := strconv.Atoi(value)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("must be between 1 and 65535")
	}
	return nil
}

func validateIPAddress(value string) error {
	if net.ParseIP(strings.TrimSpace(value)) == nil {
		return fmt.Errorf("invalid IP address")
	}
	return nil
}

func splitCIDRList(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == '\t'
	})
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

func validateCIDRList(value string) error {
	items := splitCIDRList(value)
	if len(items) == 0 {
		return fmt.Errorf("enter at least one CIDR")
	}
	_, err := subnet.SummarizeCIDRs(items)
	return err
}

func validateHopLimit(value string) error {
	n, err := strconv.Atoi(value)
	if err != nil || n < 1 || n > 64 {
		return fmt.Errorf("must be between 1 and 64")
	}
	return nil
}

func validateBoolean(value string) error {
	_, err := strconv.ParseBool(value)
	if err != nil {
		return fmt.Errorf("use true or false")
	}
	return nil
}
