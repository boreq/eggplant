package auth

type Remove struct {
	Username string
}

type RemoveHandler struct {
	transactionProvider TransactionProvider
}

func NewRemoveHandler(transactionProvider TransactionProvider) *RemoveHandler {
	return &RemoveHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *RemoveHandler) Execute(cmd Remove) error {
	return h.transactionProvider.Write(func(r *TransactableRepositories) error {
		return r.Users.Remove(cmd.Username)
	})
}
