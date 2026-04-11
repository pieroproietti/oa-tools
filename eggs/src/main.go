package main

import (
	"coa/src/middleware" // <-- Il ponte verso il cervello
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- 1. DEFINIAMO IL MESSAGGIO ---
// Bubble Tea funziona a messaggi. Questo dice: "Ehi, sono arrivati i dati della distro!"
type discoveryMsg middleware.DiscoveryData

// --- 2. IL NERVO (COMANDO) ---
// Questa funzione interroga il middleware e "impacchetta" il risultato in un messaggio.
func checkSystem() tea.Cmd {
	return func() tea.Msg {
		data := middleware.GetDiscovery()
		return discoveryMsg(data)
	}
}

// ... Stili rimangono uguali ...
var (
	purple = lipgloss.Color("63")
	pink   = lipgloss.Color("205")
	gray   = lipgloss.Color("240")

	windowStyle  = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(purple).Padding(0).Margin(1)
	sidebarStyle = lipgloss.NewStyle().Width(25).Border(lipgloss.NormalBorder(), false, true, false, false).BorderForeground(gray)
	contentStyle = lipgloss.NewStyle().Padding(1, 4)
)

var steps = []string{"01 DISCOVERY", "02 IDENTITY", "03 STRATEGY", "04 SKELETON", "05 EXCLUDES", "06 IGNITION"}

type model struct {
	cursor int
	width  int
	height int
	// --- 3. AGGIUNGIAMO I DATI AL MODELLO ---
	info middleware.DiscoveryData
}

func (m model) Init() tea.Cmd {
	// Appena parte il programma, eggs lancia il controllo del sistema
	return checkSystem()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// --- 4. GESTIAMO L'ARRIVO DEI DATI ---
	case discoveryMsg:
		m.info = middleware.DiscoveryData(msg) // Salviamo i dati reali nel modello
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(steps)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Bold(true).Padding(1, 2).Render("EGGS DASHBOARD\n\n"))

	for i, step := range steps {
		if i == m.cursor {
			sb.WriteString(lipgloss.NewStyle().Foreground(pink).Background(lipgloss.Color("235")).PaddingLeft(2).Width(23).Render("> " + step + "\n"))
		} else {
			sb.WriteString(fmt.Sprintf("  %s\n", step))
		}
		sb.WriteString("\n")
	}

	sidebar := sidebarStyle.Height(m.height - 5).Render(sb.String())

	var contentTitle string
	var body string

	switch m.cursor {
	case 0:
		contentTitle = "🔍 Discovery Phase"
		// --- 5. MOSTRIAMO I DATI REALI ---
		rootStatus := "❌ Utente Normale (usa sudo!)"
		if m.info.IsRoot {
			rootStatus = "✅ Root"
		}

		body = fmt.Sprintf(
			"Distro rilevata:  %s\nFamiglia:         %s\nArchitettura:     %s\nStato:            %s",
			m.info.DistroName,
			m.info.Family,
			m.info.Architecture,
			rootStatus,
		)
	case 1:
		contentTitle = "👤 Identity Injection"
		body = "Configurazione dell'utente Live Artisan."
	case 5:
		contentTitle = "🚀 Ignition Control"
		body = "Pronto per generare il piano di volo."
	default:
		contentTitle = steps[m.cursor]
		body = "Pannello in fase di modellazione."
	}

	rightPane := contentStyle.Width(m.width - 35).Render(
		lipgloss.JoinVertical(lipgloss.Left,
			lipgloss.NewStyle().Foreground(purple).Bold(true).Underline(true).Render(contentTitle)+"\n",
			body,
		),
	)

	return windowStyle.Render(lipgloss.JoinHorizontal(lipgloss.Top, sidebar, rightPane))
}

func main() {
	p := tea.NewProgram(model{}, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Errore eggs: %v", err)
	}
}
