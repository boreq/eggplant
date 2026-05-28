package auth

import (
	"time"

	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type CheckAccessToken struct {
	Token authdomain.AccessToken
}

type CheckAccessTokenHandler struct {
	transactionProvider TransactionProvider
	lastSeenUpdater     LastSeenUpdater
}

func NewCheckAccessTokenHandler(
	transactionProvider TransactionProvider,
	lastSeenUpdater LastSeenUpdater,
) *CheckAccessTokenHandler {
	return &CheckAccessTokenHandler{
		transactionProvider: transactionProvider,
		lastSeenUpdater:     lastSeenUpdater,
	}
}

func (h *CheckAccessTokenHandler) Execute(cmd CheckAccessToken) (AccessContext, error) {
	var foundUser *authdomain.User
	var foundSession *authdomain.Session

	if err := h.transactionProvider.Read(func(r *TransactableRepositories) error {
		username, err := r.SessionTokens.Get(cmd.Token)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "session token not found")
			}
			return errors.Wrap(err, "could not look up the session token")
		}

		u, err := r.Users.Get(username)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return errors.Wrap(ErrUnauthorized, "user not found")
			}
			return errors.Wrap(err, "could not get the user")
		}

		foundUser = u

		for _, s := range u.Sessions() {
			if s.Token() == cmd.Token {
				session := s
				foundSession = &session
				return nil
			}
		}

		return errors.Wrap(ErrUnauthorized, "invalid token")
	}); err != nil {
		return nil, errors.Wrap(err, "transaction failed")
	}

	h.lastSeenUpdater.Update(foundUser.Username(), foundSession.Token(), time.Now())

	return NewAuthenticatedAccessContext(*foundUser, foundSession.Token()), nil
}
