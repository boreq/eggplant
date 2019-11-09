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
}

func NewCheckAccessTokenHandler(
	transactionProvider TransactionProvider,
	accessTokenGenerator AccessTokenGenerator,
) *CheckAccessTokenHandler {
	return &CheckAccessTokenHandler{
		transactionProvider:  transactionProvider,
		accessTokenGenerator: accessTokenGenerator,
	}
}

func (h *CheckAccessTokenHandler) Execute(cmd CheckAccessToken) (*ReadUser, error) {
	username, err := h.accessTokenGenerator.GetUsername(cmd.Token)
	if err != nil {
		return nil, errors.Wrap(ErrUnauthorized, "could not get the username")
	}

	var foundUser *User
	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "user not found")
			}
			return errors.Wrap(err, "could not get the user")
		}

		for i := range u.Sessions {
			if u.Sessions[i].Token == cmd.Token {
				u.LastSeen = time.Now()
				u.Sessions[i].LastSeen = time.Now()
				foundUser = u
				return r.Users.Put(*u)
			}
		}

		return errors.Wrap(ErrUnauthorized, "invalid token")
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	rv := toReadUser(*foundUser)
	return &rv, nil
}
