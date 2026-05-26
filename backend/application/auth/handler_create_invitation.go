package auth

import (
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type CreateInvitationHandler struct {
	transactionProvider TransactionProvider
}

func NewCreateInvitationHandler(
	transactionProvider TransactionProvider,
) *CreateInvitationHandler {
	return &CreateInvitationHandler{
		transactionProvider: transactionProvider,
	}
}

func (h *CreateInvitationHandler) Execute() (authdomain.InvitationToken, error) {
	token, err := authdomain.NewInvitationToken()
	if err != nil {
		return authdomain.InvitationToken{}, errors.Wrap(err, "could not create a token")
	}

	i := authdomain.NewInvitation(token, time.Now())

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		_, err := r.Invitations.Get(token)
		if !errors.Is(err, ErrNotFound) {
			return errors.Wrap(err, "token already exists, if you play the lottery right now you are guaranteed to win")
		}
		return r.Invitations.Put(i)
	}); err != nil {
		return authdomain.InvitationToken{}, errors.Wrap(err, "transaction failed")
	}

	return token, nil
}
