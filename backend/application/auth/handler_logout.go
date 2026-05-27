package auth

import (
	"github.com/pkg/errors"
)

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

func (h *LogoutHandler) Execute(accessCtx AuthenticatedAccessContext) error {
	token := accessCtx.Token()
	username, err := token.Username()
	if err != nil {
		return errors.Wrap(err, "could not extract the username")
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		if !u.RemoveSession(token) {
			return errors.New("session not found")
		}

		return r.Users.Put(*u)
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
