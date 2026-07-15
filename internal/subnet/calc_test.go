package subnet

import (
	"reflect"
	"testing"
)

func TestInspectCIDRReportsRangeAndAddressCount(t *testing.T) {
	info, err := InspectCIDR("10.0.0.37/24")
	if err != nil {
		t.Fatalf("InspectCIDR returned error: %v", err)
	}
	if info.CIDR != "10.0.0.0/24" || info.First != "10.0.0.0" || info.Last != "10.0.0.255" || info.Addresses != "256" || info.Family != "IPv4" {
		t.Fatalf("info = %#v", info)
	}

	ipv6, err := InspectCIDR("2001:db8::1/126")
	if err != nil {
		t.Fatalf("InspectCIDR IPv6 returned error: %v", err)
	}
	if ipv6.CIDR != "2001:db8::/126" || ipv6.Last != "2001:db8::3" || ipv6.Addresses != "4" || ipv6.Family != "IPv6" {
		t.Fatalf("ipv6 = %#v", ipv6)
	}
}

func TestContainsAndOverlapsCIDR(t *testing.T) {
	contains, err := ContainsCIDR("10.0.0.0/24", "10.0.0.42")
	if err != nil || !contains {
		t.Fatalf("ContainsCIDR = %t, %v", contains, err)
	}
	overlaps, err := OverlapsCIDR("10.0.0.0/24", "10.0.0.128/25")
	if err != nil || !overlaps {
		t.Fatalf("OverlapsCIDR = %t, %v", overlaps, err)
	}
	overlaps, err = OverlapsCIDR("10.0.0.0/24", "10.0.1.0/24")
	if err != nil || overlaps {
		t.Fatalf("OverlapsCIDR disjoint = %t, %v", overlaps, err)
	}
}

func TestSplitCIDRProducesEqualChildren(t *testing.T) {
	children, err := SplitCIDR("10.0.0.0/24", 26)
	if err != nil {
		t.Fatalf("SplitCIDR returned error: %v", err)
	}
	want := []string{"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26"}
	if !reflect.DeepEqual(children, want) {
		t.Fatalf("children = %#v, want %#v", children, want)
	}
}

func TestSummarizeCIDRsCollapsesAdjacentAndContainedPrefixes(t *testing.T) {
	result, err := SummarizeCIDRs([]string{
		"10.0.0.0/26", "10.0.0.64/26", "10.0.0.128/26", "10.0.0.192/26", "10.0.0.1/32",
	})
	if err != nil {
		t.Fatalf("SummarizeCIDRs returned error: %v", err)
	}
	if !reflect.DeepEqual(result, []string{"10.0.0.0/24"}) {
		t.Fatalf("result = %#v", result)
	}
}
