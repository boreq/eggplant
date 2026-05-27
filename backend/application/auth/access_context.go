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
	Username() authdomain.Username
	Token() authdomain.AccessToken
}

type UserAccessContext struct {
	username      authdomain.Username
	administrator bool
	token         authdomain.AccessToken
}

func NewUserAccessContext(u authdomain.User, t authdomain.AccessToken) UserAccessContext {
	return UserAccessContext{
		username:      u.Username(),
		administrator: u.Administrator(),
		token:         t,
	}
}

func (c UserAccessContext) Can(Permission) bool {
	return c.administrator
}

func (c UserAccessContext) Username() authdomain.Username {
	return c.username
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
