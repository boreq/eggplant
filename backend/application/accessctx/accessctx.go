package accessctx

import (
	"errors"

	authdomain "github.com/boreq/eggplant/domain/auth"
	remotedomain "github.com/boreq/eggplant/domain/remote"
)

var ErrPermissionDenied = errors.New("permission denied")

type Permission struct {
	value string
}

func (p Permission) String() string {
	return p.value
}

var (
	PermissionManageUsers       = Permission{value: "manage_users"}
	PermissionManageInvitations = Permission{value: "manage_invitations"}
	PermissionManageRemotes     = Permission{value: "manage_remotes"}
)

type AccessContext interface {
	Can(p Permission) bool
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

type RemoteInstanceAccessContext struct {
	remoteInstanceID remotedomain.RemoteInstanceID
}

func NewRemoteInstanceAccessContext(remoteInstanceID remotedomain.RemoteInstanceID) RemoteInstanceAccessContext {
	return RemoteInstanceAccessContext{remoteInstanceID: remoteInstanceID}
}

func (c RemoteInstanceAccessContext) Can(Permission) bool {
	return false
}

func (c RemoteInstanceAccessContext) RemoteInstanceID() remotedomain.RemoteInstanceID {
	return c.remoteInstanceID
}

type CommandLineAccessContext struct{}

func NewCommandLineAccessContext() CommandLineAccessContext {
	return CommandLineAccessContext{}
}

func (CommandLineAccessContext) Can(Permission) bool {
	return true
}

type AnonymousAccessContext struct{}

func NewAnonymousAccessContext() AnonymousAccessContext {
	return AnonymousAccessContext{}
}

func (AnonymousAccessContext) Can(Permission) bool {
	return false
}
