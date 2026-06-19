package auth

import (
	"sync"
	"time"

	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
)

type updateCache struct {
	LastSeen time.Time
	Sessions map[authdomain.AccessToken]time.Time
}

type LastSeenUpdater struct {
	userUpdates      map[authdomain.Username]*updateCache
	userUpdatesMutex sync.Mutex
}

func NewLastSeenUpdater() *LastSeenUpdater {
	return &LastSeenUpdater{
		userUpdates: make(map[authdomain.Username]*updateCache),
	}
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

func (u *LastSeenUpdater) PopUpdates() []auth.LastSeenUpdate {
	u.userUpdatesMutex.Lock()
	defer u.userUpdatesMutex.Unlock()

	if len(u.userUpdates) == 0 {
		return nil
	}

	updates := make([]auth.LastSeenUpdate, 0, len(u.userUpdates))
	for username, cache := range u.userUpdates {
		updates = append(updates, auth.LastSeenUpdate{
			Username: username,
			LastSeen: cache.LastSeen,
			Sessions: cache.Sessions,
		})
	}

	u.userUpdates = make(map[authdomain.Username]*updateCache)

	return updates
}
