package api

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"smieci-sms/internal/messages"
	"smieci-sms/internal/model"
	"smieci-sms/internal/repository"
	"smieci-sms/internal/service"
)

// MockUserRepository implements repository.UserRepository for testing
type MockUserRepository struct {
	repository.UserRepository
	deletedChatID int64
	deleteCalled  bool
}

func (m *MockUserRepository) DeleteUserLocationByChatID(ctx context.Context, chatID int64) error {
	m.deletedChatID = chatID
	m.deleteCalled = true
	return nil
}

// MockGarbageService implements service.GarbageService for testing
type MockGarbageService struct {
	service.GarbageService
}

// MockTelegramService implements service.TelegramService for testing
type MockTelegramService struct {
	service.TelegramService
	sentChatID int64
	sentText   string
	sendCalled bool
}

func (m *MockTelegramService) SendMessage(ctx context.Context, chatID int64, text string, markup *model.TelegramInlineMenu) error {
	m.sentChatID = chatID
	m.sentText = text
	m.sendCalled = true
	return nil
}

func TestTelegramHandler_AnulujCommand(t *testing.T) {
	// Set up dependencies
	repo := &MockUserRepository{}
	garbageSvc := &MockGarbageService{}
	telegramSvc := &MockTelegramService{}
	secretToken := "my-secret-token"

	handler := NewTelegramHandler(repo, garbageSvc, secretToken, telegramSvc)

	// Add a session beforehand to verify it gets deleted
	chatID := int64(987654321)
	session := getSession(chatID)
	session.State = StateAwaitingStreet
	session.Street = "Testowa"

	// Prepare payload
	reqPayload := model.TelegramRequest{
		UpdateID: 12345,
		Message: &model.TelegramMessage{
			MessageID: 1,
			Chat: model.TelegramChat{
				ID:   chatID,
				Type: "private",
			},
			Text: "/anuluj",
		},
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/telegram/start", bytes.NewBuffer(body))
	req.Header.Set("X-Telegram-Bot-Api-Secret-Token", secretToken)

	w := httptest.NewRecorder()

	// Call the handler
	handler.Start(w, req)

	// Verify HTTP status code
	if w.Code != http.StatusOK {
		t.Errorf("expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Verify repository delete was called
	if !repo.deleteCalled {
		t.Errorf("expected DeleteUserLocationByChatID to be called")
	}
	if repo.deletedChatID != chatID {
		t.Errorf("expected deleted chat ID to be %d, got %d", chatID, repo.deletedChatID)
	}

	// Verify Telegram message was sent
	if !telegramSvc.sendCalled {
		t.Errorf("expected SendMessage to be called")
	}
	if telegramSvc.sentChatID != chatID {
		t.Errorf("expected telegram chat ID to be %d, got %d", chatID, telegramSvc.sentChatID)
	}
	if telegramSvc.sentText != messages.DataDeleted {
		t.Errorf("expected telegram text to be %q, got %q", messages.DataDeleted, telegramSvc.sentText)
	}

	// Verify that the session has been deleted/reset (so State is now StateNone)
	sessionMutex.Lock()
	_, exists := sessions[chatID]
	sessionMutex.Unlock()

	if exists {
		t.Errorf("expected session for chat ID %d to be deleted from global map", chatID)
	}

	// Double check that getting a new session starts with StateNone
	newSess := getSession(chatID)
	if newSess.State != StateNone {
		t.Errorf("expected new session state to be StateNone, got %v", newSess.State)
	}
}
