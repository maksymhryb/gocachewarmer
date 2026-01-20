package ui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	p "github.com/maksymhryb/gocachewarmer/pipeline"
)

type TotalCounterMsg int

type model struct {
	results         []p.WarmupResult
	processed       int
	total           int
	width           int
	height          int
	footerHeight    int
	aggregateResult p.AggregateResult
	progress        progress.Model
	started         time.Time
	dryRun          bool
}

func InitialModel(dryRun bool) model {
	return model{
		results:      make([]p.WarmupResult, 0),
		footerHeight: 3,
		progress:     progress.New(progress.WithScaledGradient(COLOR_PROGRESS_GRADIENT_A, COLOR_PROGRESS_GRADIENT_B)),
		started:      time.Now(),
		dryRun:       dryRun,
	}
}

func (m model) GetColumnWidth() (int, int, int, int) {
	size1 := int((float64(m.width) * 0.7))
	size2 := (m.width - size1) / 3

	return size2, size1, size2, size2
}

func (m model) CalculateETA() time.Duration {
	if m.processed == 0 {
		return 0
	}
	timeElapsed := time.Since(m.started)
	perItem := timeElapsed / time.Duration(m.processed)
	eta := perItem * time.Duration(m.total-m.processed)

	return eta.Truncate(time.Second)
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.progress.Width = msg.Width

	case p.WarmupResult:
		m.results = append(
			m.results,
			msg,
		)
		m.processed++

	case p.AggregateResult:
		m.aggregateResult = msg

	case TotalCounterMsg:
		m.total = int(msg)
	}

	return m, nil
}

func (m model) View() string {
	contentHeight := max(m.height-m.footerHeight-1, 0)
	contentLines := make([]string, contentHeight)
	startIdx := 0
	if len(m.results) > contentHeight {
		startIdx = len(m.results) - contentHeight
	}
	for i := range contentHeight {
		idx := startIdx + i
		if idx < len(m.results) {
			c1, c2, c3, c4 := m.GetColumnWidth()
			contentLines[i] = fmt.Sprintf(
				"%-*d %-*s %-*s %-*s",
				c1, idx+1,
				c2, m.results[idx].Url,
				c3, renderStatus(m.results[idx].StatusCode),
				c4, m.results[idx].ResponseTime,
			)
		} else {
			contentLines[i] = ""
		}
	}

	contentArea := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	separator := styles["separator"].Render(strings.Repeat("-", m.width))

	additionalInfo := ""
	if m.dryRun {
		additionalInfo += "| Running in dry-run mode"
	}
	footerText := fmt.Sprintf(
		"Processed: %5d/%d URLs | Success: %5d | Redirect: %5d | Error: %5d | Press Q or Ctrl+C to quit\n"+
			"ETA: %8s | Avg Response Time: %5.2fms | Success Rate: %.2f %s",
		m.processed,
		m.total,
		m.aggregateResult.SuccessResponses,
		m.aggregateResult.RedirectResponses,
		m.aggregateResult.ErrorResponses,
		m.CalculateETA(),
		float64(m.aggregateResult.AvgResponseTime)/float64(time.Millisecond),
		m.aggregateResult.SuccessRate,
		additionalInfo,
	)
	footer := styles["footer"].Width(m.width).Render(footerText)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		contentArea,
		separator,
		footer,
		m.progress.ViewAs(float64(m.processed)/float64(m.total)),
	)
}

func renderStatus(status int) string {
	var styleType string
	if status >= 200 && status < 300 {
		styleType = "success"
	} else if status >= 300 && status < 400 {
		styleType = "redirect"
	} else if status >= 400 && status < 600 {
		styleType = "error"
	}

	return styles[styleType].Render(fmt.Sprintf("%d", status))
}
