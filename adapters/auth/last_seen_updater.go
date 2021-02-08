package auth

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/logging"
	"github.com/boreq/errors"
)

type updateCache struct {
	LastSeen time.Time
	Sessions map[auth.AccessToken]time.Time
}

type LastSeenUpdater struct {
	log                 logging.Logger
	transactionProvider auth.TransactionProvider
	userUpdates         map[string]*updateCache
	userUpdatesMutex    sync.Mutex
}

func NewLastSeenUpdater(transactionProvider auth.TransactionProvider) (*LastSeenUpdater, error) {
	return &LastSeenUpdater{
		transactionProvider: transactionProvider,
		userUpdates:         make(map[string]*updateCache),
	}, nil
}

func (u *LastSeenUpdater) Update(username string, token auth.AccessToken, t time.Time) {
	u.userUpdatesMutex.Lock()
	defer u.userUpdatesMutex.Unlock()

	c, ok := u.userUpdates[username]
	if !ok {
		u.userUpdates[username] = &updateCache{
			LastSeen: t,
			Sessions: map[auth.AccessToken]time.Time{
				token: t,
			},
		}
	} else {
		if t.After(c.LastSeen) {
			c.LastSeen = t
		}

		if t.After(c.Sessions[token]) {
			c.Sessions[token] = t
		}
	}
}

func (u *LastSeenUpdater) Run(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	for {
		select {
		case <-t.C:
			if err := u.flush(); err != nil {
				u.log.Error("last seen updater error", "err", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (u *LastSeenUpdater) flush() error {
	u.userUpdatesMutex.Lock()
	defer u.userUpdatesMutex.Unlock()

	if len(u.userUpdates) == 0 {
		return nil
	}

	if err := u.transactionProvider.Write(
		func(adapters *auth.TransactableRepositories) error {
			for username, cache := range u.userUpdates {
				u, err := adapters.Users.Get(username)
				if err != nil {
					if errors.Is(err, auth.ErrNotFound) {
						continue
					}
					return errors.Wrap(err, "could not get the user")
				}

				if cache.LastSeen.After(u.LastSeen) {
					u.LastSeen = cache.LastSeen
				}

				for token, t := range cache.Sessions {
					for i := range u.Sessions {
						if u.Sessions[i].Token == token {
							if t.After(u.Sessions[i].LastSeen) {
								u.Sessions[i].LastSeen = t
							}
						}
					}
				}

				if err := adapters.Users.Put(*u); err != nil {
					return errors.Wrap(err, "failed to put the user")
				}
			}

			return nil
		},
	); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	u.userUpdates = make(map[string]*updateCache)

	return nil
}
