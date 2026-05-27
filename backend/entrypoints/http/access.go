package http

import (
	"net/http"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/music/library"
	"github.com/boreq/errors"
)

type AccessContext struct {
	auth    auth.AccessContext
	library library.AccessContext
}

func NewAccessContext(authCtx auth.AccessContext) AccessContext {
	return AccessContext{
		auth:    authCtx,
		library: libraryAccessFor(authCtx),
	}
}

func (c AccessContext) Auth() auth.AccessContext {
	return c.auth
}

func (c AccessContext) Library() library.AccessContext {
	return c.library
}

func libraryAccessFor(authCtx auth.AccessContext) library.AccessContext {
	if _, ok := authCtx.(auth.AuthenticatedAccessContext); ok {
		return library.NewLoggedInAccessContext()
	}
	return library.NewAnonymousAccessContext()
}

type AccessContextProvider struct {
	app *application.Application
}

func NewAccessContextProvider(app *application.Application) *AccessContextProvider {
	return &AccessContextProvider{
		app: app,
	}
}

func (h *AccessContextProvider) Get(r *http.Request) (AccessContext, error) {
	token, ok := h.getToken(r)
	if !ok {
		return NewAccessContext(auth.NewAnonymousAccessContext()), nil
	}

	authCtx, err := h.app.Auth.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token})
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return NewAccessContext(auth.NewAnonymousAccessContext()), nil
		}
		return AccessContext{}, errors.Wrap(err, "could not check the access token")
	}

	return NewAccessContext(authCtx), nil
}

func (h *AccessContextProvider) getToken(r *http.Request) (authdomain.AccessToken, bool) {
	c, err := r.Cookie("auth-token")
	if err != nil {
		return authdomain.AccessToken{}, false
	}
	token, err := authdomain.NewAccessTokenFromString(c.Value)
	if err != nil {
		return authdomain.AccessToken{}, false
	}
	return token, true
}
