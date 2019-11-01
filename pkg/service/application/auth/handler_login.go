package auth

type Login struct {
	Username string
	Password string
}

type LoginHandler struct {
	userRepository UserRepository
}

func NewLoginHandler(userRepository UserRepository) *LoginHandler {
	return &LoginHandler{
		userRepository: userRepository,
	}
}

func (h *LoginHandler) Execute(cmd Login) (AccessToken, error) {
	return h.userRepository.Login(cmd.Username, cmd.Password)
}
