package cmd

import (
	"github.com/mathwro/nwcli/internal/subnet"
	"github.com/spf13/cobra"
)

func newSubnetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnet",
		Short: "Subnet planning tools",
	}
	cmd.AddCommand(newSubnetCarveCommand())
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
				return err
			}
			plan, err := subnet.Plan(parentCIDR, prefixes)
			if err != nil {
				return err
			}
			return subnet.Render(cmd.OutOrStdout(), plan)
		},
	}
	cmd.Flags().StringVar(&parentCIDR, "cidr", "", "parent IPv4 CIDR")
	cmd.Flags().StringVar(&sizes, "sizes", "", "comma-separated child prefix lengths")
	_ = cmd.MarkFlagRequired("cidr")
	_ = cmd.MarkFlagRequired("sizes")
	return cmd
}
