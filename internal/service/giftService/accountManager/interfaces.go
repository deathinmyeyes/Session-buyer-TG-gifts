package accountManager

import "github.com/gotd/td/tg"

// UserCache defines the interface for caching user information.
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
}

type ChannelCache interface {
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
