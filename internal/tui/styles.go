package tui

import "github.com/charmbracelet/lipgloss"

// Theme represents a color theme for the TUI
type Theme struct {
	Surface  lipgloss.Color
	Overlay  lipgloss.Color
	Text     lipgloss.Color
	Subtext  lipgloss.Color
	Lavender lipgloss.Color
	Blue     lipgloss.Color
	Green    lipgloss.Color
	Peach    lipgloss.Color
	Red      lipgloss.Color
	Yellow   lipgloss.Color
}

// DefaultTheme returns the default Catppuccin Mocha inspired theme
func DefaultTheme() Theme {
	return Theme{
		Surface:  lipgloss.Color("#313244"),
		Overlay:  lipgloss.Color("#45475a"),
		Text:     lipgloss.Color("#cdd6f4"),
		Subtext:  lipgloss.Color("#a6adc8"),
		Lavender: lipgloss.Color("#b4befe"),
		Blue:     lipgloss.Color("#89b4fa"),
		Green:    lipgloss.Color("#a6e3a1"),
		Peach:    lipgloss.Color("#fab387"),
		Red:      lipgloss.Color("#f38ba8"),
		Yellow:   lipgloss.Color("#f9e2af"),
	}
}

// GruvboxTheme returns a Gruvbox dark theme
func GruvboxTheme() Theme {
	return Theme{
		Surface:  lipgloss.Color("#282828"),
		Overlay:  lipgloss.Color("#3c3836"),
		Text:     lipgloss.Color("#ebdbb2"),
		Subtext:  lipgloss.Color("#a89984"),
		Lavender: lipgloss.Color("#d3869b"),
		Blue:     lipgloss.Color("#83a598"),
		Green:    lipgloss.Color("#b8bb26"),
		Peach:    lipgloss.Color("#fe8019"),
		Red:      lipgloss.Color("#fb4934"),
		Yellow:   lipgloss.Color("#fabd2f"),
	}
}

// TokyoNightTheme returns a Tokyo Night theme
func TokyoNightTheme() Theme {
	return Theme{
		Surface:  lipgloss.Color("#1a1b26"),
		Overlay:  lipgloss.Color("#24283b"),
		Text:     lipgloss.Color("#c0caf5"),
		Subtext:  lipgloss.Color("#9aa5ce"),
		Lavender: lipgloss.Color("#bb9af7"),
		Blue:     lipgloss.Color("#7aa2f7"),
		Green:    lipgloss.Color("#9ece6a"),
		Peach:    lipgloss.Color("#ff9e64"),
		Red:      lipgloss.Color("#f7768e"),
		Yellow:   lipgloss.Color("#e0af68"),
	}
}

// Styles holds all lipgloss styles for the TUI
type Styles struct {
	Title                lipgloss.Style
	Subtitle             lipgloss.Style
	Subtext              lipgloss.Style
	Project              lipgloss.Style
	SelectedProject      lipgloss.Style
	ProjectName          lipgloss.Style
	SelectedProjectName  lipgloss.Style
	Path                 lipgloss.Style
	DockerBadge          lipgloss.Style
	LocalBadge           lipgloss.Style
	Task                 lipgloss.Style
	Help                 lipgloss.Style
	HelpKey              lipgloss.Style
	HelpDesc             lipgloss.Style
	HelpSeparator        lipgloss.Style
	Error                lipgloss.Style
	Success              lipgloss.Style
	Container            lipgloss.Style
	FormLabel            lipgloss.Style
	Form                 lipgloss.Style
	DeleteBox            lipgloss.Style
	SearchInput          lipgloss.Style
	FilterBadge          lipgloss.Style
	StatusBar            lipgloss.Style
	Count                lipgloss.Style
}

// NewStyles creates a new Styles instance from a theme
func NewStyles(theme Theme) Styles {
	return Styles{
		Title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Lavender),

		Subtitle: lipgloss.NewStyle().
			Foreground(theme.Subtext),

		Subtext: lipgloss.NewStyle().
			Foreground(theme.Subtext),

		Project: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(0).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(theme.Overlay),

		SelectedProject: lipgloss.NewStyle().
			PaddingLeft(1).
			PaddingRight(1).
			MarginBottom(0).
			BorderStyle(lipgloss.ThickBorder()).
			BorderLeft(true).
			BorderForeground(theme.Lavender),

		ProjectName: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Text),

		SelectedProjectName: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Lavender),

		Path: lipgloss.NewStyle().
			Foreground(theme.Subtext),

		DockerBadge: lipgloss.NewStyle().
			Foreground(theme.Blue).
			Bold(true),

		LocalBadge: lipgloss.NewStyle().
			Foreground(theme.Green).
			Bold(true),

		Task: lipgloss.NewStyle().
			Foreground(theme.Peach),

		Help: lipgloss.NewStyle().
			Foreground(theme.Subtext).
			Background(theme.Surface).
			Padding(0, 1),

		HelpKey: lipgloss.NewStyle().
			Foreground(theme.Lavender).
			Bold(true),

		HelpDesc: lipgloss.NewStyle().
			Foreground(theme.Subtext),

		HelpSeparator: lipgloss.NewStyle().
			Foreground(theme.Overlay).
			SetString("│"),

		Error: lipgloss.NewStyle().
			Foreground(theme.Red).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(theme.Green).
			Bold(true),

		Container: lipgloss.NewStyle().
			PaddingTop(1).
			PaddingBottom(1).
			PaddingLeft(2).
			PaddingRight(2),

		FormLabel: lipgloss.NewStyle().
			Foreground(theme.Lavender).
			Bold(true).
			Width(12),

		Form: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Lavender).
			Padding(1, 2).
			MarginTop(1),

		DeleteBox: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Red).
			Padding(1, 2).
			MarginTop(1),

		SearchInput: lipgloss.NewStyle().
			Foreground(theme.Text).
			Background(theme.Surface).
			Padding(0, 1).
			Bold(true),

		FilterBadge: lipgloss.NewStyle().
			Foreground(theme.Yellow).
			Background(theme.Surface).
			Padding(0, 1).
			Bold(true),

		StatusBar: lipgloss.NewStyle().
			Foreground(theme.Subtext).
			Background(theme.Surface).
			Padding(0, 1),

		Count: lipgloss.NewStyle().
			Foreground(theme.Blue).
			Bold(true),
	}
}

