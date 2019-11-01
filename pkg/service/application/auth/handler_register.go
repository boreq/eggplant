package auth

type Register struct {
	Username string
	Password string
	Token    InvitationToken
}

type RegisterHandler struct {
	userRepository UserRepository
}

func NewRegisterHandler(userRepository UserRepository) *RegisterHandler {
	return &RegisterHandler{
		userRepository: userRepository,
	}
}

func (h *RegisterHandler) Execute(cmd Register) error {
	return h.userRepository.Register(cmd.Username, cmd.Password, cmd.Token)
}
