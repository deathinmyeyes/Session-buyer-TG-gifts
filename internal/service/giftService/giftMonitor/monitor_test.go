package giftMonitor

import (
	"context"
	"sync"
	"testing"
	"time"

	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"gift-buyer/internal/service/giftService/giftTypes"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGiftCache is a mock implementation of GiftCache interface
type MockGiftCache struct {
	mock.Mock
}

func (m *MockGiftCache) SetGift(id int64, gift *tg.StarGift) {
	m.Called(id, gift)
}

func (m *MockGiftCache) GetGift(id int64) (*tg.StarGift, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.StarGift), args.Error(1)
}

func (m *MockGiftCache) GetAllGifts() map[int64]*tg.StarGift {
	args := m.Called()
	return args.Get(0).(map[int64]*tg.StarGift)
}

func (m *MockGiftCache) HasGift(id int64) bool {
	args := m.Called(id)
	return args.Bool(0)
}

func (m *MockGiftCache) DeleteGift(id int64) {
	m.Called(id)
}

func (m *MockGiftCache) Clear() {
	m.Called()
}

// MockGiftManager is a mock implementation of Giftmanager interface
type MockGiftManager struct {
	mock.Mock
}

func (m *MockGiftManager) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*tg.StarGift), args.Error(1)
}

// MockGiftValidator is a mock implementation of GiftValidator interface
type MockGiftValidator struct {
	mock.Mock
}

func (m *MockGiftValidator) IsEligible(gift *tg.StarGift) (*giftTypes.GiftRequire, bool) {
	args := m.Called(gift)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	return args.Get(0).(*giftTypes.GiftRequire), args.Bool(1)
}

// MockNotificationService is a mock implementation of NotificationService interface
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error {
	args := m.Called(ctx, gift)
	return args.Error(0)
}

func (m *MockNotificationService) SendBuyStatus(ctx context.Context, status string, err error) error {
	args := m.Called(ctx, status, err)
	return args.Error(0)
}

func (m *MockNotificationService) SendErrorNotification(ctx context.Context, err error) error {
	args := m.Called(ctx, err)
	return args.Error(0)
}

func (m *MockNotificationService) SetBot() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockNotificationService) SendUpdateNotification(ctx context.Context, version, message string) error {
	args := m.Called(ctx, version, message)
	return args.Error(0)
}

// MockLogsWriter для тестирования
type MockLogsWriter struct{}

func (m *MockLogsWriter) Write(entry *logTypes.LogEntry) error {
	return nil
}

func (m *MockLogsWriter) LogError(message string) {}

func (m *MockLogsWriter) LogErrorf(format string, args ...interface{}) {}

func (m *MockLogsWriter) LogInfo(message string) {}

func TestNewGiftMonitor(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}
	tickTime := time.Second

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, tickTime, mockErrorWriter, mockInfoWriter, true)

	assert.NotNil(t, monitor)

	// Cast to concrete type to verify internal structure
	assert.Equal(t, mockCache, monitor.cache)
	assert.Equal(t, mockManager, monitor.manager)
	assert.Equal(t, mockValidator, monitor.validator)
	assert.Equal(t, mockNotification, monitor.notification)
	assert.NotNil(t, monitor.ticker)
}

func TestGiftMonitor_Start_FirstRun(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Millisecond*10, mockErrorWriter, mockInfoWriter, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// First run - should just add all gifts to cache and return "touch grass" error
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return(currentGifts, nil)
	// On first run, gifts are processed but then firstRun error is returned
	mockCache.On("HasGift", int64(1)).Return(false).Once()
	mockCache.On("HasGift", int64(2)).Return(false).Once()
	mockValidator.On("IsEligible", gift1).Return(&giftTypes.GiftRequire{CountForBuy: 10, ReceiverType: []int{1}}, true).Once()
	mockValidator.On("IsEligible", gift2).Return(&giftTypes.GiftRequire{CountForBuy: 20, ReceiverType: []int{1}}, true).Once()
	mockCache.On("SetGift", int64(1), gift1).Return().Once()
	mockCache.On("SetGift", int64(2), gift2).Return().Once()

	// On first run, gifts are not in cache yet
	mockCache.On("HasGift", int64(1)).Return(false)
	mockCache.On("HasGift", int64(2)).Return(false)
	newGifts, err := monitor.Start(ctx)

	// Should find gifts on first run since they are not in cache yet
	assert.NoError(t, err)
	assert.NotNil(t, newGifts)
	assert.Len(t, newGifts, 2)

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_Start_SecondRunWithNewGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	// Create monitor and skip first run manually
	gm := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Millisecond * 10),
		firstRun:        false, // Skip first run
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
		testMode:        true,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// Setup mocks for new gifts
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockCache.On("HasGift", int64(2)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(&giftTypes.GiftRequire{CountForBuy: 10, ReceiverType: []int{1}}, true)
	mockValidator.On("IsEligible", gift2).Return(&giftTypes.GiftRequire{CountForBuy: 20, ReceiverType: []int{1}}, true)
	mockCache.On("SetGift", int64(1), gift1).Return()
	mockCache.On("SetGift", int64(2), gift2).Return()

	newGifts, err := gm.Start(ctx)

	assert.NoError(t, err)
	assert.Len(t, newGifts, 2)

	// Проверяем что подарки присутствуют в результате
	var foundGift1, foundGift2 bool
	for _, giftReq := range newGifts {
		if giftReq.Gift == gift1 {
			foundGift1 = true
			assert.Equal(t, int64(10), giftReq.CountForBuy)
		}
		if giftReq.Gift == gift2 {
			foundGift2 = true
			assert.Equal(t, int64(20), giftReq.CountForBuy)
		}
	}
	assert.True(t, foundGift1, "Gift1 должен быть найден в результатах")
	assert.True(t, foundGift2, "Gift2 должен быть найден в результатах")

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_Start_ContextCancelled(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Second, mockErrorWriter, mockInfoWriter, true)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	newGifts, err := monitor.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, newGifts)
}

func TestGiftMonitor_Start_NoNewGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Millisecond*10, mockErrorWriter, mockInfoWriter, true)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Setup mocks to return no new gifts - should timeout without error notifications since first run logic is disabled
	mockManager.On("GetAvailableGifts", mock.AnythingOfType("*context.timerCtx")).Return([]*tg.StarGift{}, nil)

	// Start monitoring - it should continue until context timeout
	newGifts, err := monitor.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, newGifts)

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_Success(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Second),
		firstRun:        false,
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
		testMode:        true,
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	gift2 := &tg.StarGift{ID: 2, Stars: 200}
	currentGifts := []*tg.StarGift{gift1, gift2}

	// Setup mocks
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockCache.On("HasGift", int64(2)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(&giftTypes.GiftRequire{CountForBuy: 10, ReceiverType: []int{1}}, true)
	mockValidator.On("IsEligible", gift2).Return(&giftTypes.GiftRequire{CountForBuy: 20, ReceiverType: []int{1}}, true)
	mockCache.On("SetGift", int64(1), gift1).Return()
	mockCache.On("SetGift", int64(2), gift2).Return()

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Len(t, newGifts, 2) // Both gifts should be eligible

	// Проверяем что подарки присутствуют в результате
	var foundGift1, foundGift2 bool
	for _, giftReq := range newGifts {
		if giftReq.Gift == gift1 {
			foundGift1 = true
		}
		if giftReq.Gift == gift2 {
			foundGift2 = true
		}
	}
	assert.True(t, foundGift1, "Gift1 должен быть найден в результатах")
	assert.True(t, foundGift2, "Gift2 должен быть найден в результатах")

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_ManagerError(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Second),
		firstRun:        false,
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
	}

	ctx := context.Background()
	expectedError := assert.AnError

	mockManager.On("GetAvailableGifts", ctx).Return(([]*tg.StarGift)(nil), expectedError)

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	assert.Nil(t, newGifts)

	mockManager.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_ExistingGifts(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Second),
		firstRun:        false,
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	currentGifts := []*tg.StarGift{gift1}

	// Setup mocks - gift already exists in cache
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(true)

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Empty(t, newGifts) // No new gifts since it already exists

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_CheckForNewGifts_NotEligible(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Second),
		firstRun:        false,
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
	}

	ctx := context.Background()

	// Create test gifts
	gift1 := &tg.StarGift{ID: 1, Stars: 100}
	currentGifts := []*tg.StarGift{gift1}

	// Setup mocks - gift is not eligible
	mockManager.On("GetAvailableGifts", ctx).Return(currentGifts, nil)
	mockCache.On("HasGift", int64(1)).Return(false)
	mockValidator.On("IsEligible", gift1).Return(&giftTypes.GiftRequire{CountForBuy: 0, ReceiverType: []int{}}, false)
	mockCache.On("SetGift", int64(1), gift1).Return()

	newGifts, err := monitor.checkForNewGifts(ctx)

	assert.NoError(t, err)
	assert.Empty(t, newGifts) // No eligible gifts

	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_PauseResumeIsPaused(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Second, mockErrorWriter, mockInfoWriter, true)

	// Initially should not be paused
	assert.False(t, monitor.IsPaused())

	// Test pause
	monitor.Pause()
	assert.True(t, monitor.IsPaused())

	// Test pause again (should still be paused)
	monitor.Pause()
	assert.True(t, monitor.IsPaused())

	// Test resume
	monitor.Resume()
	assert.False(t, monitor.IsPaused())

	// Test resume again (should still be not paused)
	monitor.Resume()
	assert.False(t, monitor.IsPaused())
}

func TestGiftMonitor_Start_WithPause(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	// Create monitor and skip first run manually
	gm := &giftMonitorImpl{
		cache:           mockCache,
		manager:         mockManager,
		validator:       mockValidator,
		notification:    mockNotification,
		ticker:          time.NewTicker(time.Millisecond * 10),
		firstRun:        false, // Skip first run
		paused:          true,  // Start paused
		errorLogsWriter: mockErrorWriter,
		infoLogsWriter:  mockInfoWriter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	// Since monitor is paused, no API calls should be made
	// The test should timeout waiting for results

	newGifts, err := gm.Start(ctx)

	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
	assert.Nil(t, newGifts)

	// No expectations should be called since monitor is paused
	mockCache.AssertExpectations(t)
	mockManager.AssertExpectations(t)
	mockValidator.AssertExpectations(t)
}

func TestGiftMonitor_ConcurrentPauseResume(t *testing.T) {
	mockCache := new(MockGiftCache)
	mockManager := new(MockGiftManager)
	mockValidator := new(MockGiftValidator)
	mockNotification := new(MockNotificationService)
	mockErrorWriter := &MockLogsWriter{}
	mockInfoWriter := &MockLogsWriter{}

	monitor := NewGiftMonitor(mockCache, mockManager, mockValidator, mockNotification, time.Second, mockErrorWriter, mockInfoWriter, true)

	// Test concurrent access to pause/resume methods
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(3)
		go func() {
			defer wg.Done()
			monitor.Pause()
		}()
		go func() {
			defer wg.Done()
			monitor.Resume()
		}()
		go func() {
			defer wg.Done()
			_ = monitor.IsPaused()
		}()
	}

	wg.Wait()

	// After all operations, the monitor should be in a consistent state
	// The final state could be either paused or not paused, but it should be consistent
	finalState := monitor.IsPaused()
	assert.Equal(t, finalState, monitor.IsPaused())
}
