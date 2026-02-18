package fuzzy

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	fuzzymatch "github.com/sahilm/fuzzy"
)

const maxVisible = 10

// Renderer tied to stderr so colors work when stdout is captured by the shell wrapper.
var renderer = lipgloss.NewRenderer(os.Stderr)

var (
	selectedStyle = renderer.NewStyle().Foreground(lipgloss.Color("212")).Bold(true)
	normalStyle   = renderer.NewStyle().Foreground(lipgloss.Color("252"))
	promptStyle   = renderer.NewStyle().Foreground(lipgloss.Color("39"))
)

type model struct {
	textInput textinput.Model
	items     []string
	filtered  []string
	cursor    int
	selected  string
	cancelled bool
}

func newModel(items []string) model {
	ti := textinput.New()
	ti.Placeholder = "Search repos..."
	ti.Focus()
	ti.Prompt = promptStyle.Render("> ")

	return model{
		textInput: ti,
		items:     items,
		filtered:  items,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.selected = m.filtered[m.cursor]
			}
			return m, tea.Quit
		case tea.KeyUp, tea.KeyCtrlP:
			if m.cursor > 0 {
				m.cursor--
			}
			return m, nil
		case tea.KeyDown, tea.KeyCtrlN:
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)

	// Re-filter on every keystroke
	query := m.textInput.Value()
	if query == "" {
		m.filtered = m.items
	} else {
		matches := fuzzymatch.Find(query, m.items)
		m.filtered = make([]string, len(matches))
		for i, match := range matches {
			m.filtered[i] = match.Str
		}
	}

	// Reset cursor if out of bounds
	if m.cursor >= len(m.filtered) {
		m.cursor = max(0, len(m.filtered)-1)
	}

	return m, cmd
}

func (m model) View() string {
	var b strings.Builder

	b.WriteString(m.textInput.View())
	b.WriteString("\n")

	if len(m.filtered) == 0 {
		b.WriteString("  no matches\n")
		return b.String()
	}

	// Show a window of items around the cursor
	start := 0
	if m.cursor >= maxVisible {
		start = m.cursor - maxVisible + 1
	}
	end := start + maxVisible
	if end > len(m.filtered) {
		end = len(m.filtered)
	}

	for i := start; i < end; i++ {
		if i == m.cursor {
			b.WriteString(fmt.Sprintf("  %s\n", selectedStyle.Render(m.filtered[i])))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", normalStyle.Render(m.filtered[i])))
		}
	}

	return b.String()
}

// Run opens an interactive fuzzy finder on stderr and returns the selected item.
// Returns an empty string if the user cancels.
func Run(items []string) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	m := newModel(items)

	// Open /dev/tty directly for input so the TUI works
	// even when stdout is captured by the shell wrapper's $()
	tty, err := os.Open("/dev/tty")
	if err != nil {
		return "", fmt.Errorf("could not open /dev/tty: %w", err)
	}
	defer tty.Close()

	// Render on stderr so stdout remains clean for shell eval
	p := tea.NewProgram(m,
		tea.WithOutput(os.Stderr),
		tea.WithInput(tty),
	)

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("fuzzy finder error: %w", err)
	}

	result := finalModel.(model)
	if result.cancelled {
		return "", nil
	}
	return result.selected, nil
}
