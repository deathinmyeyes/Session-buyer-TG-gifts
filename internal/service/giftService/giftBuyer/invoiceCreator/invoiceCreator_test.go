package invoiceCreator

import (
	"testing"

	"gift-buyer/internal/service/giftService/giftTypes"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserCache для тестирования
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

func createTestGift(id int64, stars int64) *tg.StarGift {
	return &tg.StarGift{
		ID:    id,
		Stars: stars,
	}
}

func createTestGiftRequire(gift *tg.StarGift, receiverType []int) *giftTypes.GiftRequire {
	return &giftTypes.GiftRequire{
		Gift:         gift,
		ReceiverType: receiverType,
		CountForBuy:  1,
		Hide:         true,
	}
}

func TestNewInvoiceCreator(t *testing.T) {
	mockCache := &MockUserCache{}
	userReceiver := []string{"123456"}
	channelReceiver := []string{"789012"}

	creator := NewInvoiceCreator(userReceiver, channelReceiver, mockCache)

	assert.NotNil(t, creator)
	assert.Equal(t, userReceiver, creator.userReceiver)
	assert.Equal(t, channelReceiver, creator.channelReceiver)
	assert.Equal(t, mockCache, creator.idCache)
}

func TestInvoiceCreatorImpl_CreateInvoice(t *testing.T) {
	t.Run("создание инвойса", func(t *testing.T) {
		mockCache := &MockUserCache{}

		creator := NewInvoiceCreator(
			[]string{"123456"},
			[]string{"789012"},
			mockCache,
		)

		gift := createTestGift(1, 100)
		giftRequire := createTestGiftRequire(gift, []int{0})

		// Тестируем создание инвойса для self (type 0)
		invoice, err := creator.CreateInvoice(giftRequire)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer установлен
		assert.NotNil(t, invoice.Peer)
	})
}

func TestInvoiceCreatorImpl_SelfPurchase(t *testing.T) {
	t.Run("создание инвойса для себя", func(t *testing.T) {
		mockCache := &MockUserCache{}
		creator := NewInvoiceCreator([]string{}, []string{}, mockCache)

		gift := createTestGift(1, 100)
		giftRequire := createTestGiftRequire(gift, []int{0})
		invoice, err := creator.selfPurchase(giftRequire)

		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, gift.ID, invoice.GiftID)
		assert.True(t, invoice.HideName)
		assert.NotEmpty(t, invoice.Message.Text)

		// Проверяем что peer это InputPeerSelf
		_, ok := invoice.Peer.(*tg.InputPeerSelf)
		assert.True(t, ok)
	})
}
