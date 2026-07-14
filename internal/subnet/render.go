package subnet

import (
	"fmt"
	"io"
	"strings"
)

const barWidth = 32

func Render(w io.Writer, plan AllocationPlan) error {
	if _, err := fmt.Fprintf(w, "Parent: %s  addresses: %d\n\n", plan.Parent.String(), plan.ParentSize); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "#  CIDR           FIRST        LAST          ADDRS  MAP"); err != nil {
		return err
	}
	for _, block := range plan.Blocks {
		if _, err := fmt.Fprintf(w, "%-2d %-14s %-12s %-12s %-6d %s\n", block.Index, block.CIDR.String(), block.First.String(), block.Last.String(), block.Size, allocationBar(plan.ParentSize, block.StartOffset, block.Size)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if plan.Unused != nil {
		_, err := fmt.Fprintf(w, "Unused: %s - %s  addresses: %d\n", plan.Unused.First.String(), plan.Unused.Last.String(), plan.Unused.Size)
		return err
	}
	_, err := fmt.Fprintln(w, "Unused: none")
	return err
}

func allocationBar(parentSize, start, size uint32) string {
	if parentSize == 0 || size == 0 {
		return strings.Repeat(".", barWidth)
	}
	startCell := int((uint64(start) * barWidth) / uint64(parentSize))
	endCell := int((uint64(start+size) * barWidth) / uint64(parentSize))
	if endCell <= startCell {
		endCell = startCell + 1
	}
	if endCell > barWidth {
		endCell = barWidth
	}
	var b strings.Builder
	b.Grow(barWidth)
	for i := 0; i < barWidth; i++ {
		if i >= startCell && i < endCell {
			b.WriteByte('#')
		} else {
			b.WriteByte('.')
		}
	}
	return b.String()
}
