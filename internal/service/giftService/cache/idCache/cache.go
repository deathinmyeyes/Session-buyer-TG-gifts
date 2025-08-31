package idCache

import (
	"errors"
	"sync"

	"github.com/gotd/td/tg"
)

type idCacheImpl struct {
	users    map[string]*tg.User
	channels map[string]*tg.Channel
	mu       sync.RWMutex
}

func NewIDCache() *idCacheImpl {
	return &idCacheImpl{
		users:    make(map[string]*tg.User),
		channels: make(map[string]*tg.Channel),
	}
}

func (c *idCacheImpl) SetUser(key string, user *tg.User) {
	if user == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.users[key] = user
}

func (c *idCacheImpl) GetUser(key string) (*tg.User, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	user, ok := c.users[key]
	if !ok {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (c *idCacheImpl) SetChannel(key string, channel *tg.Channel) {
	if channel == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels[key] = channel
}

func (c *idCacheImpl) GetChannel(key string) (*tg.Channel, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	channel, ok := c.channels[key]
	if !ok {
		return nil, errors.New("channel not found")
	}
	return channel, nil
}
