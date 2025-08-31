package authInterfaces

import (
	"context"

	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
)

type AuthManager interface {
	// CreateDeviceConfig creates a new device configuration for the Telegram client
	//
	// Returns:
	//   - telegram.DeviceConfig: device configuration for the Telegram client
	CreateDeviceConfig() telegram.DeviceConfig

	// InitClient initializes the Telegram client
	// Parameters:
	//   - client: Telegram client to initialize
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - *tg.Client: authenticated Telegram API client
	//   - error: authentication error, network error, or timeout
	InitClient(client *telegram.Client, ctx context.Context) (*tg.Client, error)

	// CreateBotClient creates a new bot client for the Telegram client
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - *tg.Client: authenticated Telegram API client
	//   - error: authentication error, network error, or timeout
	CreateBotClient(ctx context.Context) (*tg.Client, error)

	// Stop stops the API checker
	Stop()

	// SetMonitor устанавливает монитор подарков для управления его состоянием
	// во время переподключения
	SetMonitor(monitor GiftMonitorAndAuthController)

	// SetGlobalCancel устанавливает функцию для остановки всей программы
	// при критических ошибках переподключения
	SetGlobalCancel(cancel context.CancelFunc)
}

// SessionManager interface defines the methods for managing Telegram sessions
type SessionManager interface {
	// InitUserAPI initializes the Telegram client
	// Parameters:
	//   - client: Telegram client to initialize
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - *tg.Client: authenticated Telegram API client
	//   - error: authentication error, network error, or timeout
	InitUserAPI(client *telegram.Client, ctx context.Context) (*tg.Client, error)

	// InitBotAPI creates a new bot client for the Telegram client
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - *tg.Client: authenticated Telegram API client
	//   - error: authentication error, network error, or timeout
	InitBotAPI(ctx context.Context) (*tg.Client, error)
}

// ApiChecker interface defines the methods for checking the Telegram API
type ApiChecker interface {
	// Run checks the Telegram API continuously
	// Parameters:
	//   - ctx: context for cancellation and timeout control
	//
	// Returns:
	//   - error: error if the API is not working
	Run(ctx context.Context) error

	// // CheckApi checks the Telegram API once
	// // Parameters:
	// //   - ctx: context for cancellation and timeout control
	// //
	// // Returns:
	// //   - error: error if the API is not working
	// CheckApi(ctx context.Context) error

	// Stop stops the API checker
	Stop()
}

// GiftMonitorManager defines the interface for managing gift monitoring.
// It provides methods to pause, resume, and check the status of the gift monitoring process.
type GiftMonitorAndAuthController interface {
	// Pause pauses the gift monitoring process.
	Pause()

	// Resume resumes the gift monitoring process.
	Resume()

	// IsPaused returns the status of the gift monitoring process.
	IsPaused() bool
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
