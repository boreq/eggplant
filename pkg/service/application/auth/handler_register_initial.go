package auth

type RegisterInitial struct {
	Username string
	Password string
}

type RegisterInitialHandler struct {
	userRepository UserRepository
}

func NewRegisterInitialHandler(userRepository UserRepository) *RegisterInitialHandler {
	return &RegisterInitialHandler{
		userRepository: userRepository,
	}
}

func (h *RegisterInitialHandler) Execute(cmd RegisterInitial) error {
	return h.userRepository.RegisterInitial(cmd.Username, cmd.Password)
}
