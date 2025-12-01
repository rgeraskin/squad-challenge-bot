package handlers

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// showSuperAdminMenu shows the super admin menu
func (h *Handler) showSuperAdminMenu(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	admins, _ := h.superAdmin.GetAll()

	msg := "🔑 <b>Super Admin Panel</b>\n\n"
	msg += fmt.Sprintf("👑 Super Admins: %d\n", len(admins))

	allChallenges, _ := h.superAdmin.GetAllChallenges()
	msg += fmt.Sprintf("🏆 Total Challenges: %d\n", len(allChallenges))

	return c.Send(msg, keyboards.SuperAdminMenu(), tele.ModeHTML)
}

// showAllChallengesObserver shows challenges where the super admin is NOT a participant
func (h *Handler) showAllChallengesObserver(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	allChallenges, err := h.superAdmin.GetAllChallenges()
	if err != nil {
		return h.sendError(c, "Failed to load challenges.")
	}

	// Get user's own challenges (where they are a participant)
	userChallenges, _ := h.challenge.GetByUserID(userID)
	userChallengeIDs := make(map[string]bool)
	for _, ch := range userChallenges {
		userChallengeIDs[ch.ID] = true
	}

	// Filter out challenges where user is already a participant
	var otherChallenges []*domain.Challenge
	for _, ch := range allChallenges {
		if !userChallengeIDs[ch.ID] {
			otherChallenges = append(otherChallenges, ch)
		}
	}

	if len(otherChallenges) == 0 {
		return c.Send(
			"No other challenges to observe. You're a participant in all existing challenges!",
			keyboards.BackToSuperAdmin(),
		)
	}

	// Get task counts for filtered challenges only
	taskCounts := make(map[string]int)
	participantCounts := make(map[string]int)
	for _, ch := range otherChallenges {
		count, _ := h.task.CountByChallengeID(ch.ID)
		taskCounts[ch.ID] = count
		pCount, _ := h.participant.CountByChallengeID(ch.ID)
		participantCounts[ch.ID] = pCount
	}

	msg := "👁 <b>Other Challenges (Observer Mode)</b>\n\n"
	msg += "Challenges where you're not a participant. Select a challenge to observe"

	return c.Send(
		msg,
		keyboards.AllChallengesObserver(otherChallenges, taskCounts, participantCounts),
		tele.ModeHTML,
	)
}

// handleObserveChallenge enters observer mode for a challenge
func (h *Handler) handleObserveChallenge(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "Challenge not found.")
	}

	// Set current challenge for navigation
	h.state.SetCurrentChallenge(userID, challengeID)

	// Check if super admin is also a participant
	participant, _ := h.participant.GetByChallengeAndUser(challengeID, userID)
	isParticipant := participant != nil

	return h.showObserverChallengeView(c, challenge, isParticipant)
}

// showObserverChallengeView shows the observer view for a challenge
func (h *Handler) showObserverChallengeView(
	c tele.Context,
	challenge *domain.Challenge,
	isParticipant bool,
) error {
	userID := c.Sender().ID

	tasks, _ := h.task.GetByChallengeID(challenge.ID)
	participants, _ := h.participant.GetByChallengeID(challenge.ID)

	msg := "👁 <b>Observer Mode</b>\n\n"
	msg += fmt.Sprintf("🏆 <b>%s</b>\n", challenge.Name)
	if challenge.Description != "" {
		msg += fmt.Sprintf("<i>%s</i>\n", challenge.Description)
	}
	msg += fmt.Sprintf("\n📋 Tasks: %d\n", len(tasks))
	msg += fmt.Sprintf("👥 Participants: %d/10\n", len(participants))
	msg += fmt.Sprintf("🆔 ID: <code>%s</code>\n", challenge.ID)
	msg += fmt.Sprintf("👤 Creator ID: <code>%d</code>\n", challenge.CreatorID)

	if challenge.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("🕓 Daily Limit: %d/day\n", challenge.DailyTaskLimit)
	} else {
		msg += "🕓 Daily Limit: Unlimited\n"
	}

	if challenge.HideFutureTasks {
		msg += "👁 Mode: Sequential\n"
	} else {
		msg += "👁 Mode: All Visible\n"
	}

	if isParticipant {
		msg += "\n✅ <i>You are a participant in this challenge</i>"
	} else {
		msg += "\n👻 <i>Observer only - you cannot complete tasks</i>"
	}

	// Store that we're in observer mode
	tempData := map[string]any{
		TempKeyObserverMode: true,
	}
	h.state.SetStateWithData(userID, domain.StateIdle, tempData)

	kb := keyboards.ObserverChallengeView(challenge.ID, isParticipant)
	return c.Send(msg, kb, tele.ModeHTML)
}

// handleGrantSuperAdmin starts the grant super admin flow
func (h *Handler) handleGrantSuperAdmin(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	h.state.SetState(userID, domain.StateAwaitingSuperAdminID)

	msg := "🔑 <b>Grant Super Admin</b>\n\n"
	msg += "Enter the Telegram User ID of the person you want to make a super admin.\n\n"
	msg += "<i>Tip: They can find their ID by messaging @userinfobot</i>"

	return c.Send(msg, keyboards.CancelOnly(), tele.ModeHTML)
}

// processGrantSuperAdmin processes the super admin grant
func (h *Handler) processGrantSuperAdmin(c tele.Context, input string) error {
	userID := c.Sender().ID

	targetID, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
	if err != nil || targetID <= 0 {
		return c.Send(
			"Invalid Telegram ID. Please enter a valid numeric ID:",
			keyboards.CancelOnly(),
		)
	}

	err = h.superAdmin.Grant(userID, targetID)
	if err != nil {
		switch err {
		case service.ErrAlreadySuperAdmin:
			h.state.Reset(userID)
			return c.Send("That user is already a super admin.", keyboards.BackToSuperAdmin())
		default:
			h.state.Reset(userID)
			return h.sendError(c, "Failed to grant super admin privileges.")
		}
	}

	h.state.Reset(userID)
	msg := fmt.Sprintf("✅ User %d is now a super admin!", targetID)
	return c.Send(msg, keyboards.BackToSuperAdmin())
}

// showManageSuperAdmins shows the manage super admins view
func (h *Handler) showManageSuperAdmins(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	admins, err := h.superAdmin.GetAll()
	if err != nil {
		return h.sendError(c, "Failed to load super admins.")
	}

	msg := "👑 <b>Super Admins</b>\n\n"
	for _, admin := range admins {
		if admin.TelegramID == userID {
			msg += fmt.Sprintf("• <code>%d</code> (you)\n", admin.TelegramID)
		} else {
			msg += fmt.Sprintf("• <code>%d</code>\n", admin.TelegramID)
		}
	}

	return c.Send(msg, keyboards.ManageSuperAdmins(admins, userID), tele.ModeHTML)
}

// handleRevokeSuperAdmin revokes super admin from a user
func (h *Handler) handleRevokeSuperAdmin(c tele.Context, targetIDStr string) error {
	userID := c.Sender().ID

	targetID, err := strconv.ParseInt(targetIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid user ID.")
	}

	err = h.superAdmin.Revoke(userID, targetID)
	if err != nil {
		switch err {
		case service.ErrCannotRemoveSelf:
			return c.Send(
				"You cannot remove yourself as super admin.",
				keyboards.BackToSuperAdmin(),
			)
		case service.ErrSuperAdminNotFound:
			return c.Send("That user is not a super admin.", keyboards.BackToSuperAdmin())
		default:
			return h.sendError(c, "Failed to revoke super admin privileges.")
		}
	}

	msg := fmt.Sprintf("✅ User %d is no longer a super admin.", targetID)
	return c.Send(msg, keyboards.BackToSuperAdmin())
}

// handleBackToObserver returns to observer view
func (h *Handler) handleBackToObserver(c tele.Context) error {
	userID := c.Sender().ID

	userState, _ := h.state.Get(userID)
	challenge, err := h.challenge.GetByID(userState.CurrentChallenge)
	if err != nil {
		return h.showAllChallengesObserver(c)
	}

	participant, _ := h.participant.GetByChallengeAndUser(userState.CurrentChallenge, userID)
	return h.showObserverChallengeView(c, challenge, participant != nil)
}
