package auth

type CheckAccessToken struct {
	Token AccessToken
}

type CheckAccessTokenHandler struct {
	userRepository UserRepository
}

func NewCheckAccessTokenHandler(userRepository UserRepository) *CheckAccessTokenHandler {
	return &CheckAccessTokenHandler{
		userRepository: userRepository,
	}
}

func (h *CheckAccessTokenHandler) Execute(cmd CheckAccessToken) (User, error) {
	return h.userRepository.CheckAccessToken(cmd.Token)
}
