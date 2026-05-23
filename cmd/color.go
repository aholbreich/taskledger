package cmd

import (
	"strings"

	"github.com/spf13/cobra"

	internalcolor "github.com/aholbreich/tl/internal/color"
)

func commandColorEnabled(cmd *cobra.Command) bool {
	mode := internalcolor.ModeAuto
	if flag := cmd.Root().Flag("color"); flag != nil {
		mode = flag.Value.String()
	}
	return internalcolor.Enabled(mode, cmd.OutOrStdout())
}

func colorStatus(enabled bool, status string) string {
	if !enabled {
		return status
	}
	return internalcolor.Status(status)
}

func colorPriority(enabled bool, priority string) string {
	if !enabled {
		return priority
	}
	return internalcolor.Priority(priority)
}

func colorListPriority(enabled bool, priority string) string {
	if !enabled {
		return priority
	}
	return internalcolor.ListPriority(priority)
}

func colorID(enabled bool, id string) string {
	if !enabled {
		return id
	}
	return internalcolor.ID(id)
}

func colorFieldLabel(enabled bool, label string) string {
	if !enabled {
		return label
	}
	return internalcolor.FieldLabel(label)
}

func colorFieldValue(enabled bool, value string) string {
	if !enabled {
		return value
	}
	return internalcolor.FieldValue(value)
}

func colorMarkdownHeadings(enabled bool, body string) string {
	if !enabled {
		return body
	}

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if isMarkdownHeading(line) {
			lines[i] = internalcolor.MarkdownHeading(line)
		}
	}
	return strings.Join(lines, "\n")
}

func isMarkdownHeading(line string) bool {
	count := 0
	for count < len(line) && count < 6 && line[count] == '#' {
		count++
	}
	return count > 0 && count < len(line) && line[count] == ' '
}

func colorClosedListLine(enabled bool, line string) string {
	if !enabled {
		return line
	}
	return internalcolor.ClosedListLine(line)
}

func colorDimCode() string {
	return internalcolor.Dim
}

func colorResetCode() string {
	return internalcolor.Reset
}
