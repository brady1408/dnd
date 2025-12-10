package styles

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	PrimaryColor    = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor  = lipgloss.Color("#EC4899") // Pink
	SuccessColor    = lipgloss.Color("#10B981") // Green
	WarningColor    = lipgloss.Color("#F59E0B") // Amber
	ErrorColor      = lipgloss.Color("#EF4444") // Red
	MutedColor      = lipgloss.Color("#6B7280") // Gray
	BackgroundColor = lipgloss.Color("#1F2937") // Dark gray
	ForegroundColor = lipgloss.Color("#F9FAFB") // Light gray
	HighlightColor  = lipgloss.Color("#A78BFA") // Light purple

	// Muted text style
	Muted = lipgloss.NewStyle().Foreground(MutedColor)

	// Base styles
	Base = lipgloss.NewStyle().
		Foreground(ForegroundColor)

	// Title
	Title = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		MarginBottom(1)

	// Subtitle
	Subtitle = lipgloss.NewStyle().
		Foreground(MutedColor).
		Italic(true)

	// Header for sections
	Header = lipgloss.NewStyle().
		Bold(true).
		Foreground(SecondaryColor).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(MutedColor).
		MarginBottom(1).
		PaddingBottom(0)

	// Box for content sections
	Box = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(MutedColor).
		Padding(1, 2)

	// Highlight box
	HighlightBox = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(PrimaryColor).
		Padding(1, 2)

	// Selected item
	Selected = lipgloss.NewStyle().
		Bold(true).
		Foreground(HighlightColor).
		Background(lipgloss.Color("#374151"))

	// Unselected item
	Unselected = lipgloss.NewStyle().
		Foreground(ForegroundColor)

	// Cursor
	Cursor = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)

	// Input field
	InputField = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(MutedColor).
		Padding(0, 1)

	// Focused input
	FocusedInput = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(PrimaryColor).
		Padding(0, 1)

	// Button
	Button = lipgloss.NewStyle().
		Foreground(ForegroundColor).
		Background(MutedColor).
		Padding(0, 2).
		MarginRight(1)

	// Focused button
	FocusedButton = lipgloss.NewStyle().
		Foreground(ForegroundColor).
		Background(PrimaryColor).
		Padding(0, 2).
		Bold(true).
		MarginRight(1)

	// Help text
	Help = lipgloss.NewStyle().
		Foreground(MutedColor).
		MarginTop(1)

	// Error text
	ErrorText = lipgloss.NewStyle().
		Foreground(ErrorColor).
		Bold(true)

	// Success text
	SuccessText = lipgloss.NewStyle().
		Foreground(SuccessColor).
		Bold(true)

	// Warning text
	WarningText = lipgloss.NewStyle().
		Foreground(WarningColor)

	// Stat value (ability score)
	StatValue = lipgloss.NewStyle().
		Bold(true).
		Foreground(PrimaryColor).
		Width(3).
		Align(lipgloss.Center)

	// Stat modifier
	StatMod = lipgloss.NewStyle().
		Foreground(SecondaryColor).
		Width(4).
		Align(lipgloss.Center)

	// Stat label
	StatLabel = lipgloss.NewStyle().
		Foreground(MutedColor).
		Width(12)

	// HP Current
	HPCurrent = lipgloss.NewStyle().
		Bold(true).
		Foreground(SuccessColor)

	// HP Max
	HPMax = lipgloss.NewStyle().
		Foreground(MutedColor)

	// HP Low (< 50%)
	HPLow = lipgloss.NewStyle().
		Bold(true).
		Foreground(WarningColor)

	// HP Critical (< 25%)
	HPCritical = lipgloss.NewStyle().
		Bold(true).
		Foreground(ErrorColor)

	// Proficient skill
	Proficient = lipgloss.NewStyle().
		Foreground(SuccessColor)

	// Non-proficient skill
	NotProficient = lipgloss.NewStyle().
		Foreground(MutedColor)

	// ASCII art header style
	Logo = lipgloss.NewStyle().
		Foreground(PrimaryColor).
		Bold(true)
)

// LogoText is the ASCII art logo
const LogoText = `
 ____  _   _ ____    ____  _                      _
|  _ \| \ | |  _ \  / ___|| |__   __ _ _ __ __ _ | |_ ___ _ __
| | | |  \| | | | | \___ \| '_ \ / _' | '__/ _' || __/ _ \ '__|
| |_| | |\  | |_| |  ___) | | | | (_| | | | (_| || ||  __/ |
|____/|_| \_|____/  |____/|_| |_|\__,_|_|  \__,_| \__\___|_|
`

// Smaller logo for limited space
const LogoSmall = `
╔═══════════════════════╗
║   D&D Character App   ║
╚═══════════════════════╝
`
