package auth

import (
	"context"

	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type PersistLastSeenHandler struct {
	transactionProvider TransactionProvider
	lastSeenUpdater     LastSeenUpdater
	log                 logging.Logger
}

func NewPersistLastSeenHandler(
	transactionProvider TransactionProvider,
	lastSeenUpdater LastSeenUpdater,
) *PersistLastSeenHandler {
	return &PersistLastSeenHandler{
		transactionProvider: transactionProvider,
		lastSeenUpdater:     lastSeenUpdater,
		log:                 logging.New("application/auth.PersistLastSeenHandler"),
	}
}

func (h *PersistLastSeenHandler) Execute(ctx context.Context) error {
	updates := h.lastSeenUpdater.PopUpdates()
	if len(updates) == 0 {
		return nil
	}

	if err := h.transactionProvider.Write(
		func(adapters *TransactableRepositories) error {
			for _, update := range updates {
				user, err := adapters.Users.Get(update.Username)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						continue
					}
					return errors.Wrap(err, "could not get the user")
				}

				user.UpdateLastSeen(update.LastSeen)
				for token, t := range update.Sessions {
					user.UpdateSessionLastSeen(token, t)
				}

				if err := adapters.Users.Put(*user); err != nil {
					return errors.Wrap(err, "failed to put the user")
				}
			}

			return nil
		},
	); err != nil {
		return errors.Wrap(err, "transaction failed")
	}

	return nil
}
