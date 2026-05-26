package auth

import (
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/pkg/errors"
)

type Logout struct {
	Token authdomain.AccessToken
}

type LogoutHandler struct {
	transactionProvider TransactionProvider
}

func NewLogoutHandler(
	transactionProvider TransactionProvider,
) *LogoutHandler {
	return &LogoutHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *LogoutHandler) Execute(cmd Logout) error {
	username, err := cmd.Token.Username()
	if err != nil {
		return errors.Wrap(err, "could not extract the username")
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		if !u.RemoveSession(cmd.Token) {
			return errors.New("session not found")
		}

		return r.Users.Put(*u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
