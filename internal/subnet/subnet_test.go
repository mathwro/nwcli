package subnet

import (
	"bytes"
	"strings"
	"testing"
)

func TestParsePrefixesParsesCommaSeparatedCIDRLengths(t *testing.T) {
	got, err := ParsePrefixes("27,27,28,28")
	if err != nil {
		t.Fatalf("ParsePrefixes returned error: %v", err)
	}
	want := []int{27, 27, 28, 28}
	if len(got) != len(want) {
		t.Fatalf("len(ParsePrefixes) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("ParsePrefixes[%d] = %d, want %d", i, got[i], want[i])
		}
	}
}

func TestPlanAllocatesIPv4BlocksInRequestedOrder(t *testing.T) {
	plan, err := Plan("10.0.0.0/24", []int{27, 27, 28, 28})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	wantCIDRs := []string{"10.0.0.0/27", "10.0.0.32/27", "10.0.0.64/28", "10.0.0.80/28"}
	wantFirst := []string{"10.0.0.0", "10.0.0.32", "10.0.0.64", "10.0.0.80"}
	wantLast := []string{"10.0.0.31", "10.0.0.63", "10.0.0.79", "10.0.0.95"}
	wantSize := []uint32{32, 32, 16, 16}

	if plan.ParentSize != 256 {
		t.Fatalf("ParentSize = %d, want 256", plan.ParentSize)
	}
	if len(plan.Blocks) != len(wantCIDRs) {
		t.Fatalf("len(Blocks) = %d, want %d", len(plan.Blocks), len(wantCIDRs))
	}
	for i, block := range plan.Blocks {
		if block.CIDR.String() != wantCIDRs[i] || block.First.String() != wantFirst[i] || block.Last.String() != wantLast[i] || block.Size != wantSize[i] {
			t.Fatalf("block %d = %s %s-%s size %d, want %s %s-%s size %d", i, block.CIDR, block.First, block.Last, block.Size, wantCIDRs[i], wantFirst[i], wantLast[i], wantSize[i])
		}
	}
	if plan.Unused == nil {
		t.Fatal("Unused is nil, want remaining parent range")
	}
	if plan.Unused.First.String() != "10.0.0.96" || plan.Unused.Last.String() != "10.0.0.255" || plan.Unused.Size != 160 {
		t.Fatalf("Unused = %s-%s size %d, want 10.0.0.96-10.0.0.255 size 160", plan.Unused.First, plan.Unused.Last, plan.Unused.Size)
	}
}

func TestPlanRejectsChildPrefixShorterThanParent(t *testing.T) {
	_, err := Plan("10.0.0.0/24", []int{23})
	if err == nil {
		t.Fatal("Plan returned nil error, want child prefix validation error")
	}
	if !strings.Contains(err.Error(), "child prefix /23 is shorter than parent /24") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestPlanRejectsUnalignedLargerBlockAfterSmallerBlock(t *testing.T) {
	_, err := Plan("10.0.0.0/24", []int{27, 26})
	if err == nil {
		t.Fatal("Plan returned nil error, want CIDR alignment error")
	}
	if !strings.Contains(err.Error(), "not aligned") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestPlanRejectsAllocationsThatExceedParent(t *testing.T) {
	_, err := Plan("10.0.0.0/30", []int{31, 31, 31})
	if err == nil {
		t.Fatal("Plan returned nil error, want capacity error")
	}
	if !strings.Contains(err.Error(), "does not fit") {
		t.Fatalf("error = %q", err.Error())
	}
}

func TestRenderPrintsRowsAndUnusedRange(t *testing.T) {
	plan, err := Plan("10.0.0.0/24", []int{27, 27, 28, 28})
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}
	var out bytes.Buffer
	if err := Render(&out, plan); err != nil {
		t.Fatalf("Render returned error: %v", err)
	}
	text := out.String()
	for _, fragment := range []string{
		"Parent: 10.0.0.0/24  addresses: 256",
		"1  10.0.0.0/27",
		"2  10.0.0.32/27",
		"3  10.0.0.64/28",
		"4  10.0.0.80/28",
		"Unused: 10.0.0.96 - 10.0.0.255  addresses: 160",
	} {
		if !strings.Contains(text, fragment) {
			t.Fatalf("render output missing %q:\n%s", fragment, text)
		}
	}
}
