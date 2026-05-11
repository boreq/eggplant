package auth

import (
	"time"

	"github.com/boreq/errors"
)

const invitationTokenBytes = 256 / 8

type CreateInvitationHandler struct {
	cryptoStringGenerator CryptoStringGenerator
	transactionProvider   TransactionProvider
}

func NewCreateInvitationHandler(
	cryptoStringGenerator CryptoStringGenerator,
	transactionProvider TransactionProvider,
) *CreateInvitationHandler {
	return &CreateInvitationHandler{
		cryptoStringGenerator: cryptoStringGenerator,
		transactionProvider:   transactionProvider,
	}
}

func (h *CreateInvitationHandler) Execute() (InvitationToken, error) {
	s, err := h.cryptoStringGenerator.Generate(invitationTokenBytes)
	if err != nil {
		return "", errors.Wrap(err, "could not create a token")
	}

	token := InvitationToken(s)

	i := Invitation{
		Token:   token,
		Created: time.Now(),
	}

	if err := h.transactionProvider.Write(func(r *TransactableRepositories) error {
		_, err := r.Invitations.Get(token)
		if !errors.Is(err, ErrNotFound) {
			return errors.Wrap(err, "token already exists, if you play the lottery right now you are guaranteed to win")
		}
		return r.Invitations.Put(i)
	}); err != nil {
		return "", errors.Wrap(err, "transaction failed")
	}

	return token, nil
}
