package auth

import "time"

type Invitation struct {
	token   InvitationToken
	created time.Time
}

func NewInvitation(token InvitationToken, created time.Time) Invitation {
	return Invitation{
		token:   token,
		created: created,
	}
}

func (i Invitation) Token() InvitationToken {
	return i.token
}

func (i Invitation) Created() time.Time {
	return i.created
}
