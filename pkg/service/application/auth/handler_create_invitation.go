package auth

type CreateInvitationHandler struct {
	userRepository UserRepository
}

func NewCreateInvitationHandler(userRepository UserRepository) *CreateInvitationHandler {
	return &CreateInvitationHandler{
		userRepository: userRepository,
	}
}

func (h *CreateInvitationHandler) Execute() (InvitationToken, error) {
	return h.userRepository.CreateInvitation()
}
