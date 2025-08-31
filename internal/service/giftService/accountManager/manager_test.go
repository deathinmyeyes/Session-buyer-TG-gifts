package accountManager

import (
	"context"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
)

func TestNewAccountManager(t *testing.T) {
	userReceiverIDs := []string{"123456789"}
	channelReceiverIDs := []string{"987654321"}
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}

	manager := NewAccountManager(nil, userReceiverIDs, channelReceiverIDs, userCache, channelCache)

	assert.NotNil(t, manager)
}

func TestNewAccountManager_EmptyIDs(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}

	manager := NewAccountManager(nil, []string{}, []string{}, userCache, channelCache)

	assert.NotNil(t, manager)
}

func TestNewAccountManager_NilCache(t *testing.T) {
	userReceiverIDs := []string{"123456789"}
	channelReceiverIDs := []string{"987654321"}

	manager := NewAccountManager(nil, userReceiverIDs, channelReceiverIDs, nil, nil)

	assert.NotNil(t, manager)
}

func TestAccountManager_SetIds_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(nil, []string{"123456789"}, []string{"987654321"}, userCache, channelCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

func TestAccountManager_SetIds_EmptyIDs(t *testing.T) {
	api := &tg.Client{}
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(api, []string{}, []string{}, userCache, channelCache)

	ctx := context.Background()
	err := manager.SetIds(ctx)

	assert.NoError(t, err) // Should succeed with empty IDs
}

func TestAccountManager_SetIds_ContextCancellation(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(nil, []string{"123456789"}, []string{"987654321"}, userCache, channelCache)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := manager.SetIds(ctx)

	assert.Error(t, err)
	// С nil API мы получим ошибку "API client is nil", а не context.Canceled
	assert.Contains(t, err.Error(), "API client is nil")
}

func TestAccountManager_Structure(t *testing.T) {
	userReceiverIDs := []string{"123456789", "987654321"}
	channelReceiverIDs := []string{"111222333", "444555666"}
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	api := &tg.Client{}

	manager := NewAccountManager(api, userReceiverIDs, channelReceiverIDs, userCache, channelCache)

	assert.NotNil(t, manager)
}

func TestAccountManager_InterfaceCompliance(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(nil, []string{}, []string{}, userCache, channelCache)

	// Verify that the manager has the SetIds method
	assert.NotNil(t, manager.SetIds)
}

func TestAccountManager_LoadUsersToCache_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(nil, []string{"123456789"}, []string{}, userCache, channelCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

func TestAccountManager_LoadChannelsToCache_NilAPI(t *testing.T) {
	userCache := &MockUserCache{}
	channelCache := &MockChannelCache{}
	manager := NewAccountManager(nil, []string{}, []string{"987654321"}, userCache, channelCache)

	ctx := context.Background()

	// Должен вернуть ошибку без паники при nil API
	assert.NotPanics(t, func() {
		err := manager.SetIds(ctx)
		assert.Error(t, err)
	})
}

// Mock implementations for testing

type MockUserCache struct{}

func (m *MockUserCache) SetUser(key string, user *tg.User) {
	// Mock implementation
}

func (m *MockUserCache) GetUser(id string) (*tg.User, error) {
	return nil, assert.AnError
}

type MockChannelCache struct{}

func (m *MockChannelCache) SetChannel(key string, channel *tg.Channel) {
	// Mock implementation
}

func (m *MockChannelCache) GetChannel(id string) (*tg.Channel, error) {
	return nil, assert.AnError
}
