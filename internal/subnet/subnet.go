package subnet

import (
	"encoding/binary"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
)

type AllocationPlan struct {
	Parent      *net.IPNet
	ParentFirst net.IP
	ParentLast  net.IP
	ParentSize  uint32
	Blocks      []Block
	Unused      *Range
}

type PrefixSpec struct {
	Name   string
	Prefix int
}

type Block struct {
	Index       int
	Name        string
	CIDR        *net.IPNet
	First       net.IP
	Last        net.IP
	Size        uint32
	StartOffset uint32
}

type Range struct {
	First       net.IP
	Last        net.IP
	Size        uint32
	StartOffset uint32
}

func ParsePrefixes(raw string) ([]int, error) {
	specs, err := ParsePrefixSpecs(raw)
	if err != nil {
		return nil, err
	}
	prefixes := make([]int, len(specs))
	for i, spec := range specs {
		prefixes[i] = spec.Prefix
	}
	return prefixes, nil
}

func ParsePrefixSpecs(raw string) ([]PrefixSpec, error) {
	parts := strings.Split(raw, ",")
	specs := make([]PrefixSpec, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			return nil, fmt.Errorf("empty prefix in %q", raw)
		}
		var name string
		prefixPart := part
		if idx := strings.LastIndexAny(part, ":="); idx >= 0 {
			name = strings.TrimSpace(part[:idx])
			prefixPart = strings.TrimSpace(part[idx+1:])
			if name == "" {
				return nil, fmt.Errorf("empty block name in %q", part)
			}
		}
		prefix, err := strconv.Atoi(strings.TrimPrefix(prefixPart, "/"))
		if err != nil {
			return nil, fmt.Errorf("invalid prefix %q: %w", part, err)
		}
		if prefix < 0 || prefix > 32 {
			return nil, fmt.Errorf("prefix /%d outside IPv4 range /0-/32", prefix)
		}
		specs = append(specs, PrefixSpec{Name: name, Prefix: prefix})
	}
	return specs, nil
}

func Plan(parentCIDR string, prefixes []int) (AllocationPlan, error) {
	specs := make([]PrefixSpec, len(prefixes))
	for i, prefix := range prefixes {
		specs[i] = PrefixSpec{Prefix: prefix}
	}
	return PlanSpecs(parentCIDR, specs, false)
}

func PlanSpecs(parentCIDR string, specs []PrefixSpec, sortLargestFirst bool) (AllocationPlan, error) {
	ip, network, err := net.ParseCIDR(parentCIDR)
	if err != nil {
		return AllocationPlan{}, fmt.Errorf("invalid parent CIDR %q: %w", parentCIDR, err)
	}
	parentIP := ip.To4()
	if parentIP == nil {
		return AllocationPlan{}, fmt.Errorf("parent CIDR %q is not IPv4", parentCIDR)
	}
	ones, bits := network.Mask.Size()
	if bits != 32 {
		return AllocationPlan{}, fmt.Errorf("parent CIDR %q is not IPv4", parentCIDR)
	}
	if ones == 0 {
		return AllocationPlan{}, fmt.Errorf("parent CIDR %q is too large for IPv4 address counts", parentCIDR)
	}
	base := ipToUint32(parentIP.Mask(network.Mask))
	parentSize := uint32(1) << uint32(32-ones)
	plan := AllocationPlan{
		Parent:      &net.IPNet{IP: uint32ToIP(base), Mask: network.Mask},
		ParentFirst: uint32ToIP(base),
		ParentLast:  uint32ToIP(base + parentSize - 1),
		ParentSize:  parentSize,
		Blocks:      make([]Block, 0, len(specs)),
	}

	ordered := append([]PrefixSpec(nil), specs...)
	if sortLargestFirst {
		sort.SliceStable(ordered, func(i, j int) bool {
			return ordered[i].Prefix < ordered[j].Prefix
		})
	}

	var offset uint32
	for i, spec := range ordered {
		prefix := spec.Prefix
		if prefix < ones {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d is shorter than parent /%d", prefix, ones)
		}
		if prefix > 32 {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d outside IPv4 range /0-/32", prefix)
		}
		size := uint32(1) << uint32(32-prefix)
		if offset%size != 0 {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d at position %d is not aligned after previous allocation (pass --sort to pack largest-first)", prefix, i+1)
		}
		if offset > parentSize || size > parentSize-offset {
			return AllocationPlan{}, fmt.Errorf("child prefix /%d at position %d does not fit in %s", prefix, i+1, plan.Parent.String())
		}
		first := base + offset
		last := first + size - 1
		plan.Blocks = append(plan.Blocks, Block{
			Index:       i + 1,
			Name:        spec.Name,
			CIDR:        &net.IPNet{IP: uint32ToIP(first), Mask: net.CIDRMask(prefix, 32)},
			First:       uint32ToIP(first),
			Last:        uint32ToIP(last),
			Size:        size,
			StartOffset: offset,
		})
		offset += size
	}
	if offset < parentSize {
		plan.Unused = &Range{First: uint32ToIP(base + offset), Last: uint32ToIP(base + parentSize - 1), Size: parentSize - offset, StartOffset: offset}
	}
	return plan, nil
}

func ipToUint32(ip net.IP) uint32 {
	return binary.BigEndian.Uint32(ip.To4())
}

func uint32ToIP(v uint32) net.IP {
	var b [4]byte
	binary.BigEndian.PutUint32(b[:], v)
	return net.IPv4(b[0], b[1], b[2], b[3])
}
