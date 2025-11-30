package views

import (
	"fmt"
	"sort"
	"strings"
)

// TeamProgressData holds data for team progress view
type TeamProgressData struct {
	ChallengeName string
	Participants  []*ParticipantProgress
}

// ParticipantProgress holds progress info for a participant
type ParticipantProgress struct {
	Emoji          string
	Name           string
	IsAdmin        bool
	CompletedTasks int
	TotalTasks     int
}

// RenderTeamProgress renders the team progress view
func RenderTeamProgress(data TeamProgressData) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("👥 Squad Progress - %s\n", data.ChallengeName))
	sb.WriteString("─────────────────────────\n\n")

	// Sort by completion percentage descending
	sort.Slice(data.Participants, func(i, j int) bool {
		pctI := float64(data.Participants[i].CompletedTasks) / float64(max(data.Participants[i].TotalTasks, 1))
		pctJ := float64(data.Participants[j].CompletedTasks) / float64(max(data.Participants[j].TotalTasks, 1))
		return pctI > pctJ
	})

	for _, p := range data.Participants {
		// Name with admin indicator
		name := p.Name
		if p.IsAdmin {
			name += " (Admin)"
		}

		// Progress bar
		var pct int
		if p.TotalTasks > 0 {
			pct = p.CompletedTasks * 100 / p.TotalTasks
		}
		bar := renderProgressBar(pct)

		sb.WriteString(fmt.Sprintf("%s %d%% (%d/%d)  %s %s\n",
			bar, pct, p.CompletedTasks, p.TotalTasks, p.Emoji, name))
	}

	return sb.String()
}

func renderProgressBar(pct int) string {
	filled := pct / 10
	empty := 10 - filled

	return strings.Repeat("█", filled) + strings.Repeat("░", empty)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
