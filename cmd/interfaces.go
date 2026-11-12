package cmd

import (
	"fmt"
	"io"
	"strings"

	"github.com/mathwro/nwcli/internal/checks"
	"github.com/spf13/cobra"
)

var listInterfaces = checks.ListInterfaces

func newInterfacesCommand() *cobra.Command {
	var includeAll bool
	cmd := &cobra.Command{
		Use:     "interfaces",
		Aliases: []string{"ifaces", "ip"},
		Short:   "Show local network interfaces and assigned CIDRs",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			items, err := listInterfaces(includeAll)
			if err != nil {
				return newCommandError(exitUnreachable, err)
			}
			return emitCommandResult(cmd, "interfaces", interfacesOutput(items), nil, func(w io.Writer) error {
				return renderInterfaces(w, items)
			})
		},
	}
	cmd.Flags().BoolVarP(&includeAll, "all", "a", false, "include down interfaces")
	return cmd
}

func renderInterfaces(w io.Writer, items []checks.InterfaceInfo) error {
	for i, item := range items {
		if i > 0 {
			if _, err := fmt.Fprintln(w); err != nil {
				return err
			}
		}
		mac := item.HardwareAddr
		if mac == "" {
			mac = "-"
		}
		if _, err := fmt.Fprintf(w, "%s  mtu=%d  mac=%s  flags=%s\n", item.Name, item.MTU, mac, strings.Join(item.Flags, ",")); err != nil {
			return err
		}
		if len(item.Addresses) == 0 {
			if _, err := fmt.Fprintln(w, "  (no addresses)"); err != nil {
				return err
			}
			continue
		}
		for _, addr := range item.Addresses {
			if _, err := fmt.Fprintf(w, "  %s\n", addr); err != nil {
				return err
			}
		}
	}
	return nil
}

type interfaceJSON struct {
	Index        int      `json:"index"`
	Name         string   `json:"name"`
	MTU          int      `json:"mtu"`
	Up           bool     `json:"up"`
	Loopback     bool     `json:"loopback"`
	HardwareAddr string   `json:"hardware_addr,omitempty"`
	Flags        []string `json:"flags"`
	Addresses    []string `json:"addresses"`
}

func interfacesOutput(items []checks.InterfaceInfo) []interfaceJSON {
	out := make([]interfaceJSON, len(items))
	for i, item := range items {
		out[i] = interfaceJSON{
			Index:        item.Index,
			Name:         item.Name,
			MTU:          item.MTU,
			Up:           item.Up,
			Loopback:     item.Loopback,
			HardwareAddr: item.HardwareAddr,
			Flags:        item.Flags,
			Addresses:    item.Addresses,
		}
	}
	return out
}

func newCompletionCommand() *cobra.Command {
	return &cobra.Command{
		Use:       "completion [bash|zsh|fish|powershell]",
		Short:     "Generate shell completion script",
		ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
		Args:      cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			out := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(out, true)
			case "zsh":
				return root.GenZshCompletion(out)
			case "fish":
				return root.GenFishCompletion(out, true)
			case "powershell":
				return root.GenPowerShellCompletionWithDesc(out)
			default:
				return newCommandError(exitUsage, fmt.Errorf("unsupported shell %q", args[0]))
			}
		},
	}
}
