package cmd

import (
	"fmt"
	"io"

	"github.com/mathwro/nwcli/internal/subnet"
	"github.com/spf13/cobra"
)

func newSubnetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "Subnet planning tools",
	}
	cmd.AddCommand(
		newSubnetCarveCommand(),
		newSubnetInspectCommand(),
		newSubnetContainsCommand(),
		newSubnetOverlapCommand(),
		newSubnetSplitCommand(),
		newSubnetSummarizeCommand(),
	)
	return cmd
}

func newSubnetCarveCommand() *cobra.Command {
	var parentCIDR string
	var sizes string
	cmd := &cobra.Command{
		Use:   "carve",
		Short: "Carve child CIDRs from an IPv4 parent subnet",
		RunE: func(cmd *cobra.Command, args []string) error {
			prefixes, err := subnet.ParsePrefixes(sizes)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			plan, err := subnet.Plan(parentCIDR, prefixes)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			return emitCommandResult(cmd, "subnet carve", subnetPlanOutput(plan), nil, func(w io.Writer) error {
				return subnet.Render(w, plan)
			})
		},
	}
	cmd.Flags().StringVar(&parentCIDR, "cidr", "", "parent IPv4 CIDR")
	cmd.Flags().StringVar(&sizes, "sizes", "", "comma-separated child prefix lengths")
	_ = cmd.MarkFlagRequired("cidr")
	_ = cmd.MarkFlagRequired("sizes")
	return cmd
}

func newSubnetInspectCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "inspect CIDR",
		Short: "Inspect an IPv4 or IPv6 CIDR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			info, err := subnet.InspectCIDR(args[0])
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			return emitCommandResult(cmd, "subnet inspect", cidrInfoOutput(info), nil, func(w io.Writer) error {
				_, err := fmt.Fprintf(w, "CIDR: %s\nFamily: %s\nPrefix: /%d\nFirst: %s\nLast: %s\nAddresses: %s\n", info.CIDR, info.Family, info.Prefix, info.First, info.Last, info.Addresses)
				return err
			})
		},
	}
}

func newSubnetContainsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "contains CIDR IP",
		Short: "Test whether a CIDR contains an IP address",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			contains, err := subnet.ContainsCIDR(args[0], args[1])
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			result := map[string]any{"cidr": args[0], "address": args[1], "contains": contains}
			return emitCommandResult(cmd, "subnet contains", result, nil, func(w io.Writer) error {
				_, err := fmt.Fprintln(w, contains)
				return err
			})
		},
	}
}

func newSubnetOverlapCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "overlap CIDR CIDR",
		Short: "Test whether two CIDRs overlap",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			overlap, err := subnet.OverlapsCIDR(args[0], args[1])
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			result := map[string]any{"cidr_a": args[0], "cidr_b": args[1], "overlap": overlap}
			return emitCommandResult(cmd, "subnet overlap", result, nil, func(w io.Writer) error {
				_, err := fmt.Fprintln(w, overlap)
				return err
			})
		},
	}
}

func newSubnetSplitCommand() *cobra.Command {
	var childPrefix int
	cmd := &cobra.Command{
		Use:   "split CIDR",
		Short: "Split a CIDR into equal child prefixes",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			children, err := subnet.SplitCIDR(args[0], childPrefix)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			result := map[string]any{"parent": args[0], "prefix": childPrefix, "children": children}
			return emitCommandResult(cmd, "subnet split", result, nil, func(w io.Writer) error {
				for _, child := range children {
					if _, err := fmt.Fprintln(w, child); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
	cmd.Flags().IntVar(&childPrefix, "prefix", 0, "child prefix length")
	_ = cmd.MarkFlagRequired("prefix")
	return cmd
}

func newSubnetSummarizeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "summarize CIDR...",
		Short: "Collapse adjacent and contained CIDRs",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			prefixes, err := subnet.SummarizeCIDRs(args)
			if err != nil {
				return newCommandError(exitUsage, err)
			}
			result := map[string]any{"prefixes": prefixes}
			return emitCommandResult(cmd, "subnet summarize", result, nil, func(w io.Writer) error {
				for _, prefix := range prefixes {
					if _, err := fmt.Fprintln(w, prefix); err != nil {
						return err
					}
				}
				return nil
			})
		},
	}
}

type cidrInfoJSON struct {
	CIDR      string `json:"cidr"`
	Family    string `json:"family"`
	Prefix    int    `json:"prefix"`
	First     string `json:"first"`
	Last      string `json:"last"`
	Addresses string `json:"addresses"`
}

func cidrInfoOutput(info subnet.CIDRInfo) cidrInfoJSON {
	return cidrInfoJSON{
		CIDR: info.CIDR, Family: info.Family, Prefix: info.Prefix,
		First: info.First, Last: info.Last, Addresses: info.Addresses,
	}
}

type subnetBlockOutput struct {
	Index     int    `json:"index"`
	CIDR      string `json:"cidr"`
	First     string `json:"first"`
	Last      string `json:"last"`
	Addresses uint32 `json:"addresses"`
}

type subnetRangeOutput struct {
	First     string `json:"first"`
	Last      string `json:"last"`
	Addresses uint32 `json:"addresses"`
}

type subnetPlanJSON struct {
	Parent    string              `json:"parent"`
	Addresses uint32              `json:"addresses"`
	Blocks    []subnetBlockOutput `json:"blocks"`
	Unused    *subnetRangeOutput  `json:"unused,omitempty"`
}

func subnetPlanOutput(plan subnet.AllocationPlan) subnetPlanJSON {
	result := subnetPlanJSON{
		Parent:    plan.Parent.String(),
		Addresses: plan.ParentSize,
		Blocks:    make([]subnetBlockOutput, len(plan.Blocks)),
	}
	for i, block := range plan.Blocks {
		result.Blocks[i] = subnetBlockOutput{
			Index:     block.Index,
			CIDR:      block.CIDR.String(),
			First:     block.First.String(),
			Last:      block.Last.String(),
			Addresses: block.Size,
		}
	}
	if plan.Unused != nil {
		result.Unused = &subnetRangeOutput{
			First:     plan.Unused.First.String(),
			Last:      plan.Unused.Last.String(),
			Addresses: plan.Unused.Size,
		}
	}
	return result
}
