package components

import (
	"strings"

	"github.com/brady1408/dnd/internal/tui/styles"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TableColumn defines a column in the table
type TableColumn struct {
	Title     string
	Width     int  // Fixed width; if 0, uses flexible sizing
	MinWidth  int  // Minimum width for flexible columns
	Alignment lipgloss.Position
}

// TableRow represents a single row of data
type TableRow struct {
	ID    string      // Unique identifier (e.g., UUID)
	Cells []string    // Cell values matching column order
	Data  interface{} // Original data object for callbacks
}

// TableSelectMsg is sent when a row is selected (Enter pressed)
type TableSelectMsg struct {
	Row TableRow
}

// TableEditMsg is sent when edit is requested (e pressed)
type TableEditMsg struct {
	Row TableRow
}

// TableDeleteMsg is sent when delete is requested (d pressed)
type TableDeleteMsg struct {
	Row TableRow
}

// TableModel is a reusable scrollable table component
type TableModel struct {
	columns     []TableColumn
	rows        []TableRow
	cursor      int // Currently highlighted row
	viewport    int // First visible row index
	visibleRows int // Number of rows visible at once
	focused     bool
	width       int // Total available width

	styles *styles.Styles

	// Computed styles (created from styles)
	headerStyle   lipgloss.Style
	rowStyle      lipgloss.Style
	selectedStyle lipgloss.Style
	cellStyle     lipgloss.Style

	// Empty state
	emptyMessage string
}

// NewTable creates a new table component
func NewTable(columns []TableColumn, s *styles.Styles) *TableModel {
	t := &TableModel{
		columns:      columns,
		rows:         []TableRow{},
		visibleRows:  10,
		focused:      false,
		width:        80,
		styles:       s,
		emptyMessage: "No items",
	}
	t.initStyles()
	return t
}

func (t *TableModel) initStyles() {
	// Header style
	t.headerStyle = t.styles.Header.
		UnsetBorderStyle().
		UnsetBorderBottom().
		UnsetMarginBottom().
		UnsetPaddingBottom().
		Bold(true)

	// Normal row style
	t.rowStyle = t.styles.Base

	// Selected row style
	t.selectedStyle = t.styles.Selected

	// Cell style (for padding)
	t.cellStyle = lipgloss.NewStyle().PaddingRight(1)
}

// SetRows replaces all rows in the table
func (t *TableModel) SetRows(rows []TableRow) {
	t.rows = rows
	// Reset cursor if out of bounds
	if t.cursor >= len(rows) {
		t.cursor = max(0, len(rows)-1)
	}
	// Adjust viewport
	t.adjustViewport()
}

// AddRow appends a row to the table
func (t *TableModel) AddRow(row TableRow) {
	t.rows = append(t.rows, row)
}

// RemoveRow removes a row by ID
func (t *TableModel) RemoveRow(id string) {
	for i, row := range t.rows {
		if row.ID == id {
			t.rows = append(t.rows[:i], t.rows[i+1:]...)
			if t.cursor >= len(t.rows) && t.cursor > 0 {
				t.cursor--
			}
			t.adjustViewport()
			return
		}
	}
}

// UpdateRow updates a row by ID
func (t *TableModel) UpdateRow(id string, newRow TableRow) {
	for i, row := range t.rows {
		if row.ID == id {
			t.rows[i] = newRow
			return
		}
	}
}

// GetSelectedRow returns the currently selected row, or nil if none
func (t *TableModel) GetSelectedRow() *TableRow {
	if len(t.rows) == 0 || t.cursor < 0 || t.cursor >= len(t.rows) {
		return nil
	}
	return &t.rows[t.cursor]
}

// SetVisibleRows sets how many rows are visible at once
func (t *TableModel) SetVisibleRows(n int) {
	t.visibleRows = n
	t.adjustViewport()
}

// SetWidth sets the total available width for the table
func (t *TableModel) SetWidth(w int) {
	t.width = w
}

// SetFocused sets whether the table has focus
func (t *TableModel) SetFocused(focused bool) {
	t.focused = focused
}

// IsFocused returns whether the table has focus
func (t *TableModel) IsFocused() bool {
	return t.focused
}

// SetEmptyMessage sets the message shown when table has no rows
func (t *TableModel) SetEmptyMessage(msg string) {
	t.emptyMessage = msg
}

// RowCount returns the number of rows
func (t *TableModel) RowCount() int {
	return len(t.rows)
}

// adjustViewport ensures cursor is visible
func (t *TableModel) adjustViewport() {
	if t.cursor < t.viewport {
		t.viewport = t.cursor
	}
	if t.cursor >= t.viewport+t.visibleRows {
		t.viewport = t.cursor - t.visibleRows + 1
	}
	// Clamp viewport
	maxViewport := max(0, len(t.rows)-t.visibleRows)
	if t.viewport > maxViewport {
		t.viewport = maxViewport
	}
	if t.viewport < 0 {
		t.viewport = 0
	}
}

// Init implements tea.Model
func (t *TableModel) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (t *TableModel) Update(msg tea.Msg) (*TableModel, tea.Cmd) {
	if !t.focused {
		return t, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if t.cursor > 0 {
				t.cursor--
				t.adjustViewport()
			}
		case "down", "j":
			if t.cursor < len(t.rows)-1 {
				t.cursor++
				t.adjustViewport()
			}
		case "pgup":
			t.cursor = max(0, t.cursor-t.visibleRows)
			t.adjustViewport()
		case "pgdown":
			t.cursor = min(len(t.rows)-1, t.cursor+t.visibleRows)
			t.adjustViewport()
		case "home", "g":
			t.cursor = 0
			t.adjustViewport()
		case "end", "G":
			t.cursor = max(0, len(t.rows)-1)
			t.adjustViewport()
		case "enter":
			if row := t.GetSelectedRow(); row != nil {
				return t, func() tea.Msg { return TableSelectMsg{Row: *row} }
			}
		case "e":
			if row := t.GetSelectedRow(); row != nil {
				return t, func() tea.Msg { return TableEditMsg{Row: *row} }
			}
		case "d", "delete":
			if row := t.GetSelectedRow(); row != nil {
				return t, func() tea.Msg { return TableDeleteMsg{Row: *row} }
			}
		}
	}

	return t, nil
}

// View implements tea.Model
func (t *TableModel) View() string {
	var b strings.Builder

	// Calculate column widths
	widths := t.calculateWidths()

	// Render header
	b.WriteString(t.renderHeader(widths))
	b.WriteString("\n")

	// Render separator
	b.WriteString(t.renderSeparator(widths))
	b.WriteString("\n")

	// Render rows or empty state
	if len(t.rows) == 0 {
		b.WriteString(t.styles.Muted.Render(t.emptyMessage))
		b.WriteString("\n")
	} else {
		// Render visible rows
		endRow := min(t.viewport+t.visibleRows, len(t.rows))
		for i := t.viewport; i < endRow; i++ {
			isSelected := t.focused && i == t.cursor
			b.WriteString(t.renderRow(t.rows[i], widths, isSelected))
			b.WriteString("\n")
		}

		// Pad with empty lines if needed
		renderedRows := endRow - t.viewport
		for i := renderedRows; i < t.visibleRows; i++ {
			b.WriteString("\n")
		}
	}

	// Render scroll indicator if needed
	if len(t.rows) > t.visibleRows {
		b.WriteString(t.renderScrollIndicator())
	}

	return b.String()
}

// calculateWidths calculates the width for each column
func (t *TableModel) calculateWidths() []int {
	widths := make([]int, len(t.columns))
	fixedWidth := 0
	flexCount := 0

	// First pass: calculate fixed widths and count flexible columns
	for i, col := range t.columns {
		if col.Width > 0 {
			widths[i] = col.Width
			fixedWidth += col.Width + 1 // +1 for padding
		} else {
			flexCount++
		}
	}

	// Second pass: distribute remaining width to flexible columns
	if flexCount > 0 {
		remaining := t.width - fixedWidth - 2 // -2 for borders
		flexWidth := max(10, remaining/flexCount)
		for i, col := range t.columns {
			if col.Width == 0 {
				w := flexWidth
				if col.MinWidth > 0 && w < col.MinWidth {
					w = col.MinWidth
				}
				widths[i] = w
			}
		}
	}

	return widths
}

// renderHeader renders the table header row
func (t *TableModel) renderHeader(widths []int) string {
	var cells []string
	for i, col := range t.columns {
		cell := truncateOrPad(col.Title, widths[i])
		cells = append(cells, t.headerStyle.Render(cell))
	}
	return strings.Join(cells, " ")
}

// renderSeparator renders a separator line
func (t *TableModel) renderSeparator(widths []int) string {
	var parts []string
	for _, w := range widths {
		parts = append(parts, strings.Repeat("─", w))
	}
	return t.styles.Muted.Render(strings.Join(parts, "─"))
}

// renderRow renders a single data row
func (t *TableModel) renderRow(row TableRow, widths []int, selected bool) string {
	var cells []string
	for i, cell := range row.Cells {
		if i >= len(widths) {
			break
		}
		text := truncateOrPad(cell, widths[i])
		cells = append(cells, text)
	}

	// Pad with empty cells if row has fewer cells than columns
	for i := len(row.Cells); i < len(widths); i++ {
		cells = append(cells, strings.Repeat(" ", widths[i]))
	}

	line := strings.Join(cells, " ")

	if selected {
		return t.selectedStyle.Render(line)
	}
	return t.rowStyle.Render(line)
}

// renderScrollIndicator shows scroll position
func (t *TableModel) renderScrollIndicator() string {
	total := len(t.rows)
	current := t.viewport + 1
	end := min(t.viewport+t.visibleRows, total)

	return t.styles.Muted.Render(
		strings.Repeat(" ", 2) +
			"↑↓ " +
			string(rune('0'+current/10)) + string(rune('0'+current%10)) +
			"-" +
			string(rune('0'+end/10)) + string(rune('0'+end%10)) +
			" of " +
			string(rune('0'+total/10)) + string(rune('0'+total%10)),
	)
}

// truncateOrPad ensures a string is exactly the given width
func truncateOrPad(s string, width int) string {
	if width <= 0 {
		return ""
	}

	// Count actual display width (simplified - assumes ASCII)
	runes := []rune(s)
	if len(runes) > width {
		if width > 3 {
			return string(runes[:width-3]) + "..."
		}
		return string(runes[:width])
	}

	// Pad with spaces
	return s + strings.Repeat(" ", width-len(runes))
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
