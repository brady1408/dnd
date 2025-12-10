package components

import (
	"strings"

	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/lipgloss"
)

// Panel represents a bordered panel with optional title
type Panel struct {
	title   string
	content string
	width   int
	height  int
	styles  *styles.Styles
	focused bool
}

// NewPanel creates a new panel
func NewPanel(title string, s *styles.Styles) *Panel {
	return &Panel{
		title:  title,
		styles: s,
		width:  40,
		height: 0, // 0 = auto height
	}
}

// SetTitle sets the panel title
func (p *Panel) SetTitle(title string) {
	p.title = title
}

// SetContent sets the panel content
func (p *Panel) SetContent(content string) {
	p.content = content
}

// SetWidth sets the panel width
func (p *Panel) SetWidth(w int) {
	p.width = w
}

// SetHeight sets the panel height (0 = auto)
func (p *Panel) SetHeight(h int) {
	p.height = h
}

// SetFocused sets whether the panel is focused
func (p *Panel) SetFocused(focused bool) {
	p.focused = focused
}

// View renders the panel
func (p *Panel) View() string {
	borderColor := styles.MutedColor
	if p.focused {
		borderColor = styles.PrimaryColor
	}

	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(0, 1).
		Width(p.width - 2) // -2 for borders

	if p.height > 0 {
		style = style.Height(p.height - 2) // -2 for borders
	}

	var content string
	if p.title != "" {
		titleStyle := p.styles.Header.
			UnsetBorderStyle().
			UnsetBorderBottom().
			UnsetMarginBottom().
			UnsetPaddingBottom()
		content = titleStyle.Render(p.title) + "\n" + p.content
	} else {
		content = p.content
	}

	return style.Render(content)
}

// HStack arranges items horizontally with optional spacing
func HStack(spacing int, items ...string) string {
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		insertSpacing(spacing, items)...,
	)
}

// VStack arranges items vertically with optional spacing
func VStack(spacing int, items ...string) string {
	if spacing == 0 {
		return lipgloss.JoinVertical(lipgloss.Left, items...)
	}

	var result []string
	spacer := strings.Repeat("\n", spacing)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, spacer)
		}
	}
	return strings.Join(result, "")
}

// insertSpacing adds spacing between items
func insertSpacing(spacing int, items []string) []string {
	if spacing == 0 || len(items) <= 1 {
		return items
	}

	spacer := strings.Repeat(" ", spacing)
	result := make([]string, 0, len(items)*2-1)
	for i, item := range items {
		result = append(result, item)
		if i < len(items)-1 {
			result = append(result, spacer)
		}
	}
	return result
}

// Columns distributes items across columns with specified widths
// If widths has fewer elements than items, remaining items get equal share
func Columns(totalWidth int, widths []int, items ...string) string {
	if len(items) == 0 {
		return ""
	}

	// Calculate widths for each column
	colWidths := make([]int, len(items))
	usedWidth := 0

	for i := range items {
		if i < len(widths) && widths[i] > 0 {
			colWidths[i] = widths[i]
			usedWidth += widths[i]
		}
	}

	// Distribute remaining width to columns without explicit width
	remaining := totalWidth - usedWidth
	flexCount := 0
	for i := range items {
		if i >= len(widths) || widths[i] == 0 {
			flexCount++
		}
	}

	if flexCount > 0 {
		flexWidth := max(10, remaining/flexCount)
		for i := range items {
			if i >= len(widths) || widths[i] == 0 {
				colWidths[i] = flexWidth
			}
		}
	}

	// Render each item with its width
	renderedItems := make([]string, len(items))
	for i, item := range items {
		renderedItems[i] = lipgloss.NewStyle().
			Width(colWidths[i]).
			Render(item)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, renderedItems...)
}

// Center centers content within a given width and height
func Center(content string, width, height int) string {
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

// Padded adds padding around content
func Padded(content string, top, right, bottom, left int) string {
	return lipgloss.NewStyle().
		PaddingTop(top).
		PaddingRight(right).
		PaddingBottom(bottom).
		PaddingLeft(left).
		Render(content)
}

// Bordered adds a border around content
func Bordered(content string, color lipgloss.Color) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(color).
		Padding(0, 1).
		Render(content)
}

// KeyValue renders a label-value pair
func KeyValue(label, value string, labelWidth int, s *styles.Styles) string {
	labelStyled := s.Muted.Width(labelWidth).Align(lipgloss.Right).Render(label + ":")
	return labelStyled + " " + value
}

// ProgressBar renders a simple progress bar
func ProgressBar(current, max, width int, filledColor, emptyColor lipgloss.Color) string {
	if max <= 0 {
		max = 1
	}
	if current < 0 {
		current = 0
	}
	if current > max {
		current = max
	}

	filled := (current * width) / max
	empty := width - filled

	filledStyle := lipgloss.NewStyle().Foreground(filledColor)
	emptyStyle := lipgloss.NewStyle().Foreground(emptyColor)

	return filledStyle.Render(strings.Repeat("█", filled)) +
		emptyStyle.Render(strings.Repeat("░", empty))
}

// CheckboxRow renders a row of checkboxes (like death saves or spell slots)
func CheckboxRow(checked, total int, checkedChar, uncheckedChar string) string {
	var result strings.Builder
	for i := 0; i < total; i++ {
		if i < checked {
			result.WriteString(checkedChar)
		} else {
			result.WriteString(uncheckedChar)
		}
	}
	return result.String()
}

// SlotTracker renders spell slot style tracker: [*][*][ ][ ]
func SlotTracker(used, total int) string {
	return CheckboxRow(total-used, total, "[●]", "[ ]")
}

// DeathSaves renders death save trackers
func DeathSaves(successes, failures int, s *styles.Styles) string {
	successRow := s.SuccessText.Render(CheckboxRow(successes, 3, "●", "○"))
	failureRow := s.ErrorText.Render(CheckboxRow(failures, 3, "●", "○"))
	return "Success: " + successRow + "\n" + "Failure: " + failureRow
}
