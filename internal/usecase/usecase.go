package usecase

import (
	"context"
	"fmt"
	"gift-buyer/internal/infrastructure/gitVersion/gitInterfaces"
	"gift-buyer/internal/service/giftService/giftInterfaces"
	"gift-buyer/pkg/logger"
	"sync"
	"time"

	"github.com/gotd/td/tg"
)

// UseCase defines the main interface for the gift buying service.
// It provides lifecycle management methods for starting and stopping the service.
type UseCase interface {
	// Start begins the gift monitoring and purchasing process.
	// This method runs continuously until stopped or context cancelled.
	Start()

	// Stop gracefully shuts down the gift service and all its components.
	Stop()

	// SetIds sets the IDs of the accounts
	SetIds(ctx context.Context) error

	// CheckForUpdates checks for updates and sends a notification if available
	CheckForUpdates()
}

// useCaseImpl implements the UseCase interface and orchestrates all gift buying operations.
// It manages the lifecycle of monitoring, validation, purchasing, and notification components,
// providing a unified service that automatically discovers and purchases eligible gifts.
type useCaseImpl struct {
	// manager handles gift retrieval and API communication
	manager giftInterfaces.Giftmanager

	// validator evaluates gifts against purchase criteria
	validator giftInterfaces.GiftValidator

	// cache provides persistent storage for processed gifts
	cache giftInterfaces.GiftCache

	// notification sends alerts about discoveries and purchase status
	notification giftInterfaces.NotificationService

	// monitor continuously checks for new eligible gifts
	monitor giftInterfaces.GiftMonitor

	// buyer handles the actual gift purchase transactions
	buyer giftInterfaces.GiftBuyer

	// ctx provides cancellation context for the service
	ctx context.Context

	// cancel function to stop the service gracefully
	cancel context.CancelFunc

	// wg coordinates graceful shutdown of goroutines
	wg sync.WaitGroup

	// api is the main Telegram client for API operations
	api *tg.Client

	// accountManager handles account-related operations
	accountManager giftInterfaces.AccountManager

	// version is the current version of the service
	gitVersion              gitInterfaces.GitVersionController
	updateTicker            *time.Ticker
	lastNotificationVersion string
	subFlag                 bool
}

// NewUseCase creates a new UseCase instance with all required dependencies.
// It wires together all components needed for automated gift buying operations.
//
// Parameters:
//   - manager: gift manager for API communication
//   - validator: gift validator for eligibility checking
//   - cache: gift cache for state persistence
//   - notification: notification service for alerts
//   - monitor: gift monitor for continuous discovery
//   - buyer: gift buyer for purchase operations
//   - ctx: context for cancellation control
//   - cancel: cancel function for graceful shutdown
//   - api: Telegram API client
//
// Returns:
//   - GiftService: configured gift service ready for operation
func NewUseCase(
	manager giftInterfaces.Giftmanager,
	validator giftInterfaces.GiftValidator,
	cache giftInterfaces.GiftCache,
	notification giftInterfaces.NotificationService,
	monitor giftInterfaces.GiftMonitor,
	buyer giftInterfaces.GiftBuyer,
	ctx context.Context,
	cancel context.CancelFunc,
	api *tg.Client,
	accountManager giftInterfaces.AccountManager,
	gitVersion gitInterfaces.GitVersionController,
	updateTicker *time.Ticker,
) UseCase {
	return &useCaseImpl{
		manager:        manager,
		validator:      validator,
		cache:          cache,
		notification:   notification,
		monitor:        monitor,
		buyer:          buyer,
		ctx:            ctx,
		cancel:         cancel,
		api:            api,
		accountManager: accountManager,
		gitVersion:     gitVersion,
		updateTicker:   updateTicker,
		subFlag:        false,
	}
}

// Start begins the main gift buying service loop.
// It continuously monitors for new gifts, validates them against criteria,
// and automatically purchases eligible gifts until the service is stopped.
//
// The service loop:
//  1. Monitors for new eligible gifts
//  2. Validates discovered gifts against criteria
//  3. Attempts to purchase eligible gifts
//  4. Handles errors and continues operation
//  5. Respects context cancellation for graceful shutdown
//
// This method blocks until the service is stopped or context is cancelled.
func (tc *useCaseImpl) Start() {
	for {
		select {
		case <-tc.ctx.Done():
			tc.wg.Wait()
			return
		default:
			newGifts, err := tc.monitor.Start(tc.ctx)
			if err != nil {
				if tc.ctx.Err() != nil {
					logger.GlobalLogger.Info("Context cancelled, stopping service")
					tc.wg.Wait()
					return
				}
				logger.GlobalLogger.Error("Error checking for new gifts", "error", err)
				continue
			}

			if len(newGifts) > 0 {
				logger.GlobalLogger.Infof("Found %d new gift types to process", len(newGifts))
				tc.wg.Add(2)
				go func() {
					defer tc.wg.Done()
					for _, require := range newGifts {
						if err := tc.notification.SendNewGiftNotification(tc.ctx, require.Gift); err != nil {
							logger.GlobalLogger.Errorf("Error sending notification: %v, gift_id: %d, count: %d", err, require.Gift.ID, require.CountForBuy)
						}
					}
				}()
				go func() {
					defer tc.wg.Done()
					tc.buyer.BuyGift(tc.ctx, newGifts)
				}()

				continue
			}
		}
	}
}

// Stop gracefully shuts down the gift service.
// It cancels the service context and waits for all goroutines to complete
// before returning, ensuring clean shutdown of all components.
func (tc *useCaseImpl) Stop() {
	if tc.cancel != nil {
		tc.cancel()
	}
	tc.wg.Wait()

	if tc.buyer != nil {
		tc.buyer.Close()
	}
}

func (tc *useCaseImpl) SetIds(ctx context.Context) error {
	return tc.accountManager.SetIds(ctx)
}

func (tc *useCaseImpl) CheckForUpdates() {
	if err := tc.checkNewUpdates(); err != nil {
		logger.GlobalLogger.Errorf("Error checking for updates: %v", err)
	}
	for {
		select {
		case <-tc.ctx.Done():
			return
		case <-tc.updateTicker.C:
			if err := tc.checkNewUpdates(); err != nil {
				logger.GlobalLogger.Errorf("Error checking for updates: %v", err)
			}
		}
	}
}

func (tc *useCaseImpl) checkNewUpdates() error {
	localVersion, err := tc.gitVersion.GetCurrentVersion()
	if err != nil {
		logger.GlobalLogger.Errorf("Error getting current version: %v", err)
		return err
	}

	remoteVersion, err := tc.gitVersion.GetLatestVersion()
	if err != nil {
		logger.GlobalLogger.Errorf("Error getting latest version: %v", err)
		return err
	}

	ok, err := tc.gitVersion.CompareVersions(localVersion, remoteVersion.TagName)
	if err != nil {
		logger.GlobalLogger.Errorf("Error comparing versions: %v", err)
		return err
	}

	if ok && tc.lastNotificationVersion != remoteVersion.TagName {
		if err := tc.notification.SendUpdateNotification(tc.ctx, remoteVersion.TagName, fmt.Sprintf("%s\n", remoteVersion.Body)); err != nil {
			logger.GlobalLogger.Errorf("Error sending update notification: %v", err)
		}
		tc.lastNotificationVersion = remoteVersion.TagName
	}
	return nil
}
