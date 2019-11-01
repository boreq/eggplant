package auth

type Logout struct {
	Token AccessToken
}

type LogoutHandler struct {
	userRepository UserRepository
}

func NewLogoutHandler(userRepository UserRepository) *LogoutHandler {
	return &LogoutHandler{
		userRepository: userRepository,
	}
}

func (h *LogoutHandler) Execute(cmd Logout) error {
	return h.userRepository.Logout(cmd.Token)
}
