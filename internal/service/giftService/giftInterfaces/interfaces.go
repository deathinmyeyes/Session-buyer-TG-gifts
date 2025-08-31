// Package giftInterfaces defines the core interfaces for the gift buying system.
// These interfaces provide abstractions for gift management, validation, purchasing,
// caching, monitoring, and notifications, enabling loose coupling and testability.
package giftInterfaces

import (
	"context"
	"gift-buyer/internal/service/giftService/giftTypes"

	"github.com/gotd/td/tg"
)

// Giftmanager defines the interface for managing gift operations with Telegram API.
// It provides methods to retrieve available gifts from the Telegram platform.
type Giftmanager interface {
	// GetAvailableGifts retrieves all currently available star gifts from Telegram.
	// It returns a slice of StarGift objects representing gifts that can be purchased.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//
	// Returns:
	//   - []*tg.StarGift: slice of available star gifts
	//   - error: API communication error or parsing error
	GetAvailableGifts(ctx context.Context) ([]*tg.StarGift, error)
}

// GiftValidator defines the interface for validating gifts against purchase criteria.
// It determines whether a gift meets the configured requirements for automatic purchase.
type GiftValidator interface {
	// IsEligible checks if a gift meets the configured purchase criteria.
	// It evaluates the gift against price ranges, supply limits, and other constraints.
	//
	// Parameters:
	//   - gift: the star gift to validate
	//
	// Returns:
	//   - int64: number of gifts to purchase if eligible (0 if not eligible)
	//   - bool: true if the gift meets criteria, false otherwise
	IsEligible(gift *tg.StarGift) (*giftTypes.GiftRequire, bool)
}

// GiftBuyer defines the interface for purchasing gifts through Telegram API.
// It handles the actual purchase transactions and manages purchase limits.
type GiftBuyer interface {
	// BuyGift attempts to purchase the specified gifts with their respective quantities.
	// It handles payment processing, retry logic, and purchase confirmation.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gifts: map of gifts to their desired purchase quantities
	//
	// Returns:
	//   - error: purchase error, payment failure, or API communication error
	BuyGift(ctx context.Context, gifts []*giftTypes.GiftRequire)

	// Close releases any resources held by the gift buyer (e.g., rate limiter).
	Close()
}

// GiftCache defines the interface for caching gift information.
// It provides persistent storage for gift data to avoid redundant API calls
// and maintain state across application restarts.
type GiftCache interface {
	// SetGift stores a gift in the cache with the specified ID as key.
	//
	// Parameters:
	//   - id: unique identifier for the gift
	//   - gift: the star gift object to cache
	SetGift(id int64, gift *tg.StarGift)

	// GetGift retrieves a cached gift by its ID.
	//
	// Parameters:
	//   - id: unique identifier of the gift to retrieve
	//
	// Returns:
	//   - *tg.StarGift: the cached gift object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetGift(id int64) (*tg.StarGift, error)

	// GetAllGifts returns a copy of all cached gifts.
	//
	// Returns:
	//   - map[int64]*tg.StarGift: map of gift IDs to gift objects
	GetAllGifts() map[int64]*tg.StarGift

	// HasGift checks if a gift with the specified ID exists in the cache.
	//
	// Parameters:
	//   - id: unique identifier of the gift to check
	//
	// Returns:
	//   - bool: true if the gift exists in cache, false otherwise
	HasGift(id int64) bool

	// DeleteGift removes a gift from the cache.
	//
	// Parameters:
	//   - id: unique identifier of the gift to remove
	DeleteGift(id int64)

	// Clear removes all gifts from the cache.
	Clear()
}

// GiftMonitor defines the interface for monitoring new gifts.
// It continuously checks for new gifts and identifies those eligible for purchase.
type GiftMonitor interface {
	// Start begins the gift monitoring process and returns newly discovered eligible gifts.
	// It runs continuously until the context is cancelled, checking for new gifts
	// at configured intervals.
	//
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - []*giftTypes.GiftRequire: slice of eligible gifts with their purchase requirements
	//   - error: monitoring error, API communication error, or context cancellation
	Start(ctx context.Context) ([]*giftTypes.GiftRequire, error)

	// Pause pauses the gift monitoring process.
	Pause()

	// Resume resumes the gift monitoring process.
	Resume()

	// IsPaused returns the status of the gift monitoring process.
	//
	// Returns:
	//   - bool: true if the monitoring is paused, false if active
	IsPaused() bool
}

// NotificationService defines the interface for sending notifications.
// It provides methods to notify users about new gifts and purchase status updates.
type NotificationService interface {
	// SendNewGiftNotification sends a notification about a newly discovered gift.
	// The notification includes gift details such as price, supply, and availability.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gift: the star gift to notify about
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendNewGiftNotification(ctx context.Context, gift *tg.StarGift) error

	// SendBuyStatus sends a notification about the purchase operation status.
	// It reports successful purchases or error conditions to the configured chat.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - status: human-readable status message
	//   - err: error that occurred during purchase (nil for success)
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendBuyStatus(ctx context.Context, status string, err error) error

	// SendErrorNotification sends a notification about an error.
	// It reports error conditions to the configured chat.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - err: error to notify about
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendErrorNotification(ctx context.Context, err error) error

	// SetBot sets the bot client
	SetBot() bool

	// SendUpdateNotification sends a notification about an update.
	// It reports update conditions to the configured chat.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - version: the version to notify about
	//
	// Returns:
	//   - error: notification sending error or API communication error
	SendUpdateNotification(ctx context.Context, version, message string) error
}

// UserCache defines the interface for caching user and channel information.
// It provides persistent storage for user data to avoid redundant API calls
// and maintain state across application restarts.
type UserCache interface {
	// SetUser stores a user in the cache with the specified key.
	//
	// Parameters:
	//   - key: unique identifier for the user (username or ID)
	//   - user: the user object to cache
	SetUser(key string, user *tg.User)

	// GetUser retrieves a cached user by their key.
	//
	// Parameters:
	//   - key: unique identifier of the user to retrieve
	//
	// Returns:
	//   - *tg.User: the cached user object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetUser(key string) (*tg.User, error)

	// SetChannel stores a channel in the cache with the specified key.
	//
	// Parameters:
	//   - key: unique identifier for the channel (username or ID)
	//   - channel: the channel object to cache
	SetChannel(key string, channel *tg.Channel)

	// GetChannel retrieves a cached channel by its key.
	//
	// Parameters:
	//   - key: unique identifier of the channel to retrieve
	//
	// Returns:
	//   - *tg.Channel: the cached channel object, nil if not found
	//   - error: retrieval error (currently always nil)
	GetChannel(key string) (*tg.Channel, error)
}

// RateLimiter defines the interface for rate limiting API calls.
// It provides methods to acquire and release tokens for API requests.
type RateLimiter interface {
	// Acquire acquires a token from the rate limiter.
	// It blocks until a token is available or the context is cancelled.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//
	// Returns:
	Acquire(ctx context.Context) error

	// Close releases any resources held by the rate limiter.
	// It should be called when the rate limiter is no longer needed.
	Close()
}

type AccountManager interface {
	SetIds(ctx context.Context) error
}

// InvoiceCreator defines the interface for creating invoices for gift purchases.
// It provides methods to generate appropriate invoices based on the receiver type
// and include necessary peer information and gift details.
type InvoiceCreator interface {
	// CreateInvoice creates a Telegram invoice for the specified gift.
	// It configures the invoice based on the receiver type (self, user, or channel)
	// and includes appropriate peer information and gift details.
	//
	// Parameters:
	//   - gift: the star gift to create an invoice for
	//
	CreateInvoice(*giftTypes.GiftRequire) (*tg.InputInvoiceStarGift, error)
}

// PaymentProcessor defines the interface for processing purchases.
// It provides methods to create payment forms and validate purchases.
type PaymentProcessor interface {
	// CreatePaymentForm creates a payment form for the specified gift.
	// It generates the appropriate payment form based on the gift type and includes necessary details.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gift: the star gift to create a payment form for
	//
	// Returns:
	//   - tg.PaymentsPaymentFormClass: the payment form object
	//   - *tg.InputInvoiceStarGift: the invoice object
	//   - error: payment form creation error or API communication failure
	CreatePaymentForm(ctx context.Context, gift *giftTypes.GiftRequire) (tg.PaymentsPaymentFormClass, *tg.InputInvoiceStarGift, error)
}

// PurchaseProcessor defines the interface for processing purchases.
// It provides methods to purchase gifts and handle different payment form types.
type PurchaseProcessor interface {
	// PurchaseGift executes the actual gift purchase through Telegram's payment API.
	// It creates an invoice, retrieves the payment form, and processes the star payment.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - gift: the star gift to purchase
	//
	// Returns:
	//   - error: payment processing error or API communication failure
	PurchaseGift(ctx context.Context, gift *giftTypes.GiftRequire) error
}

// MonitorProcessor defines the interface for monitoring the purchase process.
// It provides methods to monitor the purchase process and send notifications.
type MonitorProcessor interface {
	// MonitorProcess monitors the purchase process and sends notifications.
	// It receives results from the purchase process and updates the summaries.
	//
	// Parameters:
	//   - ctx: context for request cancellation and timeout control
	//   - resultsCh: channel to receive purchase results
	//   - doneChan: channel to signal completion
	//   - gifts: map of gifts to their purchase quantities
	MonitorProcess(ctx context.Context, resultsCh chan giftTypes.GiftResult, doneChan chan struct{}, gifts []*giftTypes.GiftRequire)
}

// Counter defines the interface for managing a counter with atomic operations.
// It provides methods to increment, decrement, and retrieve the current count.
type Counter interface {
	// TryIncrement attempts to increment the counter by one.
	// It returns true if the increment was successful, false if the maximum count has been reached.
	TryIncrement() bool

	// Decrement decrements the counter by one.
	Decrement()

	// Get returns the current count value.
	//
	// Returns:
	//   - int64: current count value
	Get() int64

	// GetMax returns the maximum allowed count value.
	//
	// Returns:
	//   - int64: maximum count limit
	GetMax() int64
}

// ErrorLogger defines the interface for logging errors.
// It provides methods to log errors and formatted errors.
type ErrorLogger interface {
	// LogError logs an error message.
	//
	// Parameters:
	//   - message: the error message to log
	LogError(message string)

	// LogErrorf logs a formatted error message.
	// It formats the message using the provided format and arguments.
	//
	// Parameters:
	//   - format: the format string for the error message
	//   - args: arguments to be formatted into the message
	LogErrorf(format string, args ...interface{})
}

// InfoLogger defines the interface for logging information.
// It provides methods to log information messages.
type InfoLogger interface {
	// LogInfo logs an information message.
	//
	// Parameters:
	//   - message: the information message to log
	LogInfo(message string)
}
