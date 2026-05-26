package auth

import "time"

type User struct {
	username      Username
	password      PasswordHash
	administrator bool
	created       time.Time
	lastSeen      time.Time
	sessions      []Session
}

func NewUser(username Username, password PasswordHash, administrator bool, created, lastSeen time.Time, sessions []Session) User {
	return User{
		username:      username,
		password:      password,
		administrator: administrator,
		created:       created,
		lastSeen:      lastSeen,
		sessions:      sessions,
	}
}

func (u *User) Username() Username {
	return u.username
}

func (u *User) Password() PasswordHash {
	return u.password
}

func (u *User) Administrator() bool {
	return u.administrator
}

func (u *User) Created() time.Time {
	return u.created
}

func (u *User) LastSeen() time.Time {
	return u.lastSeen
}

func (u *User) Sessions() []Session {
	return u.sessions
}

func (u *User) AddSession(s Session) {
	u.sessions = append(u.sessions, s)
}

func (u *User) RemoveSession(token AccessToken) bool {
	for i := range u.sessions {
		if u.sessions[i].token == token {
			u.sessions = append(u.sessions[:i], u.sessions[i+1:]...)
			return true
		}
	}
	return false
}

func (u *User) SetPassword(hash PasswordHash) {
	u.password = hash
}

func (u *User) UpdateLastSeen(t time.Time) {
	if t.After(u.lastSeen) {
		u.lastSeen = t
	}
}

func (u *User) UpdateSessionLastSeen(token AccessToken, t time.Time) {
	for i := range u.sessions {
		if u.sessions[i].token == token && t.After(u.sessions[i].lastSeen) {
			u.sessions[i].lastSeen = t
		}
	}
}
