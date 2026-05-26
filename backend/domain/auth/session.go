package auth

import "time"

type Session struct {
	token    AccessToken
	lastSeen time.Time
}

func NewSession(token AccessToken, lastSeen time.Time) Session {
	return Session{
		token:    token,
		lastSeen: lastSeen,
	}
}

func (s Session) Token() AccessToken {
	return s.token
}

func (s Session) LastSeen() time.Time {
	return s.lastSeen
}
