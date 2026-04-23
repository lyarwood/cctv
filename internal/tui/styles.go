package tui

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle         lipgloss.Style
	statusBarStyle     lipgloss.Style
	runningStyle       lipgloss.Style
	selectedStyle      lipgloss.Style
	dimStyle           lipgloss.Style
	detailBorderStyle  lipgloss.Style
	detailLabelStyle   lipgloss.Style
	detailValueStyle   lipgloss.Style
	headerStyle        lipgloss.Style
	errorStyle         lipgloss.Style
	helpStyle          lipgloss.Style
	statsPopupStyle    lipgloss.Style
	statsSectionStyle  lipgloss.Style
	statsLabelStyle    lipgloss.Style
	statsValueStyle    lipgloss.Style
	statsBarFill       lipgloss.Style
	statsBarEmpty      lipgloss.Style
	tokensLowStyle     lipgloss.Style
	tokensMedStyle     lipgloss.Style
	tokensHighStyle    lipgloss.Style
	searchMatchStyle   lipgloss.Style
)

func init() {
	applyTheme()
}

func applyTheme() {
	t := activeTheme

	titleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Accent).
		Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
		Foreground(t.Dim).
		Padding(0, 1)

	runningStyle = lipgloss.NewStyle().
		Foreground(t.Running).
		Bold(true)

	selectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.SelectedFg).
		Background(t.SelectedBg)

	dimStyle = lipgloss.NewStyle().
		Foreground(t.Dim)

	detailBorderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2)

	detailLabelStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Accent).
		Width(16)

	detailValueStyle = lipgloss.NewStyle().
		Foreground(t.Text)

	headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Text).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(t.Dim)

	errorStyle = lipgloss.NewStyle().
		Foreground(t.Error).
		Bold(true)

	helpStyle = lipgloss.NewStyle().
		Foreground(t.Dim)

	statsPopupStyle = lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(t.Accent).
		Padding(1, 3).
		Background(t.StatsBg)

	statsSectionStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Accent).
		MarginTop(1)

	statsLabelStyle = lipgloss.NewStyle().
		Foreground(t.Text).
		Width(20)

	statsValueStyle = lipgloss.NewStyle().
		Foreground(t.StatsHighlight).
		Bold(true)

	statsBarFill = lipgloss.NewStyle().
		Foreground(t.BarFill)

	statsBarEmpty = lipgloss.NewStyle().
		Foreground(t.BarEmpty)

	tokensLowStyle = lipgloss.NewStyle().
		Foreground(t.TokensLow)

	tokensMedStyle = lipgloss.NewStyle().
		Foreground(t.TokensMed)

	tokensHighStyle = lipgloss.NewStyle().
		Foreground(t.TokensHigh)

	searchMatchStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(t.Accent)
}
