package auth

import "github.com/boreq/errors"

type ListHandler struct {
	transactionProvider TransactionProvider
}

func NewListHandler(transactionProvider TransactionProvider) *ListHandler {
	return &ListHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *ListHandler) Execute() ([]ReadUser, error) {
	var users []User
	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		u, err := r.Users.List()
		if err != nil {
			return errors.Wrap(err, "could not list the users")
		}
		users = u
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}
	return toReadUsers(users), nil
}
