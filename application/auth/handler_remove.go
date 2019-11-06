package auth

type Remove struct {
	Username string
}

type RemoveHandler struct {
	userRepository UserRepository
}

func NewRemoveHandler(userRepository UserRepository) *RemoveHandler {
	return &RemoveHandler{
		userRepository: userRepository,
	}
}

func (h *RemoveHandler) Execute(cmd Remove) error {
	return h.userRepository.Remove(cmd.Username)
}
