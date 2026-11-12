package checks

import (
	"fmt"
	"net"
	"sort"
)

type InterfaceInfo struct {
	Index        int
	Name         string
	MTU          int
	Up           bool
	Loopback     bool
	HardwareAddr string
	Flags        []string
	Addresses    []string
}

func ListInterfaces(includeDown bool) ([]InterfaceInfo, error) {
	return collectInterfaces(includeDown, net.Interfaces, func(iface net.Interface) ([]net.Addr, error) {
		return iface.Addrs()
	})
}

func collectInterfaces(
	includeDown bool,
	fetch func() ([]net.Interface, error),
	addrs func(net.Interface) ([]net.Addr, error),
) ([]InterfaceInfo, error) {
	rawList, err := fetch()
	if err != nil {
		return nil, fmt.Errorf("list network interfaces: %w", err)
	}
	items := make([]InterfaceInfo, 0, len(rawList))
	for _, iface := range rawList {
		up := iface.Flags&net.FlagUp != 0
		if !includeDown && !up {
			continue
		}
		addrList, err := addrs(iface)
		if err != nil {
			return nil, fmt.Errorf("interface %s addresses: %w", iface.Name, err)
		}
		cidrs := make([]string, 0, len(addrList))
		for _, addr := range addrList {
			cidrs = append(cidrs, addr.String())
		}
		sort.Strings(cidrs)
		items = append(items, InterfaceInfo{
			Index:        iface.Index,
			Name:         iface.Name,
			MTU:          iface.MTU,
			Up:           up,
			Loopback:     iface.Flags&net.FlagLoopback != 0,
			HardwareAddr: iface.HardwareAddr.String(),
			Flags:        formatInterfaceFlags(iface.Flags),
			Addresses:    cidrs,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].Index < items[j].Index
	})
	return items, nil
}

func formatInterfaceFlags(flags net.Flags) []string {
	var out []string
	if flags&net.FlagUp != 0 {
		out = append(out, "up")
	} else {
		out = append(out, "down")
	}
	if flags&net.FlagLoopback != 0 {
		out = append(out, "loopback")
	}
	if flags&net.FlagBroadcast != 0 {
		out = append(out, "broadcast")
	}
	if flags&net.FlagPointToPoint != 0 {
		out = append(out, "pointtopoint")
	}
	if flags&net.FlagMulticast != 0 {
		out = append(out, "multicast")
	}
	return out
}
