package components

import (
	"strings"

	"github.com/brady1408/dnd/internal/tui/styles"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// FieldType represents the type of form field
type FieldType int

const (
	FieldText FieldType = iota
	FieldNumber
	FieldSelect
	FieldCheckbox
)

// FormField represents a single input field in a form
type FormField struct {
	Key         string    // Unique key for this field
	Label       string    // Display label
	Type        FieldType // Field type
	Value       string    // Current value
	Options     []string  // Options for select type
	Required    bool      // Whether field is required
	Placeholder string    // Placeholder text
	Width       int       // Input width (0 = auto)
}

// ModalSaveMsg is sent when the form is saved
type ModalSaveMsg struct {
	Values map[string]string
}

// ModalCancelMsg is sent when the modal is cancelled
type ModalCancelMsg struct{}

// ModalModel is a modal dialog for forms
type ModalModel struct {
	title       string
	fields      []FormField
	cursor      int  // Current field index
	focused     bool
	visible     bool

	// Dimensions
	width  int
	height int

	// Text inputs for each field
	inputs []textinput.Model

	// For select fields
	selectCursors []int // Current selection for each select field

	// For checkbox fields
	checkboxValues []bool

	styles *styles.Styles

	// Computed styles
	overlayStyle lipgloss.Style
	boxStyle     lipgloss.Style
	titleStyle   lipgloss.Style
	labelStyle   lipgloss.Style
	helpStyle    lipgloss.Style
}

// NewModal creates a new modal form
func NewModal(title string, fields []FormField, s *styles.Styles) *ModalModel {
	m := &ModalModel{
		title:          title,
		fields:         fields,
		cursor:         0,
		focused:        true,
		visible:        true,
		width:          60,
		height:         20,
		styles:         s,
		inputs:         make([]textinput.Model, len(fields)),
		selectCursors:  make([]int, len(fields)),
		checkboxValues: make([]bool, len(fields)),
	}

	// Initialize inputs for each field
	for i, field := range fields {
		ti := textinput.New()
		ti.Placeholder = field.Placeholder
		if field.Width > 0 {
			ti.Width = field.Width
		} else {
			ti.Width = 40
		}
		ti.SetValue(field.Value)

		if i == 0 {
			ti.Focus()
		}

		m.inputs[i] = ti

		// Initialize select cursor to current value
		if field.Type == FieldSelect {
			for j, opt := range field.Options {
				if opt == field.Value {
					m.selectCursors[i] = j
					break
				}
			}
		}

		// Initialize checkbox
		if field.Type == FieldCheckbox {
			m.checkboxValues[i] = field.Value == "true" || field.Value == "1" || field.Value == "yes"
		}
	}

	m.initStyles()
	return m
}

func (m *ModalModel) initStyles() {
	m.overlayStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("#000000"))

	m.boxStyle = m.styles.HighlightBox.
		Width(m.width).
		Padding(1, 2)

	m.titleStyle = m.styles.Title

	m.labelStyle = m.styles.Base.
		Width(15).
		Align(lipgloss.Right).
		MarginRight(1)

	m.helpStyle = m.styles.Help
}

// SetSize sets the terminal size for centering
func (m *ModalModel) SetSize(width, height int) {
	// Modal takes up a portion of the screen
	m.width = min(60, width-10)
	m.height = min(len(m.fields)*3+8, height-6)
	m.boxStyle = m.boxStyle.Width(m.width)
}

// Show makes the modal visible
func (m *ModalModel) Show() {
	m.visible = true
	m.focused = true
	m.cursor = 0
	if len(m.inputs) > 0 {
		m.inputs[0].Focus()
	}
}

// Hide hides the modal
func (m *ModalModel) Hide() {
	m.visible = false
	m.focused = false
}

// IsVisible returns whether the modal is visible
func (m *ModalModel) IsVisible() bool {
	return m.visible
}

// GetValues returns all field values as a map
func (m *ModalModel) GetValues() map[string]string {
	values := make(map[string]string)
	for i, field := range m.fields {
		switch field.Type {
		case FieldText, FieldNumber:
			values[field.Key] = m.inputs[i].Value()
		case FieldSelect:
			if m.selectCursors[i] < len(field.Options) {
				values[field.Key] = field.Options[m.selectCursors[i]]
			}
		case FieldCheckbox:
			if m.checkboxValues[i] {
				values[field.Key] = "true"
			} else {
				values[field.Key] = "false"
			}
		}
	}
	return values
}

// SetValue sets a field's value by key
func (m *ModalModel) SetValue(key, value string) {
	for i, field := range m.fields {
		if field.Key == key {
			m.inputs[i].SetValue(value)
			if field.Type == FieldSelect {
				for j, opt := range field.Options {
					if opt == value {
						m.selectCursors[i] = j
						break
					}
				}
			}
			if field.Type == FieldCheckbox {
				m.checkboxValues[i] = value == "true" || value == "1" || value == "yes"
			}
			return
		}
	}
}

// Init implements tea.Model
func (m *ModalModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update implements tea.Model
func (m *ModalModel) Update(msg tea.Msg) (*ModalModel, tea.Cmd) {
	if !m.visible || !m.focused {
		return m, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.Hide()
			return m, func() tea.Msg { return ModalCancelMsg{} }

		case "ctrl+s":
			// Validate required fields
			for i, field := range m.fields {
				if field.Required && m.inputs[i].Value() == "" {
					// Could add error display here
					return m, nil
				}
			}
			values := m.GetValues()
			m.Hide()
			return m, func() tea.Msg { return ModalSaveMsg{Values: values} }

		case "tab", "down":
			// Move to next field
			m.inputs[m.cursor].Blur()
			m.cursor = (m.cursor + 1) % len(m.fields)
			if m.fields[m.cursor].Type == FieldText || m.fields[m.cursor].Type == FieldNumber {
				m.inputs[m.cursor].Focus()
				cmds = append(cmds, textinput.Blink)
			}
			return m, tea.Batch(cmds...)

		case "shift+tab", "up":
			// Move to previous field
			m.inputs[m.cursor].Blur()
			m.cursor = (m.cursor - 1 + len(m.fields)) % len(m.fields)
			if m.fields[m.cursor].Type == FieldText || m.fields[m.cursor].Type == FieldNumber {
				m.inputs[m.cursor].Focus()
				cmds = append(cmds, textinput.Blink)
			}
			return m, tea.Batch(cmds...)

		case "left":
			// For select fields, move selection left
			if m.fields[m.cursor].Type == FieldSelect {
				opts := m.fields[m.cursor].Options
				if len(opts) > 0 {
					m.selectCursors[m.cursor] = (m.selectCursors[m.cursor] - 1 + len(opts)) % len(opts)
				}
				return m, nil
			}

		case "right":
			// For select fields, move selection right
			if m.fields[m.cursor].Type == FieldSelect {
				opts := m.fields[m.cursor].Options
				if len(opts) > 0 {
					m.selectCursors[m.cursor] = (m.selectCursors[m.cursor] + 1) % len(opts)
				}
				return m, nil
			}

		case " ", "enter":
			// For checkbox, toggle
			if m.fields[m.cursor].Type == FieldCheckbox {
				m.checkboxValues[m.cursor] = !m.checkboxValues[m.cursor]
				return m, nil
			}
			// For select, also allow enter to cycle (or could open dropdown)
			if m.fields[m.cursor].Type == FieldSelect {
				opts := m.fields[m.cursor].Options
				if len(opts) > 0 {
					m.selectCursors[m.cursor] = (m.selectCursors[m.cursor] + 1) % len(opts)
				}
				return m, nil
			}
		}
	}

	// Update the current text input
	if m.cursor < len(m.inputs) {
		field := m.fields[m.cursor]
		if field.Type == FieldText || field.Type == FieldNumber {
			var cmd tea.Cmd
			m.inputs[m.cursor], cmd = m.inputs[m.cursor].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

// View implements tea.Model
func (m *ModalModel) View() string {
	if !m.visible {
		return ""
	}

	var b strings.Builder

	// Title
	b.WriteString(m.titleStyle.Render(m.title))
	b.WriteString("\n\n")

	// Fields
	for i, field := range m.fields {
		isFocused := m.focused && i == m.cursor

		// Label
		label := field.Label
		if field.Required {
			label += " *"
		}
		b.WriteString(m.labelStyle.Render(label))

		// Input based on type
		switch field.Type {
		case FieldText, FieldNumber:
			if isFocused {
				b.WriteString(m.styles.FocusedInput.Render(m.inputs[i].View()))
			} else {
				b.WriteString(m.styles.InputField.Render(m.inputs[i].View()))
			}

		case FieldSelect:
			opts := field.Options
			selected := m.selectCursors[i]
			var optView string
			if len(opts) > 0 && selected < len(opts) {
				optView = "◀ " + opts[selected] + " ▶"
			} else {
				optView = "(none)"
			}
			if isFocused {
				b.WriteString(m.styles.FocusedInput.Render(optView))
			} else {
				b.WriteString(m.styles.InputField.Render(optView))
			}

		case FieldCheckbox:
			var checkView string
			if m.checkboxValues[i] {
				checkView = "[✓]"
			} else {
				checkView = "[ ]"
			}
			if isFocused {
				b.WriteString(m.styles.FocusedButton.Render(checkView))
			} else {
				b.WriteString(m.styles.Button.Render(checkView))
			}
		}

		b.WriteString("\n\n")
	}

	// Help text
	b.WriteString("\n")
	b.WriteString(m.helpStyle.Render("tab: next field • ctrl+s: save • esc: cancel"))

	// Wrap in box
	return m.boxStyle.Render(b.String())
}

// ViewWithOverlay renders the modal centered over a background
func (m *ModalModel) ViewWithOverlay(background string, termWidth, termHeight int) string {
	if !m.visible {
		return background
	}

	modalContent := m.View()

	// Center the modal
	return lipgloss.Place(
		termWidth,
		termHeight,
		lipgloss.Center,
		lipgloss.Center,
		modalContent,
		lipgloss.WithWhitespaceChars(" "),
		lipgloss.WithWhitespaceForeground(lipgloss.Color("#333333")),
	)
}
