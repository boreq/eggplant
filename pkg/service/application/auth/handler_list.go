package auth

type ListHandler struct {
	userRepository UserRepository
}

func NewListHandler(userRepository UserRepository) *ListHandler {
	return &ListHandler{
		userRepository: userRepository,
	}
}

func (h *ListHandler) Execute() ([]User, error) {
	return h.userRepository.List()
}
