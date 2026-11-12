package subnet

import (
	"fmt"
	"math/big"
	"net/netip"
	"sort"
)

const maxSplitPrefixes = 1 << 16

type CIDRInfo struct {
	CIDR            string
	Family          string
	Prefix          int
	Netmask         string
	Wildcard        string
	First           string
	Last            string
	UsableFirst     string
	UsableLast      string
	Addresses       string
	UsableAddresses string
	Scope           string
}

func InspectCIDR(raw string) (CIDRInfo, error) {
	prefix, err := parsePrefix(raw)
	if err != nil {
		return CIDRInfo{}, err
	}
	first := prefix.Addr()
	count := addressCount(prefix)
	lastValue := addrToBig(first)
	lastValue.Add(lastValue, new(big.Int).Sub(new(big.Int).Set(count), big.NewInt(1)))
	last, err := bigToAddr(lastValue, first.Is4())
	if err != nil {
		return CIDRInfo{}, err
	}
	family := "IPv6"
	var netmask, wildcard string
	usableFirst := first.String()
	usableLast := last.String()
	usableCount := new(big.Int).Set(count)
	if first.Is4() {
		family = "IPv4"
		bits := prefix.Bits()
		var mask uint32
		if bits > 0 {
			mask = ^uint32(0) << uint(32-bits)
		}
		netmask = uint32ToIP(mask).String()
		wildcard = uint32ToIP(^mask).String()
		if bits <= 30 {
			uFirst, _ := bigToAddr(new(big.Int).Add(addrToBig(first), big.NewInt(1)), true)
			uLast, _ := bigToAddr(new(big.Int).Sub(new(big.Int).Set(lastValue), big.NewInt(1)), true)
			usableFirst = uFirst.String()
			usableLast = uLast.String()
			usableCount.Sub(usableCount, big.NewInt(2))
		}
	}
	return CIDRInfo{
		CIDR:            prefix.String(),
		Family:          family,
		Prefix:          prefix.Bits(),
		Netmask:         netmask,
		Wildcard:        wildcard,
		First:           first.String(),
		Last:            last.String(),
		UsableFirst:     usableFirst,
		UsableLast:      usableLast,
		Addresses:       count.String(),
		UsableAddresses: usableCount.String(),
		Scope:           classifyScope(prefix),
	}, nil
}

func classifyScope(prefix netip.Prefix) string {
	addr := prefix.Addr()
	if addr.IsUnspecified() {
		return "unspecified"
	}
	if addr.IsLoopback() {
		return "loopback"
	}
	if addr.IsMulticast() {
		return "multicast"
	}
	if addr.IsLinkLocalUnicast() {
		return "link-local"
	}
	if addr.Is4() {
		switch {
		case addr.IsPrivate():
			return "private (RFC 1918)"
		case netip.MustParsePrefix("100.64.0.0/10").Contains(addr):
			return "shared CGNAT (RFC 6598)"
		case netip.MustParsePrefix("192.0.2.0/24").Contains(addr) ||
			netip.MustParsePrefix("198.51.100.0/24").Contains(addr) ||
			netip.MustParsePrefix("203.0.113.0/24").Contains(addr):
			return "documentation (RFC 5737)"
		case netip.MustParsePrefix("198.18.0.0/15").Contains(addr):
			return "benchmarking (RFC 2544)"
		default:
			return "global unicast"
		}
	}
	switch {
	case addr.IsPrivate():
		return "unique local (RFC 4193)"
	case netip.MustParsePrefix("2001:db8::/32").Contains(addr):
		return "documentation (RFC 3849)"
	default:
		return "global unicast"
	}
}

func ContainsCIDR(rawCIDR, rawAddress string) (bool, error) {
	prefix, err := parsePrefix(rawCIDR)
	if err != nil {
		return false, err
	}
	address, err := netip.ParseAddr(rawAddress)
	if err != nil {
		return false, fmt.Errorf("invalid IP address %q: %w", rawAddress, err)
	}
	return prefix.Contains(address), nil
}

func OverlapsCIDR(rawA, rawB string) (bool, error) {
	a, err := parsePrefix(rawA)
	if err != nil {
		return false, err
	}
	b, err := parsePrefix(rawB)
	if err != nil {
		return false, err
	}
	if a.Addr().BitLen() != b.Addr().BitLen() {
		return false, nil
	}
	return a.Contains(b.Addr()) || b.Contains(a.Addr()), nil
}

func SplitCIDR(raw string, childPrefix int) ([]string, error) {
	parent, err := parsePrefix(raw)
	if err != nil {
		return nil, err
	}
	if childPrefix <= parent.Bits() || childPrefix > parent.Addr().BitLen() {
		return nil, fmt.Errorf("child prefix /%d must be greater than /%d and no greater than /%d", childPrefix, parent.Bits(), parent.Addr().BitLen())
	}
	delta := childPrefix - parent.Bits()
	if delta > 16 {
		return nil, fmt.Errorf("split would produce more than %d prefixes; choose a shorter child prefix", maxSplitPrefixes)
	}
	count := 1 << delta
	blockSize := new(big.Int).Lsh(big.NewInt(1), uint(parent.Addr().BitLen()-childPrefix))
	base := addrToBig(parent.Addr())
	children := make([]string, count)
	for i := range count {
		offset := new(big.Int).Mul(new(big.Int).SetInt64(int64(i)), blockSize)
		address, err := bigToAddr(new(big.Int).Add(new(big.Int).Set(base), offset), parent.Addr().Is4())
		if err != nil {
			return nil, err
		}
		children[i] = netip.PrefixFrom(address, childPrefix).String()
	}
	return children, nil
}

func SummarizeCIDRs(raw []string) ([]string, error) {
	prefixes := make([]netip.Prefix, len(raw))
	for i, value := range raw {
		prefix, err := parsePrefix(value)
		if err != nil {
			return nil, err
		}
		prefixes[i] = prefix
	}
	sort.Slice(prefixes, func(i, j int) bool {
		a, b := prefixes[i], prefixes[j]
		if a.Addr().BitLen() != b.Addr().BitLen() {
			return a.Addr().BitLen() < b.Addr().BitLen()
		}
		if cmp := a.Addr().Compare(b.Addr()); cmp != 0 {
			return cmp < 0
		}
		return a.Bits() < b.Bits()
	})

	stack := make([]netip.Prefix, 0, len(prefixes))
	for _, prefix := range prefixes {
		if len(stack) > 0 {
			last := stack[len(stack)-1]
			if last.Addr().BitLen() == prefix.Addr().BitLen() && last.Contains(prefix.Addr()) && last.Bits() <= prefix.Bits() {
				continue
			}
		}
		stack = append(stack, prefix)
		for len(stack) >= 2 && siblingPrefixes(stack[len(stack)-2], stack[len(stack)-1]) {
			parent := netip.PrefixFrom(stack[len(stack)-1].Addr(), stack[len(stack)-1].Bits()-1).Masked()
			stack = append(stack[:len(stack)-2], parent)
		}
	}
	result := make([]string, len(stack))
	for i, prefix := range stack {
		result[i] = prefix.String()
	}
	return result, nil
}

func siblingPrefixes(a, b netip.Prefix) bool {
	if a.Addr().BitLen() != b.Addr().BitLen() || a.Bits() != b.Bits() || a.Bits() == 0 || a == b {
		return false
	}
	return netip.PrefixFrom(a.Addr(), a.Bits()-1).Masked() == netip.PrefixFrom(b.Addr(), b.Bits()-1).Masked()
}

func parsePrefix(raw string) (netip.Prefix, error) {
	prefix, err := netip.ParsePrefix(raw)
	if err != nil {
		return netip.Prefix{}, fmt.Errorf("invalid CIDR %q: %w", raw, err)
	}
	return prefix.Masked(), nil
}

func addressCount(prefix netip.Prefix) *big.Int {
	return new(big.Int).Lsh(big.NewInt(1), uint(prefix.Addr().BitLen()-prefix.Bits()))
}

func addrToBig(address netip.Addr) *big.Int {
	if address.Is4() {
		bytes := address.As4()
		return new(big.Int).SetBytes(bytes[:])
	}
	bytes := address.As16()
	return new(big.Int).SetBytes(bytes[:])
}

func bigToAddr(value *big.Int, ipv4 bool) (netip.Addr, error) {
	size := 16
	if ipv4 {
		size = 4
	}
	bytes := value.Bytes()
	if len(bytes) > size {
		return netip.Addr{}, fmt.Errorf("address value exceeds %d-bit range", size*8)
	}
	padded := make([]byte, size)
	copy(padded[size-len(bytes):], bytes)
	address, ok := netip.AddrFromSlice(padded)
	if !ok {
		return netip.Addr{}, fmt.Errorf("cannot construct address from %d bytes", size)
	}
	return address, nil
}
