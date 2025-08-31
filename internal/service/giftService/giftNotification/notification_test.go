package giftNotification

import (
	"gift-buyer/internal/config"
	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

// MockLogsWriter для тестирования
type MockLogsWriter struct{}

func (m *MockLogsWriter) Write(entry *logTypes.LogEntry) error {
	return nil
}

func (m *MockLogsWriter) LogError(message string) {}

func (m *MockLogsWriter) LogErrorf(format string, args ...interface{}) {}

func (m *MockLogsWriter) LogInfo(message string) {}

func TestNewNotification(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	mockLogsWriter := &MockLogsWriter{}

	service := NewNotification(mockClient, mockConfig, mockLogsWriter)

	assert.NotNil(t, service)
}

func TestNotificationService_Interface_Compliance(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	mockLogsWriter := &MockLogsWriter{}

	service := NewNotification(mockClient, mockConfig, mockLogsWriter)

	// Verify that the service implements the NotificationService interface
	// This is a compile-time check, but we can also verify at runtime
	assert.NotNil(t, service)
}

func TestNotificationService_Structure(t *testing.T) {
	mockClient := &tg.Client{}
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	mockLogsWriter := &MockLogsWriter{}

	service := NewNotification(mockClient, mockConfig, mockLogsWriter)

	// Cast to concrete type to verify internal structure
	assert.Equal(t, mockClient, service.Bot)
	assert.Equal(t, mockConfig, service.Config)
}

func TestNotificationService_NilClient(t *testing.T) {
	mockConfig := &config.TgSettings{
		NotificationChatID: 12345,
		TgBotKey:           "test_bot_token",
	}
	mockLogsWriter := &MockLogsWriter{}

	// Test with nil client - should not panic during creation
	service := NewNotification(nil, mockConfig, mockLogsWriter)
	assert.NotNil(t, service)

	// Cast to concrete type to verify nil client is stored
	assert.Nil(t, service.Bot)
	assert.Equal(t, mockConfig, service.Config)
}

func TestNotificationService_NilConfig(t *testing.T) {
	mockClient := &tg.Client{}
	mockLogsWriter := &MockLogsWriter{}

	// Test with nil config - should not panic during creation
	service := NewNotification(mockClient, nil, mockLogsWriter)
	assert.NotNil(t, service)

	// Cast to concrete type to verify nil config is stored
	assert.Equal(t, mockClient, service.Bot)
	assert.Nil(t, service.Config)
}
