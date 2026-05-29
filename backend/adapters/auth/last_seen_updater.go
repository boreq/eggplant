package auth

import (
	"context"
	"sync"
	"time"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type updateCache struct {
	LastSeen time.Time
	Sessions map[authdomain.AccessToken]time.Time
}

type LastSeenUpdater struct {
	log                 logging.Logger
	transactionProvider auth.TransactionProvider
	userUpdates         map[authdomain.Username]*updateCache
	userUpdatesMutex    sync.Mutex
}

func NewLastSeenUpdater(transactionProvider auth.TransactionProvider) (*LastSeenUpdater, error) {
	return &LastSeenUpdater{
		log:                 logging.New("adapters/auth.LastSeenUpdater"),
		transactionProvider: transactionProvider,
		userUpdates:         make(map[authdomain.Username]*updateCache),
	}, nil
}

func (u *LastSeenUpdater) Update(username authdomain.Username, token authdomain.AccessToken, t time.Time) {
	u.userUpdatesMutex.Lock()
	defer u.userUpdatesMutex.Unlock()

	c, ok := u.userUpdates[username]
	if !ok {
		u.userUpdates[username] = &updateCache{
			LastSeen: t,
			Sessions: map[authdomain.AccessToken]time.Time{
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
				user, err := adapters.Users.Get(username)
				if err != nil {
					if errors.Is(err, auth.ErrNotFound) {
						continue
					}
					return errors.Wrap(err, "could not get the user")
				}

				user.UpdateLastSeen(cache.LastSeen)
				for token, t := range cache.Sessions {
					user.UpdateSessionLastSeen(token, t)
				}

				if err := adapters.Users.Put(*user); err != nil {
					return errors.Wrap(err, "failed to put the user")
				}
			}

			return nil
		},
	); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	u.userUpdates = make(map[authdomain.Username]*updateCache)

	return nil
}
