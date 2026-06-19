package auth

import (
	"github.com/boreq/eggplant/application/accessctx"
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

func (h *LogoutHandler) Execute(accessCtx accessctx.UserAccessContext) error {
	token := accessCtx.Token()
	username := accessCtx.Username()

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		u, err := r.Users.Get(username)
		if err != nil {
			return errors.Wrap(err, "could not get the user")
		}

		if !u.RemoveSession(token) {
			return errors.New("session not found")
		}

		if err := r.Users.Put(*u); err != nil {
			return errors.Wrap(err, "could not put the user")
		}

		return nil
	}); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
