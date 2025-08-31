package giftBuyer

import (
	"context"
	"gift-buyer/internal/infrastructure/logsWriter/logTypes"
	"gift-buyer/internal/service/giftService/giftBuyer/atomicCounter"
	"gift-buyer/internal/service/giftService/giftTypes"
	"sync"
	"testing"
	"time"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockGiftManager struct {
	mock.Mock
}

func (m *MockGiftManager) GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*tg.StarGift), args.Error(1)
}

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

type MockUserCache struct {
	mock.Mock
}

func (m *MockUserCache) SetUser(key string, user *tg.User) {
	m.Called(key, user)
}

func (m *MockUserCache) GetUser(id string) (*tg.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.User), args.Error(1)
}

func (m *MockUserCache) SetChannel(key string, channel *tg.Channel) {
	m.Called(key, channel)
}

func (m *MockUserCache) GetChannel(id string) (*tg.Channel, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.Channel), args.Error(1)
}

type MockRateLimiter struct {
	mock.Mock
	mu     sync.Mutex
	closed bool
}

func (m *MockRateLimiter) Acquire(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRateLimiter) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.closed {
		m.closed = true
		m.Called()
	}
}

type MockInvoiceCreator struct {
	mock.Mock
}

func (m *MockInvoiceCreator) CreateInvoice(gift *giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error) {
	args := m.Called(gift)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tg.InputInvoiceStarGift), args.Error(1)
}

type MockPurchaseProcessor struct {
	mock.Mock
}

func (m *MockPurchaseProcessor) PurchaseGift(ctx context.Context, gift *giftTypes.GiftRequire) error {
	args := m.Called(ctx, gift)
	return args.Error(0)
}

type MockMonitorProcessor struct {
	mock.Mock
}

func (m *MockMonitorProcessor) MonitorProcess(ctx context.Context, resultsCh chan giftTypes.GiftResult, doneCh chan struct{}, gifts []*giftTypes.GiftRequire) {
	m.Called(ctx, resultsCh, doneCh, gifts)
}

// Helper functions
func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func createMockBuyer() (*giftBuyerImpl, *MockGiftManager, *MockNotificationService, *MockUserCache, *MockRateLimiter, *MockInvoiceCreator, *MockPurchaseProcessor, *MockMonitorProcessor) {
	mockManager := &MockGiftManager{}
	mockNotification := &MockNotificationService{}
	mockUserCache := &MockUserCache{}
	mockRateLimiter := &MockRateLimiter{}
	mockInvoiceCreator := &MockInvoiceCreator{}
	mockPurchaseProcessor := &MockPurchaseProcessor{}
	mockMonitorProcessor := &MockMonitorProcessor{}

	buyer := &giftBuyerImpl{
		manager:              mockManager,
		idCache:              mockUserCache,
		notification:         mockNotification,
		api:                  nil, // nil API для тестирования
		userReceiver:         []string{"123456"},
		channelReceiver:      []string{"789012"},
		counter:              atomicCounter.NewAtomicCounter(100),
		retryCount:           3,
		retryDelay:           1.0,
		concurrentGifts:      5,
		concurrentOperations: 10,
		requestCounter:       0,
		rateLimiter:          mockRateLimiter,
		invoiceCreator:       mockInvoiceCreator,
		purchaseProcessor:    mockPurchaseProcessor,
		monitorProcessor:     mockMonitorProcessor,
		errorLogsWriter:      &MockLogsWriter{},
	}

	return buyer, mockManager, mockNotification, mockUserCache, mockRateLimiter, mockInvoiceCreator, mockPurchaseProcessor, mockMonitorProcessor
}

func TestNewGiftBuyer(t *testing.T) {
	t.Run("создание нового GiftBuyer", func(t *testing.T) {
		mockManager := &MockGiftManager{}
		mockNotification := &MockNotificationService{}
		mockUserCache := &MockUserCache{}
		mockRateLimiter := &MockRateLimiter{}
		mockInvoiceCreator := &MockInvoiceCreator{}
		mockPurchaseProcessor := &MockPurchaseProcessor{}
		mockMonitorProcessor := &MockMonitorProcessor{}
		mockCounter := atomicCounter.NewAtomicCounter(100)
		mockLogsWriter := &MockLogsWriter{}

		buyer := NewGiftBuyer(
			nil, // api
			[]string{"123456"},
			[]string{"789012"},
			mockManager,
			mockNotification,
			100,  // maxBuyCount
			3,    // retryCount
			1.0,  // retryDelay
			true, // prioritization
			mockUserCache,
			5, // concurrentGifts
			mockRateLimiter,
			10, // concurrentOperations
			mockInvoiceCreator,
			mockPurchaseProcessor,
			mockMonitorProcessor,
			mockCounter,
			mockLogsWriter,
		)

		assert.NotNil(t, buyer)
		assert.Equal(t, []string{"123456"}, buyer.userReceiver)
		assert.Equal(t, []string{"789012"}, buyer.channelReceiver)
		assert.Equal(t, 3, buyer.retryCount)
		assert.Equal(t, 5, buyer.concurrentGifts)
		assert.Equal(t, 10, buyer.concurrentOperations)
		assert.NotNil(t, buyer.counter)
		assert.NotNil(t, buyer.invoiceCreator)
		assert.NotNil(t, buyer.purchaseProcessor)
		assert.NotNil(t, buyer.monitorProcessor)
	})
}

func TestGiftBuyerImpl_BuyGift(t *testing.T) {
	t.Run("успешная покупка подарков", func(t *testing.T) {
		buyer, _, _, _, _, _, mockPurchaseProcessor, mockMonitorProcessor := createMockBuyer()

		// Настраиваем моки
		mockMonitorProcessor.On("MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockPurchaseProcessor.On("PurchaseGift", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		gifts := []*giftTypes.GiftRequire{
			{Gift: createTestGift(1, 100), CountForBuy: 2, ReceiverType: []int{1}},
			{Gift: createTestGift(2, 200), CountForBuy: 1, ReceiverType: []int{1}},
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(3 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Проверяем что мониторинг был запущен
		mockMonitorProcessor.AssertCalled(t, "MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("покупка с пустым списком подарков", func(t *testing.T) {
		buyer, _, _, _, _, _, _, mockMonitorProcessor := createMockBuyer()

		mockMonitorProcessor.On("MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		gifts := []*giftTypes.GiftRequire{}

		ctx := context.Background()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(1 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Мониторинг должен быть запущен даже для пустого списка
		mockMonitorProcessor.AssertCalled(t, "MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})

	t.Run("покупка с отменой контекста", func(t *testing.T) {
		buyer, _, _, _, _, _, _, mockMonitorProcessor := createMockBuyer()

		mockMonitorProcessor.On("MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()

		gifts := []*giftTypes.GiftRequire{
			{Gift: createTestGift(1, 100), CountForBuy: 1, ReceiverType: []int{1}},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Отменяем контекст сразу

		buyer.BuyGift(ctx, gifts)

		// Тест должен завершиться без паники
	})
}

func TestGiftBuyerImpl_Close(t *testing.T) {
	t.Run("закрытие buyer", func(t *testing.T) {
		buyer, _, _, _, mockRateLimiter, _, _, _ := createMockBuyer()

		mockRateLimiter.On("Close").Return()

		buyer.Close()

		mockRateLimiter.AssertCalled(t, "Close")
	})
}

func TestGiftBuyerImpl_ConcurrentPurchases(t *testing.T) {
	t.Run("конкурентные покупки", func(t *testing.T) {
		buyer, _, _, _, _, _, mockPurchaseProcessor, mockMonitorProcessor := createMockBuyer()

		mockMonitorProcessor.On("MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockPurchaseProcessor.On("PurchaseGift", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Создаем много подарков для тестирования конкурентности
		gifts := make([]*giftTypes.GiftRequire, 0, 10)
		for i := int64(1); i <= 10; i++ {
			gifts = append(gifts, &giftTypes.GiftRequire{Gift: createTestGift(i, 100), CountForBuy: 2, ReceiverType: []int{1}})
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		start := time.Now()
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			elapsed := time.Since(start)
			// Конкурентные покупки должны завершиться быстро
			assert.True(t, elapsed < 5*time.Second, "elapsed time: %v", elapsed)
		case <-time.After(8 * time.Second):
			t.Fatal("Concurrent purchases took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		mockMonitorProcessor.AssertCalled(t, "MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

func TestGiftBuyerImpl_MaxBuyCountLimit(t *testing.T) {
	t.Run("ограничение максимального количества покупок", func(t *testing.T) {
		buyer, _, _, _, _, _, mockPurchaseProcessor, mockMonitorProcessor := createMockBuyer()
		buyer.counter = atomicCounter.NewAtomicCounter(2) // Ограничиваем до 2 покупок

		mockMonitorProcessor.On("MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return()
		mockPurchaseProcessor.On("PurchaseGift", mock.Anything, mock.Anything, mock.Anything).Return(nil)

		// Пытаемся купить больше чем лимит
		gifts := []*giftTypes.GiftRequire{
			{Gift: createTestGift(1, 100), CountForBuy: 5, ReceiverType: []int{1}}, // Пытаемся купить 5, но лимит 2
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Создаем канал для ожидания завершения
		done := make(chan struct{})
		go func() {
			buyer.BuyGift(ctx, gifts)
			close(done)
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Операция завершена
		case <-time.After(3 * time.Second):
			t.Fatal("BuyGift took too long to complete")
		}

		// Даем время для завершения асинхронных операций
		time.Sleep(100 * time.Millisecond)

		// Проверяем что счетчик не превысил лимит
		assert.True(t, buyer.counter.Get() <= buyer.counter.GetMax())

		mockMonitorProcessor.AssertCalled(t, "MonitorProcess", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	})
}

type MockLogsWriter struct{}

func (m *MockLogsWriter) Write(entry *logTypes.LogEntry) error {
	return nil
}

func (m *MockLogsWriter) LogError(message string) {}

func (m *MockLogsWriter) LogErrorf(format string, args ...interface{}) {}

func (m *MockLogsWriter) LogInfo(message string) {}
