// Package accountManager manages accounts and their validation.s
package accountManager

import (
	"context"
	"fmt"
	"gift-buyer/pkg/errors"
	"gift-buyer/pkg/logger"
	"strings"

	"github.com/gotd/td/tg"
)

type accountManagerImpl struct {
	api                     *tg.Client
	usernames, channelNames []string
	userCache               UserCache
	channelCache            ChannelCache
}

func NewAccountManager(api *tg.Client, usernames, channelNames []string, userCache UserCache, channelCache ChannelCache) *accountManagerImpl {
	return &accountManagerImpl{
		api:          api,
		usernames:    usernames,
		channelNames: channelNames,
		userCache:    userCache,
		channelCache: channelCache,
	}
}

func (am *accountManagerImpl) SetIds(ctx context.Context) error {
	if am.api == nil {
		return errors.New("API client is nil")
	}

	if len(am.usernames) > 0 {
		if err := am.loadUsersToCache(ctx); err != nil {
			return errors.Wrap(err, "failed to load users to cache")
		}
	}

	if len(am.channelNames) > 0 {
		if err := am.loadChannelsToCache(ctx); err != nil {
			return errors.Wrap(err, "failed to load channels to cache")
		}
	}

	return nil
}

func (am *accountManagerImpl) loadUsersToCache(ctx context.Context) error {
	if am.api == nil {
		return errors.New("API client is nil")
	}

	for _, username := range am.usernames {
		withoutTag := strings.TrimPrefix(username, "@")

		res, err := am.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
			Username: withoutTag,
		})
		if err != nil {
			return errors.Wrap(err, "failed to resolve username")
		}
		for _, user := range res.Users {
			if u, ok := user.(*tg.User); ok {
				am.userCache.SetUser(withoutTag, u)
			}
		}
	}

	return nil
}

func (am *accountManagerImpl) loadChannelsToCache(ctx context.Context) error {
	cachedCount := 0
	notFoundChannels := []string{}

	for _, channelName := range am.channelNames {
		withoutTag := strings.TrimPrefix(channelName, "@")

		channel, err := am.loadSingleChannel(ctx, withoutTag)
		if err != nil {
			logger.GlobalLogger.Errorf("failed to load channel %s: %v", channelName, err)
			notFoundChannels = append(notFoundChannels, channelName)
			continue
		}

		am.channelCache.SetChannel(withoutTag, channel)
		cachedCount++
	}

	if len(notFoundChannels) > 0 {
		logger.GlobalLogger.Warnf("Channels not found or inaccessible: %v", notFoundChannels)
	}

	return nil
}

func (am *accountManagerImpl) loadSingleChannel(ctx context.Context, channelName string) (*tg.Channel, error) {
	if am.api == nil {
		return nil, errors.New("API client is nil")
	}
	res, err := am.api.ContactsResolveUsername(ctx, &tg.ContactsResolveUsernameRequest{
		Username: channelName,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve username")
	}
	for _, channel := range res.Chats {
		if c, ok := channel.(*tg.Channel); ok {
			am.channelCache.SetChannel(channelName, c)
			return c, nil
		}
	}

	return nil, errors.New(fmt.Sprintf("channel %s not found in response", channelName))
}
