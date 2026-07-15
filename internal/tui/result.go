package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type ResultTone string

const (
	ToneNeutral ResultTone = "neutral"
	ToneSuccess ResultTone = "success"
	ToneWarning ResultTone = "warning"
	ToneDanger  ResultTone = "danger"
)

type Result struct {
	Headline   string
	Summary    string
	Fields     []ResultField
	Sections   []ResultSection
	Allocation *AllocationMap
	Text       string
}

type ResultField struct {
	Label string
	Value string
	Tone  ResultTone
}

type ResultSection struct {
	Title string
	Lines []string
}

type AllocationMap struct {
	Start     string
	End       string
	Total     uint64
	Segments  []AllocationSegment
	Available *AllocationSegment
}

type AllocationSegment struct {
	Index  int
	Label  string
	Start  string
	End    string
	Offset uint64
	Size   uint64
}

func renderResultHeader(title string, result Result, runErr error) string {
	var b strings.Builder
	if runErr != nil {
		b.WriteString(errorStyle.Render("!  NEEDS ATTENTION"))
	} else {
		b.WriteString(successStyle.Render("✓  COMPLETE"))
	}
	b.WriteString("  ")
	b.WriteString(titleStyle.Render(title))
	if result.Headline != "" {
		b.WriteByte('\n')
		b.WriteString(resultHeadlineStyle.Render(result.Headline))
	}
	if result.Summary != "" {
		b.WriteByte('\n')
		b.WriteString(mutedStyle.Render(result.Summary))
	}
	return b.String()
}

func renderResultBody(result Result, runErr error, width int) string {
	var b strings.Builder
	if runErr != nil {
		writeResultHeading(&b, "ISSUE")
		b.WriteString(errorDetailStyle.Render(runErr.Error()))
		b.WriteString("\n")
	}
	if result.Allocation != nil {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(renderAllocationMap(*result.Allocation, width))
	}
	if len(result.Fields) > 0 {
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		writeResultHeading(&b, "OVERVIEW")
		b.WriteString(renderResultFields(result.Fields))
	}
	for _, section := range result.Sections {
		if len(section.Lines) == 0 {
			continue
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		writeResultHeading(&b, section.Title)
		for _, line := range section.Lines {
			b.WriteString(bodyStyle.Render(line))
			b.WriteByte('\n')
		}
	}
	if b.Len() == 0 {
		text := strings.TrimSuffix(result.Text, "\n")
		if text == "" && runErr == nil {
			text = "Done"
		}
		b.WriteString(bodyStyle.Render(text))
	}
	return strings.TrimSuffix(b.String(), "\n")
}

func writeResultHeading(b *strings.Builder, heading string) {
	b.WriteString(sectionStyle.Render(strings.ToUpper(heading)))
	b.WriteByte('\n')
}

func renderResultFields(fields []ResultField) string {
	labelWidth := 0
	for _, field := range fields {
		labelWidth = max(labelWidth, lipgloss.Width(field.Label))
	}
	var b strings.Builder
	for _, field := range fields {
		label := fieldLabelStyle.Render(fmt.Sprintf("%-*s", labelWidth, field.Label))
		b.WriteString(label)
		b.WriteString("  ")
		b.WriteString(resultToneStyle(field.Tone).Render(field.Value))
		b.WriteByte('\n')
	}
	return b.String()
}

func renderAllocationMap(allocation AllocationMap, width int) string {
	var b strings.Builder
	writeResultHeading(&b, "ADDRESS SPACE")
	railWidth := min(64, max(24, width-2))
	b.WriteString(renderRailEndpoints(allocation.Start, allocation.End, railWidth))
	b.WriteByte('\n')
	b.WriteString(renderAllocationRail(allocation, railWidth))
	b.WriteByte('\n')

	used := uint64(0)
	for _, segment := range allocation.Segments {
		used += segment.Size
	}
	available := allocation.Total - min(allocation.Total, used)
	b.WriteString(signalStyle.Render(fmt.Sprintf("%d allocated", used)))
	b.WriteString(mutedStyle.Render(fmt.Sprintf("  /  %d available", available)))
	b.WriteString("\n\n")

	writeResultHeading(&b, fmt.Sprintf("ALLOCATIONS  %d", len(allocation.Segments)))
	for i, segment := range allocation.Segments {
		index := segment.Index
		if index == 0 {
			index = i + 1
		}
		b.WriteString(allocationSegmentStyles[i%len(allocationSegmentStyles)].Render(fmt.Sprintf("%02d", index)))
		b.WriteString("  ")
		b.WriteString(resultValueStyle.Render(segment.Label))
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  %d addresses", segment.Size)))
		b.WriteByte('\n')
		b.WriteString("    ")
		b.WriteString(mutedStyle.Render(segment.Start + " → " + segment.End))
		b.WriteByte('\n')
	}
	if allocation.Available != nil {
		b.WriteByte('\n')
		writeResultHeading(&b, "AVAILABLE")
		b.WriteString(resultValueStyle.Render(allocation.Available.Start + " → " + allocation.Available.End))
		b.WriteString(mutedStyle.Render(fmt.Sprintf("  %d addresses", allocation.Available.Size)))
		b.WriteByte('\n')
	}
	return b.String()
}

func renderRailEndpoints(start, end string, width int) string {
	gap := max(1, width-lipgloss.Width(start)-lipgloss.Width(end))
	return mutedStyle.Render(start + strings.Repeat(" ", gap) + end)
}

func renderAllocationRail(allocation AllocationMap, width int) string {
	if allocation.Total == 0 {
		return unusedRailStyle.Render(strings.Repeat("░", width))
	}
	owners := make([]int, width)
	for i := range owners {
		owners[i] = -1
	}
	for segmentIndex, segment := range allocation.Segments {
		start := int(segment.Offset * uint64(width) / allocation.Total)
		end := int((segment.Offset + segment.Size) * uint64(width) / allocation.Total)
		if end <= start {
			end = start + 1
		}
		end = min(width, end)
		for cell := max(0, start); cell < end; cell++ {
			owners[cell] = segmentIndex
		}
	}

	var b strings.Builder
	for start := 0; start < width; {
		owner := owners[start]
		end := start + 1
		for end < width && owners[end] == owner {
			end++
		}
		cells := end - start
		if owner < 0 {
			b.WriteString(unusedRailStyle.Render(strings.Repeat("░", cells)))
		} else {
			style := allocationSegmentStyles[owner%len(allocationSegmentStyles)]
			b.WriteString(style.Render(strings.Repeat("█", cells)))
		}
		start = end
	}
	return b.String()
}

func resultToneStyle(tone ResultTone) lipgloss.Style {
	switch tone {
	case ToneSuccess:
		return successStyle
	case ToneWarning:
		return warningStyle
	case ToneDanger:
		return errorStyle
	default:
		return resultValueStyle
	}
}

func resultClipboardText(result Result, runErr error) string {
	var b strings.Builder
	if runErr != nil {
		b.WriteString("Error: ")
		b.WriteString(runErr.Error())
		if result.Text != "" {
			b.WriteString("\n\n")
		}
	}
	if result.Text != "" {
		b.WriteString(strings.TrimSuffix(result.Text, "\n"))
	} else if runErr == nil {
		if result.Headline != "" {
			b.WriteString(result.Headline)
		}
		if result.Summary != "" {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			b.WriteString(result.Summary)
		}
		for _, field := range result.Fields {
			if b.Len() > 0 {
				b.WriteByte('\n')
			}
			fmt.Fprintf(&b, "%s: %s", field.Label, field.Value)
		}
		if b.Len() == 0 {
			b.WriteString("Done")
		}
	}
	return b.String()
}
