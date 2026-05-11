package auth

import "github.com/pkg/errors"

type Logout struct {
	Token AccessToken
}

type LogoutHandler struct {
	transactionProvider  TransactionProvider
	accessTokenGenerator AccessTokenGenerator
}

func NewLogoutHandler(
	transactionProvider TransactionProvider,
	accessTokenGenerator AccessTokenGenerator,
) *LogoutHandler {
	return &LogoutHandler{
		transactionProvider:  transactionProvider,
		accessTokenGenerator: accessTokenGenerator,
	}
}

func (h *LogoutHandler) Execute(cmd Logout) error {
	username, err := h.accessTokenGenerator.GetUsername(cmd.Token)
	if err != nil {
		return errors.Wrap(err, "could not extract the username")
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		for i := range u.Sessions {
			if u.Sessions[i].Token == cmd.Token {
				u.Sessions = append(u.Sessions[:i], u.Sessions[i+1:]...)
				return r.Users.Put(*u)
			}
		}

		return errors.New("session not found")
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
