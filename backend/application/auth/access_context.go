package auth

import authdomain "github.com/boreq/eggplant/domain/auth"

type Permission struct {
	value string
}

func (p Permission) String() string {
	return p.value
}

var (
	PermissionManageUsers       = Permission{value: "manage_users"}
	PermissionCreateInvitations = Permission{value: "create_invitations"}
)

type AccessContext interface {
	Can(p Permission) bool
}

type AuthenticatedAccessContext interface {
	AccessContext
	User() authdomain.User
	Token() authdomain.AccessToken
}

type UserAccessContext struct {
	user  authdomain.User
	token authdomain.AccessToken
}

func NewUserAccessContext(u authdomain.User, t authdomain.AccessToken) UserAccessContext {
	return UserAccessContext{user: u, token: t}
}

func (c UserAccessContext) Can(Permission) bool {
	return c.user.Administrator()
}

func (c UserAccessContext) User() authdomain.User {
	return c.user
}

func (c UserAccessContext) Token() authdomain.AccessToken {
	return c.token
}

type AdminAccessContext struct{}

func NewAdminAccessContext() AdminAccessContext {
	return AdminAccessContext{}
}

func (AdminAccessContext) Can(Permission) bool {
	return true
}

type AnonymousAccessContext struct{}

func NewAnonymousAccessContext() AnonymousAccessContext {
	return AnonymousAccessContext{}
}

func (AnonymousAccessContext) Can(Permission) bool {
	return false
}
