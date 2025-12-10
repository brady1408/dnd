package styles

import "github.com/charmbracelet/lipgloss"

// Colors
var (
	PrimaryColor    = lipgloss.Color("#7C3AED") // Purple
	SecondaryColor  = lipgloss.Color("#EC4899") // Pink
	SuccessColor    = lipgloss.Color("#10B981") // Green
	WarningColor    = lipgloss.Color("#F59E0B") // Amber
	ErrorColor      = lipgloss.Color("#EF4444") // Red
	MutedColor      = lipgloss.Color("#6B7280") // Gray
	BackgroundColor = lipgloss.Color("#1F2937") // Dark gray
	ForegroundColor = lipgloss.Color("#F9FAFB") // Light gray
	HighlightColor  = lipgloss.Color("#A78BFA") // Light purple
)

// Styles holds all lipgloss styles for the application, bound to a specific renderer
type Styles struct {
	Muted         lipgloss.Style
	Base          lipgloss.Style
	Title         lipgloss.Style
	Subtitle      lipgloss.Style
	Header        lipgloss.Style
	Box           lipgloss.Style
	HighlightBox  lipgloss.Style
	Selected      lipgloss.Style
	Unselected    lipgloss.Style
	Cursor        lipgloss.Style
	InputField    lipgloss.Style
	FocusedInput  lipgloss.Style
	Button        lipgloss.Style
	FocusedButton lipgloss.Style
	Help          lipgloss.Style
	ErrorText     lipgloss.Style
	SuccessText   lipgloss.Style
	WarningText   lipgloss.Style
	StatValue     lipgloss.Style
	StatMod       lipgloss.Style
	StatLabel     lipgloss.Style
	HPCurrent     lipgloss.Style
	HPMax         lipgloss.Style
	HPLow         lipgloss.Style
	HPCritical    lipgloss.Style
	Proficient    lipgloss.Style
	NotProficient lipgloss.Style
	Logo          lipgloss.Style
}

// NewStyles creates a new Styles instance bound to the given renderer
func NewStyles(r *lipgloss.Renderer) *Styles {
	return &Styles{
		Muted: r.NewStyle().Foreground(MutedColor),

		Base: r.NewStyle().Foreground(ForegroundColor),

		Title: r.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			MarginBottom(1),

		Subtitle: r.NewStyle().
			Foreground(MutedColor).
			Italic(true),

		Header: r.NewStyle().
			Bold(true).
			Foreground(SecondaryColor).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(MutedColor).
			MarginBottom(1).
			PaddingBottom(0),

		Box: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(MutedColor).
			Padding(1, 2),

		HighlightBox: r.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(PrimaryColor).
			Padding(1, 2),

		Selected: r.NewStyle().
			Bold(true).
			Foreground(HighlightColor).
			Background(lipgloss.Color("#374151")),

		Unselected: r.NewStyle().Foreground(ForegroundColor),

		Cursor: r.NewStyle().
			Foreground(PrimaryColor).
			Bold(true),

		InputField: r.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(MutedColor).
			Padding(0, 1),

		FocusedInput: r.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(PrimaryColor).
			Padding(0, 1),

		Button: r.NewStyle().
			Foreground(ForegroundColor).
			Background(MutedColor).
			Padding(0, 2).
			MarginRight(1),

		FocusedButton: r.NewStyle().
			Foreground(ForegroundColor).
			Background(PrimaryColor).
			Padding(0, 2).
			Bold(true).
			MarginRight(1),

		Help: r.NewStyle().
			Foreground(MutedColor).
			MarginTop(1),

		ErrorText: r.NewStyle().
			Foreground(ErrorColor).
			Bold(true),

		SuccessText: r.NewStyle().
			Foreground(SuccessColor).
			Bold(true),

		WarningText: r.NewStyle().Foreground(WarningColor),

		StatValue: r.NewStyle().
			Bold(true).
			Foreground(PrimaryColor).
			Width(3).
			Align(lipgloss.Center),

		StatMod: r.NewStyle().
			Foreground(SecondaryColor).
			Width(4).
			Align(lipgloss.Center),

		StatLabel: r.NewStyle().
			Foreground(MutedColor).
			Width(12),

		HPCurrent: r.NewStyle().
			Bold(true).
			Foreground(SuccessColor),

		HPMax: r.NewStyle().Foreground(MutedColor),

		HPLow: r.NewStyle().
			Bold(true).
			Foreground(WarningColor),

		HPCritical: r.NewStyle().
			Bold(true).
			Foreground(ErrorColor),

		Proficient: r.NewStyle().Foreground(SuccessColor),

		NotProficient: r.NewStyle().Foreground(MutedColor),

		Logo: r.NewStyle().
			Foreground(PrimaryColor).
			Bold(true),
	}
}

// LogoText is the ASCII art logo
const LogoText = `
 ____  _   _ ____    ____  _                      _
|  _ \| \ | |  _ \  / ___|| |__   __ _ _ __ __ _ | |_ ___ _ __
| | | |  \| | | | | \___ \| '_ \ / _' | '__/ _' || __/ _ \ '__|
| |_| | |\  | |_| |  ___) | | | | (_| | | | (_| || ||  __/ |
|____/|_| \_|____/  |____/|_| |_|\__,_|_|  \__,_| \__\___|_|
`

// LogoSmall is a smaller logo for limited space
const LogoSmall = `
╔═══════════════════════╗
║   D&D Character App   ║
╚═══════════════════════╝
`
