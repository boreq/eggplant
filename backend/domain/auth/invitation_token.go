package auth

import "github.com/boreq/errors"

const invitationTokenBytes = 256 / 8

type InvitationToken struct {
	value string
}

func NewInvitationToken() (InvitationToken, error) {
	s, err := generateCryptoString(invitationTokenBytes)
	if err != nil {
		return InvitationToken{}, errors.Wrap(err, "could not generate the token")
	}
	return InvitationToken{value: s}, nil
}

func NewInvitationTokenFromString(s string) (InvitationToken, error) {
	if s == "" {
		return InvitationToken{}, errors.New("invitation token must not be empty")
	}
	return InvitationToken{value: s}, nil
}

func (t InvitationToken) String() string {
	return t.value
}
