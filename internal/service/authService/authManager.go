package authService

import (
	"context"
	"errors"
	"gift-buyer/internal/config"
	"gift-buyer/internal/service/authService/apiChecker"
	"gift-buyer/internal/service/authService/authInterfaces"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type AuthManagerImpl struct {
	api             *tg.Client
	botApi          *tg.Client
	mu              sync.RWMutex
	sessionManager  authInterfaces.SessionManager
	apiChecker      authInterfaces.ApiChecker
	cfg             *config.TgSettings
	reconnect       chan struct{}
	stopCh          chan struct{}
	wg              sync.WaitGroup
	monitor         authInterfaces.GiftMonitorAndAuthController
	infoLogsWriter  authInterfaces.InfoLogger
	errorLogsWriter authInterfaces.ErrorLogger
}

func NewAuthManager(sessionManager authInterfaces.SessionManager, apiChecker authInterfaces.ApiChecker, cfg *config.TgSettings, infoLogsWriter authInterfaces.InfoLogger, errorLogsWriter authInterfaces.ErrorLogger) *AuthManagerImpl {
	return &AuthManagerImpl{
		sessionManager:  sessionManager,
		apiChecker:      apiChecker,
		cfg:             cfg,
		reconnect:       make(chan struct{}, 1),
		stopCh:          make(chan struct{}),
		infoLogsWriter:  infoLogsWriter,
		errorLogsWriter: errorLogsWriter,
	}
}

func (f *AuthManagerImpl) InitClient(ctx context.Context) (*tg.Client, error) {
	if f.cfg == nil {
		return nil, errors.New("configuration is nil")
	}

	if f.sessionManager == nil {
		return nil, errors.New("session manager is nil")
	}

	opts := telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: "session.json",
		},
	}

	// Set datacenter if specified
	if f.cfg.Datacenter > 0 {
		opts.DC = f.cfg.Datacenter
	}

	client := telegram.NewClient(f.cfg.AppId, f.cfg.ApiHash, opts)

	api, err := f.sessionManager.InitUserAPI(client, ctx)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.api = api
	f.mu.Unlock()
	return api, nil
}

func (f *AuthManagerImpl) RunApiChecker(ctx context.Context) {
	if f.apiChecker == nil {
		f.errorLogsWriter.LogError("API checker is nil, skipping")
		return
	}

	f.infoLogsWriter.LogInfo("Starting API monitoring")

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				f.infoLogsWriter.LogInfo("API monitoring stopped due to context cancellation")
				return
			case <-f.stopCh:
				f.infoLogsWriter.LogInfo("API monitoring stopped due to stop signal")
				return
			case <-ticker.C:
				if err := f.apiChecker.Run(ctx); err != nil {
					f.errorLogsWriter.LogErrorf("API check failed: %v", err)
					if f.isCriticalError(err) {
						f.errorLogsWriter.LogErrorf("Critical API error detected, triggering reconnect: %v", err)
						select {
						case f.reconnect <- struct{}{}:
							f.stopCh <- struct{}{}
							f.infoLogsWriter.LogInfo("Reconnect signal sent")
						default:
							f.errorLogsWriter.LogError("Reconnect channel is full")
						}
					}
				} else {
					f.infoLogsWriter.LogInfo("API check successful")
				}
			}
		}
	}()

	f.wg.Add(1)
	go func() {
		defer f.wg.Done()
		f.handleReconnectSignals(ctx)
	}()
}

func (f *AuthManagerImpl) isCriticalError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	return strings.Contains(errStr, "auth_key_unregistered") ||
		strings.Contains(errStr, "connection_not_inited") ||
		strings.Contains(errStr, "session_revoked")
}

func (f *AuthManagerImpl) handleReconnectSignals(ctx context.Context) {
	for {
		select {
		case <-f.reconnect:
			f.infoLogsWriter.LogInfo("Processing reconnect signal")

			if f.monitor != nil {
				f.infoLogsWriter.LogInfo("Pausing gift monitoring during reconnection")
				f.monitor.Pause()
			}

			if _, err := f.Reconnect(ctx); err != nil {
				f.errorLogsWriter.LogErrorf("Reconnect failed: %v", err)
				if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline") {
					f.errorLogsWriter.LogError("Reconnection timeout, exiting program")
					os.Exit(1)
				}
			} else {
				if f.monitor != nil {
					f.infoLogsWriter.LogInfo("Resuming gift monitoring after reconnection")
					f.monitor.Resume()
				}
				<-f.stopCh
				f.infoLogsWriter.LogInfo("Reconnection completed successfully")
			}
		case <-f.stopCh:
			f.infoLogsWriter.LogInfo("Stopping reconnect handler")
			return
		case <-ctx.Done():
			f.infoLogsWriter.LogInfo("Context cancelled, stopping reconnect handler")
			return
		}
	}
}

func (f *AuthManagerImpl) InitBotClient(ctx context.Context) (*tg.Client, error) {
	if f.sessionManager == nil {
		return nil, errors.New("session manager is nil")
	}

	botApi, err := f.sessionManager.InitBotAPI(ctx)
	if err != nil {
		return nil, err
	}
	f.mu.Lock()
	f.botApi = botApi
	f.mu.Unlock()
	return botApi, nil
}

func (f *AuthManagerImpl) GetBotApi() *tg.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.botApi
}

func (f *AuthManagerImpl) GetApi() *tg.Client {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.api
}

func (f *AuthManagerImpl) Reconnect(ctx context.Context) (*tg.Client, error) {
	f.infoLogsWriter.LogInfo("Starting reconnection process")

	tgc, err := f.InitClient(ctx)
	if err != nil {
		return nil, err
	}

	if f.apiChecker != nil {
		newApiChecker := apiChecker.NewApiChecker(tgc, time.NewTicker(2*time.Second))
		f.SetApiChecker(newApiChecker)
		f.infoLogsWriter.LogInfo("API checker updated with new client")
	}

	f.infoLogsWriter.LogInfo("Reconnection successful")
	return tgc, nil
}

func (f *AuthManagerImpl) Stop() {
	f.infoLogsWriter.LogInfo("Stopping AuthManager...")

	close(f.stopCh)

	if f.apiChecker != nil {
		f.apiChecker.Stop()
	}

	close(f.reconnect)

	f.wg.Wait()

	f.infoLogsWriter.LogInfo("AuthManager stopped")
}

func (f *AuthManagerImpl) SetApiChecker(apiChecker authInterfaces.ApiChecker) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.apiChecker = apiChecker
}

func (f *AuthManagerImpl) SetMonitor(monitor authInterfaces.GiftMonitorAndAuthController) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.monitor = monitor
	f.infoLogsWriter.LogInfo("Gift monitor set for auth manager")
}
