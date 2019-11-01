package queries

import "github.com/boreq/errors"

type UserRepository interface {
	// Count should return the number of users.
	Count() (int, error)
}

type Stats struct {
	Users int `json:"users"`
}

type StatsHandler struct {
	userRepository UserRepository
}

func NewStatsHandler(userRepository UserRepository) *StatsHandler {
	return &StatsHandler{
		userRepository: userRepository,
	}
}

func (h *StatsHandler) Execute() (Stats, error) {
	users, err := h.userRepository.Count()
	if err != nil {
		return Stats{}, errors.Wrap(err, "could not count the users")
	}

	rv := Stats{
		Users: users,
	}

	return rv, nil
}
