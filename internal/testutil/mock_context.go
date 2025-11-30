package testutil

import (
	"context"
	"time"

	tele "gopkg.in/telebot.v3"
)

// MockContext implements tele.Context for testing
type MockContext struct {
	SenderUser    *tele.User
	MessageObj    *tele.Message
	CallbackObj   *tele.Callback
	SentMessages  []interface{}
	SentOptions   []interface{}
	RespondCalled bool
	EditedMessage interface{}
	DeleteCalled  bool
}

// NewMockContext creates a new mock context with default user
func NewMockContext(userID int64) *MockContext {
	return &MockContext{
		SenderUser: &tele.User{
			ID:        userID,
			FirstName: "Test",
			LastName:  "User",
			Username:  "testuser",
		},
		MessageObj: &tele.Message{
			ID: 1,
			Sender: &tele.User{
				ID: userID,
			},
		},
		SentMessages: make([]interface{}, 0),
		SentOptions:  make([]interface{}, 0),
	}
}

// WithCallback adds a callback to the context
// Note: telebot v3 prefixes callback data with \f (form feed) character
// and uses | as separator when using menu.Data(text, unique, data1, data2, ...)
// Format: \f<unique>|<data1>|<data2>...
// Example: WithCallback("select_emoji|🦁") simulates menu.Data(text, "select_emoji", "🦁")
func (m *MockContext) WithCallback(data string) *MockContext {
	m.CallbackObj = &tele.Callback{
		ID:   "test_callback",
		Data: "\f" + data, // Simulate telebot v3 behavior
		Sender: &tele.User{
			ID: m.SenderUser.ID,
		},
	}
	return m
}

// WithMessage adds a message with text
func (m *MockContext) WithMessage(text string) *MockContext {
	m.MessageObj = &tele.Message{
		ID:   1,
		Text: text,
		Sender: &tele.User{
			ID: m.SenderUser.ID,
		},
	}
	return m
}

// WithPayload adds a start payload (deep link)
func (m *MockContext) WithPayload(payload string) *MockContext {
	m.MessageObj.Payload = payload
	return m
}

// Bot returns nil (not needed for most tests)
func (m *MockContext) Bot() *tele.Bot {
	return nil
}

// Update returns empty update
func (m *MockContext) Update() tele.Update {
	return tele.Update{}
}

// Message returns the message
func (m *MockContext) Message() *tele.Message {
	return m.MessageObj
}

// Callback returns the callback
func (m *MockContext) Callback() *tele.Callback {
	return m.CallbackObj
}

// Query returns nil
func (m *MockContext) Query() *tele.Query {
	return nil
}

// InlineResult returns nil
func (m *MockContext) InlineResult() *tele.InlineResult {
	return nil
}

// ShippingQuery returns nil
func (m *MockContext) ShippingQuery() *tele.ShippingQuery {
	return nil
}

// PreCheckoutQuery returns nil
func (m *MockContext) PreCheckoutQuery() *tele.PreCheckoutQuery {
	return nil
}

// Poll returns nil
func (m *MockContext) Poll() *tele.Poll {
	return nil
}

// PollAnswer returns nil
func (m *MockContext) PollAnswer() *tele.PollAnswer {
	return nil
}

// ChatMember returns nil
func (m *MockContext) ChatMember() *tele.ChatMemberUpdate {
	return nil
}

// ChatJoinRequest returns nil
func (m *MockContext) ChatJoinRequest() *tele.ChatJoinRequest {
	return nil
}

// Migration returns 0, 0
func (m *MockContext) Migration() (int64, int64) {
	return 0, 0
}

// Topic returns nil
func (m *MockContext) Topic() *tele.Topic {
	return nil
}

// Boost returns nil
func (m *MockContext) Boost() *tele.BoostUpdated {
	return nil
}

// BoostRemoved returns nil
func (m *MockContext) BoostRemoved() *tele.BoostRemoved {
	return nil
}

// Sender returns the sender
func (m *MockContext) Sender() *tele.User {
	return m.SenderUser
}

// Chat returns the chat
func (m *MockContext) Chat() *tele.Chat {
	return &tele.Chat{
		ID:   m.SenderUser.ID,
		Type: tele.ChatPrivate,
	}
}

// Recipient returns the recipient
func (m *MockContext) Recipient() tele.Recipient {
	return m.Chat()
}

// Text returns message text
func (m *MockContext) Text() string {
	if m.MessageObj != nil {
		return m.MessageObj.Text
	}
	return ""
}

// Data returns callback data
func (m *MockContext) Data() string {
	if m.CallbackObj != nil {
		return m.CallbackObj.Data
	}
	return ""
}

// Args returns empty slice
func (m *MockContext) Args() []string {
	return []string{}
}

// Entities returns nil
func (m *MockContext) Entities() tele.Entities {
	return nil
}

// Send records the sent message
func (m *MockContext) Send(what interface{}, opts ...interface{}) error {
	m.SentMessages = append(m.SentMessages, what)
	m.SentOptions = append(m.SentOptions, opts)
	return nil
}

// SendAlbum records sent album
func (m *MockContext) SendAlbum(a tele.Album, opts ...interface{}) error {
	m.SentMessages = append(m.SentMessages, a)
	return nil
}

// Reply records the reply
func (m *MockContext) Reply(what interface{}, opts ...interface{}) error {
	return m.Send(what, opts...)
}

// Forward returns nil
func (m *MockContext) Forward(msg tele.Editable, opts ...interface{}) error {
	return nil
}

// ForwardTo returns nil
func (m *MockContext) ForwardTo(to tele.Recipient, opts ...interface{}) error {
	return nil
}

// Edit records the edit
func (m *MockContext) Edit(what interface{}, opts ...interface{}) error {
	m.EditedMessage = what
	return nil
}

// EditCaption returns nil
func (m *MockContext) EditCaption(caption string, opts ...interface{}) error {
	return nil
}

// EditOrSend sends or edits
func (m *MockContext) EditOrSend(what interface{}, opts ...interface{}) error {
	return m.Send(what, opts...)
}

// EditOrReply replies or edits
func (m *MockContext) EditOrReply(what interface{}, opts ...interface{}) error {
	return m.Send(what, opts...)
}

// Delete records delete
func (m *MockContext) Delete() error {
	m.DeleteCalled = true
	return nil
}

// DeleteAfter does nothing
func (m *MockContext) DeleteAfter(d time.Duration) *time.Timer {
	return nil
}

// Notify does nothing
func (m *MockContext) Notify(action tele.ChatAction) error {
	return nil
}

// Ship does nothing
func (m *MockContext) Ship(prices ...interface{}) error {
	return nil
}

// Accept does nothing
func (m *MockContext) Accept(errorMessage ...string) error {
	return nil
}

// Answer does nothing
func (m *MockContext) Answer(resp *tele.QueryResponse) error {
	return nil
}

// Respond records respond call
func (m *MockContext) Respond(resp ...*tele.CallbackResponse) error {
	m.RespondCalled = true
	return nil
}

// RespondText responds with text
func (m *MockContext) RespondText(text string) error {
	m.RespondCalled = true
	return nil
}

// RespondAlert responds with alert
func (m *MockContext) RespondAlert(text string) error {
	m.RespondCalled = true
	return nil
}

// Get returns nil
func (m *MockContext) Get(key string) interface{} {
	return nil
}

// Set does nothing
func (m *MockContext) Set(key string, val interface{}) {
}

// LastMessage returns the last sent message as string
func (m *MockContext) LastMessage() string {
	if len(m.SentMessages) == 0 {
		return ""
	}
	last := m.SentMessages[len(m.SentMessages)-1]
	if s, ok := last.(string); ok {
		return s
	}
	return ""
}

// MessageCount returns the number of sent messages
func (m *MockContext) MessageCount() int {
	return len(m.SentMessages)
}

// Implement remaining required interface methods

func (m *MockContext) Context() context.Context {
	return context.Background()
}
