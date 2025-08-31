package invoiceCreator

import (
	"context"
	"fmt"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/internal/service/giftService/giftTypes"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/utils"
	"time"

	"github.com/google/uuid"
	"github.com/gotd/td/tg"
)

type InvoiceCreatorImpl struct {
	userReceiver, channelReceiver []string
	idCache                       giftInterfaces.UserCache
}

func NewInvoiceCreator(userReceiver, channelReceiver []string, idCache giftInterfaces.UserCache) *InvoiceCreatorImpl {
	return &InvoiceCreatorImpl{
		userReceiver:    userReceiver,
		channelReceiver: channelReceiver,
		idCache:         idCache,
	}
}

// createInvoice creates a Telegram invoice for the specified gift.
// It configures the invoice based on the receiver type (self, user, or channel)
// and includes appropriate peer information and gift details.
//
// Supported receiver types:
//   - 0: Self (current user)
//   - 1: User (specified by user ID)
//   - 2: Channel (specified by channel ID with access hash)
//
// Parameters:
//   - gift: the star gift to create an invoice for
//
// Returns:
//   - *tg.InputInvoiceStarGift: configured invoice for the gift purchase
//   - error: invoice creation error or unsupported receiver type
func (ic *InvoiceCreatorImpl) CreateInvoice(gift *giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error) {
	randReceiverType := utils.SelectRandomElementFast(gift.ReceiverType)

	switch randReceiverType {
	case 0:
		return ic.selfPurchase(gift)
	case 1:
		return ic.userPurchase(gift)
	case 2:
		return ic.channelPurchase(gift)
	default:
		return nil, errors.Wrap(errors.New("unexpected receiver type"),
			fmt.Sprintf("unexpected receiver type: %d", randReceiverType))
	}
}

func (ic *InvoiceCreatorImpl) selfPurchase(gift *giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error) {
	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerSelf{},
		GiftID:   gift.Gift.ID,
		HideName: gift.Hide,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @earnfame %s_%d_%s", utils.RandString5(10), time.Now().UnixNano(), uuid.New().String()[:6]),
		},
	}
	return invoice, nil
}

func (ic *InvoiceCreatorImpl) userPurchase(gift *giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error) {
	userInfo, err := ic.getUserInfo(context.Background(), utils.SelectRandomElementFast(ic.userReceiver))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without user access hash")
	}

	invoice := &tg.InputInvoiceStarGift{
		Peer:     &tg.InputPeerUser{UserID: userInfo.ID, AccessHash: userInfo.AccessHash},
		GiftID:   gift.Gift.ID,
		HideName: gift.Hide,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @earnfame %s_%d_%s", utils.RandString5(10), time.Now().UnixNano(), uuid.New().String()[:6]),
		},
	}
	return invoice, nil
}

func (ic *InvoiceCreatorImpl) channelPurchase(gift *giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error) {
	channelInfo, err := ic.getChannelInfo(context.Background(), utils.SelectRandomElementFast(ic.channelReceiver))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create invoice without channel access hash")
	}

	invoice := &tg.InputInvoiceStarGift{
		Peer: &tg.InputPeerChannel{
			ChannelID:  ic.convertChannelID(channelInfo.ID),
			AccessHash: channelInfo.AccessHash,
		},
		GiftID:   gift.Gift.ID,
		HideName: gift.Hide,
		Message: tg.TextWithEntities{
			Text: fmt.Sprintf("By @earnfame %s_%d", utils.RandString5(10), time.Now().UnixNano()),
		},
	}
	return invoice, nil
}

// getChannelInfo retrieves channel information including access hash for invoice creation.
// It handles channel ID conversion and fetches the channel details required for
// creating invoices for channel recipients.
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - channelID: the channel ID (may be in supergroup format)
//
// Returns:
//   - *tg.Channel: channel information with access hash
//   - error: channel retrieval error or API communication failure
func (ic *InvoiceCreatorImpl) getChannelInfo(ctx context.Context, channelID string) (*tg.Channel, error) {
	channel, err := ic.idCache.GetChannel(channelID)
	if err == nil {
		return channel, nil
	}

	return nil, errors.New("channel not found")
}

// getUserInfo retrieves user information including access hash for invoice creation.
// It tries multiple methods to get user info without requiring contacts:
// 1. Direct UsersGetUsers call
// 2. Search through recent dialogs with larger limit
// 3. Try to get user through common groups/channels
// 4. Search for messages from user
// 5. Try to resolve by username if available
// 6. Search through all chats and channels
//
// Parameters:
//   - ctx: context for request cancellation and timeout control
//   - userID: the user ID
//
// Returns:
//   - *tg.User: user information with access hash
//   - error: user retrieval error or API communication failure
func (ic *InvoiceCreatorImpl) getUserInfo(ctx context.Context, userID string) (*tg.User, error) {
	user, err := ic.idCache.GetUser(userID)
	if err == nil {
		return user, nil
	}

	return nil, errors.New(fmt.Sprintf("user %s not accessible: session hasn't met this user. See logs for solutions.", userID))
}

func (ic *InvoiceCreatorImpl) convertChannelID(channelID int64) int64 {
	var realChannelID int64
	if channelID < -1000000000000 {
		realChannelID = -channelID - 1000000000000
	} else {
		realChannelID = channelID
	}
	return realChannelID
}
