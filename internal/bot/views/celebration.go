package views

import (
	"fmt"
	"strings"
	"time"
)

// CelebrationData holds data for the celebration view
type CelebrationData struct {
	ChallengeName  string
	TotalTasks     int
	CompletedTasks int
	TimeTaken      time.Duration
	TeamStatus     []*TeamMemberStatus
}

// TeamMemberStatus holds completion status for a team member
type TeamMemberStatus struct {
	Emoji          string
	Name           string
	IsCompleted    bool
	CompletedTasks int
	TotalTasks     int
}

// RenderCelebration renders the celebration view
func RenderCelebration(data CelebrationData) string {
	var sb strings.Builder

	sb.WriteString("\n🎉🎊🏆 YOU DID IT! 🏆🎊🎉\n\n")
	sb.WriteString(fmt.Sprintf("Woohoo! You crushed \"%s\"!\n\n", data.ChallengeName))

	// Stats
	sb.WriteString(fmt.Sprintf("🕓 Finished in %s\n", formatDuration(data.TimeTaken)))
	sb.WriteString(fmt.Sprintf("📊 %d/%d tasks done\n\n", data.CompletedTasks, data.TotalTasks))

	// Squad status
	sb.WriteString("👥 How's the squad doing:\n")
	for _, member := range data.TeamStatus {
		if member.IsCompleted {
			sb.WriteString(fmt.Sprintf("%s %s — ✅ Crushed it!\n", member.Emoji, member.Name))
		} else {
			sb.WriteString(fmt.Sprintf("%s %s — 🔄 %d/%d\n",
				member.Emoji, member.Name, member.CompletedTasks, member.TotalTasks))
		}
	}

	return sb.String()
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours() / 24)
	if days == 0 {
		hours := int(d.Hours())
		if hours == 0 {
			minutes := int(d.Minutes())
			return fmt.Sprintf("%d minutes", minutes)
		}
		return fmt.Sprintf("%d hours", hours)
	}
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}
