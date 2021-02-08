package auth

import (
	"time"

	"github.com/boreq/errors"
)

type CheckAccessToken struct {
	Token AccessToken
}

type CheckAccessTokenHandler struct {
	transactionProvider  TransactionProvider
	accessTokenGenerator AccessTokenGenerator
	lastSeenUpdater      LastSeenUpdater
}

func NewCheckAccessTokenHandler(
	transactionProvider TransactionProvider,
	accessTokenGenerator AccessTokenGenerator,
	lastSeenUpdater LastSeenUpdater,
) *CheckAccessTokenHandler {
	return &CheckAccessTokenHandler{
		transactionProvider:  transactionProvider,
		accessTokenGenerator: accessTokenGenerator,
		lastSeenUpdater:      lastSeenUpdater,
	}
}

func (h *CheckAccessTokenHandler) Execute(cmd CheckAccessToken) (*ReadUser, error) {
	username, err := h.accessTokenGenerator.GetUsername(cmd.Token)
	if err != nil {
		return nil, errors.Wrap(ErrUnauthorized, "could not get the username")
	}

	var foundUser *User
	var foundSession *Session

	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "user not found")
			}
			return errors.Wrap(err, "could not get the user")
		}

		foundUser = u

		for _, s := range u.Sessions {
			if s.Token == cmd.Token {
				foundSession = &s
				return nil
			}
		}

		return errors.Wrap(ErrUnauthorized, "invalid token")
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	h.lastSeenUpdater.Update(foundUser.Username, foundSession.Token, time.Now())

	rv := toReadUser(*foundUser)
	return &rv, nil
}
