package handlers

import (
	"bytes"
	"fmt"
	"strconv"

	"github.com/rgeraskin/squad-challenge-bot/internal/bot/assets"
	"github.com/rgeraskin/squad-challenge-bot/internal/bot/keyboards"
	"github.com/rgeraskin/squad-challenge-bot/internal/domain"
	"github.com/rgeraskin/squad-challenge-bot/internal/service"
	tele "gopkg.in/telebot.v3"
)

// ===== Super Admin - Templates Add Flow =====

// showTemplatesAddPanel shows the panel to select a challenge for template creation
func (h *Handler) showTemplatesAddPanel(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	// Get ALL challenges
	allChallenges, err := h.superAdmin.GetAllChallenges()
	if err != nil {
		return h.sendError(c, "Failed to load challenges.")
	}

	if len(allChallenges) == 0 {
		return c.Send(
			"No challenges found. Create a challenge first before making templates.",
			keyboards.BackToSuperAdmin(),
		)
	}

	// Build task counts map
	taskCounts := make(map[string]int)
	for _, ch := range allChallenges {
		count, _ := h.task.CountByChallengeID(ch.ID)
		taskCounts[ch.ID] = count
	}

	msg := "📋 <b>Create Template</b>\n\nSelect a challenge to create a template from:"

	return c.Send(msg, keyboards.ChallengeListForTemplate(allChallenges, taskCounts), tele.ModeHTML)
}

// showChallengeDetailsForTemplate shows challenge details before creating template
func (h *Handler) showChallengeDetailsForTemplate(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	challenge, err := h.challenge.GetByID(challengeID)
	if err != nil {
		return h.sendError(c, "Challenge not found.")
	}

	taskCount, _ := h.task.CountByChallengeID(challengeID)

	msg := "📋 <b>Template Preview</b>\n\n"
	msg += fmt.Sprintf("<b>Name:</b> %s\n", challenge.Name)
	if challenge.Description != "" {
		msg += fmt.Sprintf("<b>Description:</b> %s\n", challenge.Description)
	} else {
		msg += "<b>Description:</b> <i>No description</i>\n"
	}
	msg += "\n"

	if challenge.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("<b>Daily Limit:</b> %d/day\n", challenge.DailyTaskLimit)
	} else {
		msg += "<b>Daily Limit:</b> Unlimited\n"
	}

	if challenge.HideFutureTasks {
		msg += "<b>Mode:</b> Sequential\n"
	} else {
		msg += "<b>Mode:</b> All Visible\n"
	}

	msg += fmt.Sprintf("<b>Tasks:</b> %d\n", taskCount)

	return c.Send(msg, keyboards.ChallengeDetailsForTemplate(challengeID), tele.ModeHTML)
}

// showTemplateTasksPreview shows tasks of a challenge (read-only) for template preview
func (h *Handler) showTemplateTasksPreview(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	tasks, err := h.task.GetByChallengeID(challengeID)
	if err != nil {
		return h.sendError(c, "Failed to load tasks.")
	}

	if len(tasks) == 0 {
		return c.Send(
			"No tasks in this challenge.",
			keyboards.ChallengeDetailsForTemplate(challengeID),
		)
	}

	msg := "📋 <b>Template Tasks</b>\n\nTasks that will be included in the template:"

	return c.Send(
		msg,
		keyboards.TemplateTasksPreviewFromChallenge(tasks, challengeID),
		tele.ModeHTML,
	)
}

// showSATplTaskDetail shows task detail for super admin (challenge task preview)
func (h *Handler) showSATplTaskDetail(c tele.Context, challengeID string, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.task.GetByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	msg := fmt.Sprintf("📋 <b>Task %d</b>\n\n", task.OrderNum)
	msg += fmt.Sprintf("<b>%s</b>\n", task.Title)
	if task.Description != "" {
		msg += fmt.Sprintf("\n%s\n", task.Description)
	}

	kb := keyboards.BackToSATplTasks(challengeID)

	if task.ImageFileID != "" {
		photo := &tele.Photo{
			File:    tele.File{FileID: task.ImageFileID},
			Caption: msg,
		}
		// Try to send with image, fall back to text if image fails
		if err := c.Send(photo, kb, tele.ModeHTML); err == nil {
			return nil
		}
	}

	return c.Send(msg, kb, tele.ModeHTML)
}

// handleCreateTemplate creates a template from a challenge
func (h *Handler) handleCreateTemplate(c tele.Context, challengeID string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	template, err := h.template.CreateFromChallenge(challengeID)
	if err != nil {
		if err == service.ErrTemplateNameExists {
			return c.Send(
				"⚠️ A template with this name already exists. Choose a different challenge or delete the existing template first.",
				keyboards.BackToSuperAdmin(),
			)
		}
		return h.sendError(c, "Failed to create template.")
	}

	taskCount, _ := h.template.GetTaskCount(template.ID)

	msg := fmt.Sprintf("✅ Template '<b>%s</b>' created with %d tasks!", template.Name, taskCount)
	return c.Send(msg, keyboards.BackToSuperAdmin(), tele.ModeHTML)
}

// ===== Super Admin - Templates Delete Flow =====

// showTemplatesDeletePanel shows list of templates for deletion
func (h *Handler) showTemplatesDeletePanel(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templates, err := h.template.GetAll()
	if err != nil {
		return h.sendError(c, "Failed to load templates.")
	}

	if len(templates) == 0 {
		return c.Send(
			"No templates found. Create a template first.",
			keyboards.BackToSuperAdmin(),
		)
	}

	// Build task counts map
	taskCounts := make(map[int64]int)
	for _, tpl := range templates {
		count, _ := h.template.GetTaskCount(tpl.ID)
		taskCounts[tpl.ID] = count
	}

	msg := "🗑 <b>Delete Template</b>\n\nSelect a template to delete:"

	return c.Send(msg, keyboards.TemplatesDeleteList(templates, taskCounts), tele.ModeHTML)
}

// handleDeleteTemplateSelect shows confirmation for template deletion
func (h *Handler) handleDeleteTemplateSelect(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	taskCount, _ := h.template.GetTaskCount(templateID)

	msg := "🗑 <b>Delete Template?</b>\n\n"
	msg += fmt.Sprintf("\"<b>%s</b>\" with %d tasks will be deleted.\n\n", template.Name, taskCount)
	msg += "<i>This cannot be undone!</i>"

	return c.Send(msg, keyboards.DeleteTemplateConfirm(templateID), tele.ModeHTML)
}

// handleConfirmDeleteTemplate confirms and deletes template
func (h *Handler) handleConfirmDeleteTemplate(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	err = h.template.Delete(templateID)
	if err != nil {
		return h.sendError(c, "Failed to delete template.")
	}

	return c.Send("✅ Template deleted!", keyboards.BackToSuperAdmin())
}

// ===== User - Template Selection During Challenge Creation =====

// showTemplateOrScratchChoice shows the choice between template and scratch
func (h *Handler) showTemplateOrScratchChoice(c tele.Context) error {
	userID := c.Sender().ID

	h.state.SetState(userID, domain.StateSelectTemplateOrScratch)

	msg := "🏆 <i>Let's create a challenge!</i>\n\nHow would you like to create it?"

	return c.Send(msg, keyboards.TemplateOrScratchChoice(), tele.ModeHTML)
}

// showTemplatesList shows available templates for selection
func (h *Handler) showTemplatesList(c tele.Context) error {
	userID := c.Sender().ID

	templates, err := h.template.GetAll()
	if err != nil {
		return h.sendError(c, "Failed to load templates.")
	}

	if len(templates) == 0 {
		// No templates - go directly to scratch flow
		return h.handleFromScratch(c)
	}

	h.state.SetState(userID, domain.StateSelectTemplate)

	// Build task counts map
	taskCounts := make(map[int64]int)
	for _, tpl := range templates {
		count, _ := h.template.GetTaskCount(tpl.ID)
		taskCounts[tpl.ID] = count
	}

	msg := "📋 <b>Select Template</b>\n\nChoose a template for your challenge:"

	return c.Send(msg, keyboards.TemplatesList(templates, taskCounts), tele.ModeHTML)
}

// showTemplateDetails shows template details for user
func (h *Handler) showTemplateDetails(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	taskCount, _ := h.template.GetTaskCount(templateID)

	h.state.SetState(userID, domain.StateViewingTemplate)

	msg := fmt.Sprintf("📋 <b>Template: %s</b>\n\n", template.Name)
	if template.Description != "" {
		msg += fmt.Sprintf("<i>%s</i>\n\n", template.Description)
	}

	if template.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("<b>Daily Limit:</b> %d/day\n", template.DailyTaskLimit)
	} else {
		msg += "<b>Daily Limit:</b> Unlimited\n"
	}

	if template.HideFutureTasks {
		msg += "<b>Mode:</b> Sequential\n"
	} else {
		msg += "<b>Mode:</b> All Visible\n"
	}

	msg += fmt.Sprintf("<b>Tasks:</b> %d\n", taskCount)

	return c.Send(msg, keyboards.TemplateDetails(templateID), tele.ModeHTML)
}

// showTemplateTasksList shows template tasks (read-only)
func (h *Handler) showTemplateTasksList(c tele.Context, templateIDStr string) error {
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	tasks, err := h.template.GetTasks(templateID)
	if err != nil {
		return h.sendError(c, "Failed to load tasks.")
	}

	if len(tasks) == 0 {
		return c.Send(
			"No tasks in this template.",
			keyboards.TemplateDetails(templateID),
		)
	}

	msg := "📋 <b>Template Tasks</b>\n\nTasks that will be included:"

	return c.Send(msg, keyboards.TemplateTasksList(tasks, templateID), tele.ModeHTML)
}

// showTplTaskDetail shows task detail for user (template task preview)
func (h *Handler) showTplTaskDetail(c tele.Context, templateIDStr string, taskIDStr string) error {
	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	msg := fmt.Sprintf("📋 <b>Task %d</b>\n\n", task.OrderNum)
	msg += fmt.Sprintf("<b>%s</b>\n", task.Title)
	if task.Description != "" {
		msg += fmt.Sprintf("\n%s\n", task.Description)
	}

	kb := keyboards.BackToTplTasks(templateID)

	if task.ImageFileID != "" {
		photo := &tele.Photo{
			File:    tele.File{FileID: task.ImageFileID},
			Caption: msg,
		}
		// Try to send with image, fall back to text if image fails
		if err := c.Send(photo, kb, tele.ModeHTML); err == nil {
			return nil
		}
	}

	return c.Send(msg, kb, tele.ModeHTML)
}

// handleCreateFromTemplate starts the template-based challenge creation
func (h *Handler) handleCreateFromTemplate(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	// Store template info in temp data
	tempData := map[string]interface{}{
		TempKeyTemplateID:   templateID,
		TempKeyTemplateName: template.Name,
		TempKeyFromTemplate: true,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTemplateChallengeName, tempData)

	msg := fmt.Sprintf(
		"🏆 <i>Creating from template</i>\n\nWhat should we call your challenge?\n\n<i>Suggestion: %s</i>",
		template.Name,
	)
	return c.Send(msg, keyboards.CancelOnly(), tele.ModeHTML)
}

// handleFromScratch starts the from-scratch challenge creation flow
func (h *Handler) handleFromScratch(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingChallengeName)
	return c.Send(
		"🏆 <i>Enter challenge name</i>\n\nWhat do you want to call it?",
		keyboards.CancelOnly(),
		tele.ModeHTML,
	)
}

// ===== Template-Based Challenge Creation Flow =====

// processTemplateChallengeName processes the challenge name for template-based creation
func (h *Handler) processTemplateChallengeName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 50 {
		return c.Send("😬 Keep it between 1-50 characters, please!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["challenge_name"] = name
	h.state.SetStateWithData(userID, domain.StateAwaitingTemplateCreatorName, tempData)

	return c.Send(
		"👤 What should we call you?\n\n<i>Tap Skip to use your Telegram name</i>",
		keyboards.SkipName(getTelegramName(c)),
		tele.ModeHTML,
	)
}

// processTemplateCreatorName processes creator name for template-based creation
func (h *Handler) processTemplateCreatorName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 30 {
		return c.Send("😬 Keep it between 1-30 characters!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["display_name"] = name
	h.state.SetStateWithData(userID, domain.StateAwaitingTemplateCreatorEmoji, tempData)

	return c.Send(
		"🎨 Pick an emoji that represents you!\n\n(or send your own)",
		keyboards.EmojiSelector(nil),
	)
}

// processTemplateCreatorEmoji processes creator emoji for template-based creation
func (h *Handler) processTemplateCreatorEmoji(c tele.Context, emoji string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	tempData["emoji"] = emoji
	h.state.SetStateWithData(userID, domain.StateAwaitingTemplateCreatorSyncTime, tempData)

	return h.promptTemplateSyncTime(c)
}

// promptTemplateSyncTime shows the time sync prompt for template-based creation
func (h *Handler) promptTemplateSyncTime(c tele.Context) error {
	msg := "🕐 <i>Sync Your Clock</i>\n\nThis helps track your daily progress right!\n\nWhat time is it for you? (HH:MM format)\n\n<i>Example: 14:30 or 09:15</i>"
	return c.Send(msg, keyboards.SkipTemplateSyncTime(), tele.ModeHTML)
}

// processTemplateCreatorSyncTime processes sync time for template-based creation
func (h *Handler) processTemplateCreatorSyncTime(c tele.Context, input string) error {
	offset, err := parseTimeInput(input)
	if err != nil {
		return c.Send(
			"🤔 That doesn't look right. Try HH:MM format (e.g., 14:30):",
			keyboards.SkipTemplateSyncTime(),
		)
	}

	return h.finishTemplateBasedChallengeCreation(c, offset)
}

// skipTemplateSyncTime skips sync time for template-based creation
func (h *Handler) skipTemplateSyncTime(c tele.Context) error {
	return h.finishTemplateBasedChallengeCreation(c, 0)
}

// finishTemplateBasedChallengeCreation creates challenge from template
func (h *Handler) finishTemplateBasedChallengeCreation(c tele.Context, timeOffset int) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	templateID := int64(tempData[TempKeyTemplateID].(float64))
	challengeName := tempData["challenge_name"].(string)
	displayName := tempData["display_name"].(string)
	emoji := tempData["emoji"].(string)

	// Get template and its tasks
	template, err := h.template.GetByID(templateID)
	if err != nil {
		h.state.Reset(userID)
		return h.sendError(c, "Template not found.")
	}

	templateTasks, err := h.template.GetTasks(templateID)
	if err != nil {
		h.state.Reset(userID)
		return h.sendError(c, "Failed to load template tasks.")
	}

	// Create challenge from template
	challenge, err := h.challenge.CreateFromTemplate(template, templateTasks, challengeName, userID)
	if err != nil {
		h.state.Reset(userID)
		if err == service.ErrMaxChallengesReached {
			return h.sendError(c, "😬 Whoa, you've hit the limit of 10 challenges!")
		}
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Add creator as first participant with time offset
	_, err = h.participant.Join(challenge.ID, userID, displayName, emoji, timeOffset)
	if err != nil {
		h.state.Reset(userID)
		return h.sendError(c, "😅 Oops, something went wrong. Give it another try!")
	}

	// Set current challenge and reset state
	h.state.SetCurrentChallenge(userID, challenge.ID)
	h.state.ResetKeepChallenge(userID)

	taskCount := len(templateTasks)
	msg := fmt.Sprintf(
		"🎉 \"<b>%s</b>\" is live!\n\nCreated from template with %d tasks. You're the admin!",
		challengeName,
		taskCount,
	)

	animation := &tele.Animation{
		File:     tele.FromReader(bytes.NewReader(assets.ChallengeAcceptedGIF)),
		FileName: "challenge-accepted.gif",
		MIME:     "image/gif",
		Caption:  msg,
	}
	c.Send(animation, tele.ModeHTML)

	// Show admin panel
	return h.showAdminPanel(c, challenge.ID)
}

// ===== Super Admin - Template Edit Flow =====

// showTemplatesEditPanel shows list of templates for editing
func (h *Handler) showTemplatesEditPanel(c tele.Context) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templates, err := h.template.GetAll()
	if err != nil {
		return h.sendError(c, "Failed to load templates.")
	}

	if len(templates) == 0 {
		return c.Send(
			"No templates found. Create a template first.",
			keyboards.BackToSuperAdmin(),
		)
	}

	// Build task counts map
	taskCounts := make(map[int64]int)
	for _, tpl := range templates {
		count, _ := h.template.GetTaskCount(tpl.ID)
		taskCounts[tpl.ID] = count
	}

	msg := "✏️ <b>Edit Templates</b>\n\nSelect a template to edit:"

	return c.Send(msg, keyboards.TemplatesEditList(templates, taskCounts), tele.ModeHTML)
}

// showTemplateAdminPanel shows the admin panel for a template
func (h *Handler) showTemplateAdminPanel(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	taskCount, _ := h.template.GetTaskCount(templateID)

	msg := "🔧 <i>Template Admin Panel</i>\n\n"
	msg += fmt.Sprintf("<b>Template:</b> %s\n", template.Name)
	if template.Description != "" {
		msg += fmt.Sprintf("<b>Description:</b> %s\n", template.Description)
	} else {
		msg += "<b>Description:</b>\n"
	}
	msg += fmt.Sprintf("<b>Tasks:</b> %d\n", taskCount)
	if template.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("<b>Daily Limit:</b> %d/day\n", template.DailyTaskLimit)
	} else {
		msg += "<b>Daily Limit:</b> No daily limit\n"
	}
	if template.HideFutureTasks {
		msg += "<b>Mode:</b> Sequential\n"
	} else {
		msg += "<b>Mode:</b> All Visible\n"
	}

	return c.Send(
		msg,
		keyboards.TemplateAdminPanel(templateID, template.DailyTaskLimit, template.HideFutureTasks),
		tele.ModeHTML,
	)
}

// handleEditTemplateName starts editing template name
func (h *Handler) handleEditTemplateName(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingNewTemplateName, tempData)

	msg := fmt.Sprintf("✏️ <b>Edit Name</b>\n\nCurrent: %s\n\nEnter a new name:", template.Name)
	return c.Send(msg, keyboards.BackToTemplateAdmin(templateID), tele.ModeHTML)
}

// processNewTemplateName processes the new template name
func (h *Handler) processNewTemplateName(c tele.Context, name string) error {
	userID := c.Sender().ID

	if len(name) == 0 || len(name) > 50 {
		return c.Send("😬 Keep it between 1-50 characters, please!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))

	err := h.template.UpdateName(templateID, name)
	if err != nil {
		return h.sendError(c, "Failed to update name.")
	}

	h.state.Reset(userID)
	c.Send("✅ Template name updated!")
	return h.showTemplateAdminPanel(c, fmt.Sprintf("%d", templateID))
}

// handleEditTemplateDescription starts editing template description
func (h *Handler) handleEditTemplateDescription(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingNewTemplateDescription, tempData)

	msg := "📝 <b>Edit Description</b>\n\nCurrent: "
	if template.Description != "" {
		msg += template.Description
	} else {
		msg += "<i>No description</i>"
	}
	msg += "\n\nEnter a new description:"

	return c.Send(msg, keyboards.BackToTemplateAdmin(templateID), tele.ModeHTML)
}

// processNewTemplateDescription processes the new template description
func (h *Handler) processNewTemplateDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > 500 {
		return c.Send("😬 Keep it under 500 characters, please!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))

	err := h.template.UpdateDescription(templateID, description)
	if err != nil {
		return h.sendError(c, "Failed to update description.")
	}

	h.state.Reset(userID)
	c.Send("✅ Template description updated!")
	return h.showTemplateAdminPanel(c, fmt.Sprintf("%d", templateID))
}

// handleEditTemplateDailyLimit starts editing template daily limit
func (h *Handler) handleEditTemplateDailyLimit(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingNewTemplateDailyLimit, tempData)

	msg := "🕓 <b>Edit Daily Limit</b>\n\n"
	if template.DailyTaskLimit > 0 {
		msg += fmt.Sprintf("Current: %d/day\n\n", template.DailyTaskLimit)
	} else {
		msg += "Current: Unlimited\n\n"
	}
	msg += "Enter a new limit (0 for unlimited):"

	return c.Send(msg, keyboards.BackToTemplateAdmin(templateID), tele.ModeHTML)
}

// processNewTemplateDailyLimit processes the new template daily limit
func (h *Handler) processNewTemplateDailyLimit(c tele.Context, input string) error {
	userID := c.Sender().ID

	limit, err := strconv.Atoi(input)
	if err != nil || limit < 0 || limit > 50 {
		return c.Send("😬 Enter a number between 0 and 50!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))

	err = h.template.UpdateDailyLimit(templateID, limit)
	if err != nil {
		return h.sendError(c, "Failed to update daily limit.")
	}

	h.state.Reset(userID)
	c.Send("✅ Daily limit updated!")
	return h.showTemplateAdminPanel(c, fmt.Sprintf("%d", templateID))
}

// handleToggleTemplateHideFutureTasks toggles hide future tasks setting
func (h *Handler) handleToggleTemplateHideFutureTasks(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	template, err := h.template.GetByID(templateID)
	if err != nil {
		return h.sendError(c, "Template not found.")
	}

	newValue := !template.HideFutureTasks
	err = h.template.UpdateHideFutureTasks(templateID, newValue)
	if err != nil {
		return h.sendError(c, "Failed to update setting.")
	}

	return h.showTemplateAdminPanel(c, templateIDStr)
}

// ===== Template Task Management =====

// showEditTemplateTasksList shows the edit tasks list for a template
func (h *Handler) showEditTemplateTasksList(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	tasks, err := h.template.GetTasks(templateID)
	if err != nil {
		return h.sendError(c, "Failed to load tasks.")
	}

	if len(tasks) == 0 {
		return c.Send(
			"No tasks in this template. Add some!",
			keyboards.BackToTemplateAdmin(templateID),
		)
	}

	msg := "📋 <b>Edit Tasks</b>\n\nSelect a task to edit:"

	return c.Send(msg, keyboards.EditTemplateTasksList(tasks, templateID), tele.ModeHTML)
}

// handleAddTemplateTask starts adding a new task to a template
func (h *Handler) handleAddTemplateTask(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	// Check task limit
	taskCount, _ := h.template.GetTaskCount(templateID)
	if taskCount >= 50 {
		return c.Send(
			"😬 This template already has 50 tasks (maximum)!",
			keyboards.BackToTemplateAdmin(templateID),
		)
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTplTaskTitle, tempData)

	return c.Send(
		"📝 <b>Add Task</b>\n\nWhat's the title of this task?",
		keyboards.BackToTemplateAdmin(templateID),
		tele.ModeHTML,
	)
}

// processTplTaskTitle processes the title for a new template task
func (h *Handler) processTplTaskTitle(c tele.Context, title string) error {
	userID := c.Sender().ID

	if len(title) == 0 || len(title) > 100 {
		return c.Send("😬 Keep it between 1-100 characters!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	if tempData == nil {
		tempData = make(map[string]interface{})
	}
	tempData["task_title"] = title
	h.state.SetStateWithData(userID, domain.StateAwaitingTplTaskImage, tempData)

	return c.Send(
		"🖼 Got a picture for this task? (or skip it)",
		keyboards.SkipCancel(),
	)
}

// processTplTaskImage processes the image for a new template task
func (h *Handler) processTplTaskImage(c tele.Context, imageFileID string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	if tempData == nil {
		tempData = make(map[string]interface{})
	}
	tempData["image_file_id"] = imageFileID
	h.state.SetStateWithData(userID, domain.StateAwaitingTplTaskDescription, tempData)

	return c.Send("📝 Add some details? (or skip it)", keyboards.SkipCancel())
}

// skipTplTaskImage skips the image for a new template task
func (h *Handler) skipTplTaskImage(c tele.Context) error {
	userID := c.Sender().ID
	h.state.SetState(userID, domain.StateAwaitingTplTaskDescription)
	return c.Send("📝 Add some details? (or skip it)", keyboards.SkipCancel())
}

// processTplTaskDescription processes the description for a new template task
func (h *Handler) processTplTaskDescription(c tele.Context, description string) error {
	if len(description) > domain.MaxTaskDescriptionLength {
		return c.Send(
			fmt.Sprintf("😅 That's a bit long! Keep it under %d characters:", domain.MaxTaskDescriptionLength),
			keyboards.SkipCancel(),
		)
	}

	return h.createTemplateTask(c, description)
}

// skipTplTaskDescription skips description and creates the template task
func (h *Handler) skipTplTaskDescription(c tele.Context) error {
	return h.createTemplateTask(c, "")
}

// createTemplateTask creates the template task with collected data
func (h *Handler) createTemplateTask(c tele.Context, description string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)

	templateID := int64(tempData[TempKeyTemplateID].(float64))
	title := tempData["task_title"].(string)
	var imageFileID string
	if img, ok := tempData["image_file_id"]; ok {
		imageFileID = img.(string)
	}

	task := &domain.TemplateTask{
		TemplateID:  templateID,
		Title:       title,
		Description: description,
		ImageFileID: imageFileID,
	}

	err := h.template.CreateTask(task)
	if err != nil {
		h.state.Reset(userID)
		return h.sendError(c, "Failed to create task.")
	}

	h.state.Reset(userID)

	msg := fmt.Sprintf("✅ Task #%d added: \"%s\"", task.OrderNum, task.Title)
	return c.Send(msg, keyboards.AddTemplateTaskDone(templateID))
}

// showEditTemplateTask shows the edit panel for a single template task
func (h *Handler) showEditTemplateTask(c tele.Context, templateIDStr, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	msg := fmt.Sprintf("📝 <b>Task %d: %s</b>\n\n", task.OrderNum, task.Title)
	if task.Description != "" {
		msg += fmt.Sprintf("<i>%s</i>\n\n", task.Description)
	}
	if task.ImageFileID != "" {
		msg += "📷 Has image\n"
	}

	return c.Send(msg, keyboards.EditTemplateTask(templateID, taskID), tele.ModeHTML)
}

// handleEditTplTaskTitle starts editing a template task's title
func (h *Handler) handleEditTplTaskTitle(c tele.Context, templateIDStr, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
		TempKeyTaskID:     taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTplEditTitle, tempData)

	msg := fmt.Sprintf("📝 <b>Edit Title</b>\n\nCurrent: %s\n\nEnter a new title:", task.Title)
	return c.Send(msg, keyboards.BackToTemplateAdmin(templateID), tele.ModeHTML)
}

// processTplEditTitle processes the new title for a template task
func (h *Handler) processTplEditTitle(c tele.Context, title string) error {
	userID := c.Sender().ID

	if len(title) == 0 || len(title) > 100 {
		return c.Send("😬 Keep it between 1-100 characters!", keyboards.CancelOnly())
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))
	taskID := int64(tempData[TempKeyTaskID].(float64))

	err := h.template.UpdateTaskTitle(taskID, title)
	if err != nil {
		return h.sendError(c, "Failed to update title.")
	}

	h.state.Reset(userID)
	c.Send("✅ Title updated!")
	return h.showEditTemplateTask(c, fmt.Sprintf("%d", templateID), fmt.Sprintf("%d", taskID))
}

// handleEditTplTaskDescription starts editing a template task's description
func (h *Handler) handleEditTplTaskDescription(
	c tele.Context,
	templateIDStr, taskIDStr string,
) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
		TempKeyTaskID:     taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTplEditDescription, tempData)

	msg := "📄 <b>Edit Description</b>\n\nCurrent: "
	if task.Description != "" {
		msg += task.Description
	} else {
		msg += "<i>No description</i>"
	}
	msg += "\n\nEnter a new description:"

	return c.Send(msg, keyboards.BackToTemplateAdmin(templateID), tele.ModeHTML)
}

// processTplEditDescription processes the new description for a template task
func (h *Handler) processTplEditDescription(c tele.Context, description string) error {
	userID := c.Sender().ID

	if len(description) > domain.MaxTaskDescriptionLength {
		return c.Send(
			fmt.Sprintf("😬 Keep it under %d characters!", domain.MaxTaskDescriptionLength),
			keyboards.CancelOnly(),
		)
	}

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))
	taskID := int64(tempData[TempKeyTaskID].(float64))

	err := h.template.UpdateTaskDescription(taskID, description)
	if err != nil {
		return h.sendError(c, "Failed to update description.")
	}

	h.state.Reset(userID)
	c.Send("✅ Description updated!")
	return h.showEditTemplateTask(c, fmt.Sprintf("%d", templateID), fmt.Sprintf("%d", taskID))
}

// handleEditTplTaskImage starts editing a template task's image
func (h *Handler) handleEditTplTaskImage(c tele.Context, templateIDStr, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	tempData := map[string]interface{}{
		TempKeyTemplateID: templateID,
		TempKeyTaskID:     taskID,
	}
	h.state.SetStateWithData(userID, domain.StateAwaitingTplEditImage, tempData)

	return c.Send(
		"📷 <b>Change Image</b>\n\nSend a new image (or type 'remove' to remove current image):",
		keyboards.BackToTemplateAdmin(templateID),
		tele.ModeHTML,
	)
}

// processTplEditImage processes the new image for a template task
func (h *Handler) processTplEditImage(c tele.Context, imageFileID string) error {
	userID := c.Sender().ID

	var tempData map[string]interface{}
	h.state.GetTempData(userID, &tempData)
	templateID := int64(tempData[TempKeyTemplateID].(float64))
	taskID := int64(tempData[TempKeyTaskID].(float64))

	err := h.template.UpdateTaskImage(taskID, imageFileID)
	if err != nil {
		return h.sendError(c, "Failed to update image.")
	}

	h.state.Reset(userID)
	c.Send("✅ Image updated!")
	return h.showEditTemplateTask(c, fmt.Sprintf("%d", templateID), fmt.Sprintf("%d", taskID))
}

// handleDeleteTemplateTask shows delete confirmation for a template task
func (h *Handler) handleDeleteTemplateTask(c tele.Context, templateIDStr, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	msg := fmt.Sprintf(
		"🗑 <b>Delete Task?</b>\n\n<b>%s</b>\n\n<i>This cannot be undone!</i>",
		task.Title,
	)
	return c.Send(msg, keyboards.DeleteTemplateTaskConfirm(templateID, taskID), tele.ModeHTML)
}

// handleConfirmDeleteTemplateTask confirms and deletes a template task
func (h *Handler) handleConfirmDeleteTemplateTask(
	c tele.Context,
	templateIDStr, taskIDStr string,
) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	err = h.template.DeleteTask(taskID, templateID)
	if err != nil {
		return h.sendError(c, "Failed to delete task.")
	}

	c.Send("✅ Task deleted!")
	return h.showEditTemplateTasksList(c, fmt.Sprintf("%d", templateID))
}

// ===== Template Task Reordering =====

// showReorderTemplateTasksList shows the task list for reordering
func (h *Handler) showReorderTemplateTasksList(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	tasks, err := h.template.GetTasks(templateID)
	if err != nil {
		return h.sendError(c, "Failed to load tasks.")
	}

	if len(tasks) < 2 {
		return c.Send(
			"Need at least 2 tasks to reorder.",
			keyboards.BackToTemplateAdmin(templateID),
		)
	}

	msg := "🔀 <b>Reorder Tasks</b>\n\nSelect a task to move:"

	return c.Send(msg, keyboards.ReorderTemplateTasksList(tasks, templateID), tele.ModeHTML)
}

// handleTplReorderSelect selects a task for reordering
func (h *Handler) handleTplReorderSelect(c tele.Context, templateIDStr, taskIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	task, err := h.template.GetTaskByID(taskID)
	if err != nil || task == nil {
		return h.sendError(c, "Task not found.")
	}

	taskCount, _ := h.template.GetTaskCount(templateID)

	msg := fmt.Sprintf("🔀 Moving: <b>%s</b>\n\nSelect new position:", task.Title)
	return c.Send(
		msg,
		keyboards.ReorderTemplatePositions(templateID, taskID, taskCount, task.OrderNum),
		tele.ModeHTML,
	)
}

// handleTplReorderMove moves a template task to a new position
func (h *Handler) handleTplReorderMove(
	c tele.Context,
	templateIDStr, taskIDStr, newPosStr string,
) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	taskID, err := strconv.ParseInt(taskIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid task ID.")
	}

	newPos, err := strconv.Atoi(newPosStr)
	if err != nil {
		return h.sendError(c, "Invalid position.")
	}

	if err := h.template.MoveTask(taskID, templateID, newPos); err != nil {
		return h.sendError(c, "Failed to move task.")
	}

	// Show new order (like challenge reorder)
	tasks, err := h.template.GetTasks(templateID)
	if err != nil {
		return h.sendError(c, "Failed to load tasks.")
	}

	msg := "✅ Done! Here's the new order:\n\n"
	for _, t := range tasks {
		msg += fmt.Sprintf("%d. %s\n", t.OrderNum, t.Title)
	}

	return c.Send(msg, keyboards.ReorderTemplateDone(templateID))
}

// handleTplRandomize randomizes template task order
func (h *Handler) handleTplRandomize(c tele.Context, templateIDStr string) error {
	userID := c.Sender().ID

	if !h.isSuperAdmin(userID) {
		return h.sendError(c, "You don't have super admin privileges.")
	}

	templateID, err := strconv.ParseInt(templateIDStr, 10, 64)
	if err != nil {
		return h.sendError(c, "Invalid template ID.")
	}

	err = h.template.RandomizeTaskOrder(templateID)
	if err != nil {
		return h.sendError(c, "Failed to randomize tasks.")
	}

	// Show reorder view with updated order (like challenge randomize)
	c.Send("🎲 Tasks randomized!")
	return h.showReorderTemplateTasksList(c, templateIDStr)
}
