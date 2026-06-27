package http

import (
	"net/http"

	remoteadapter "github.com/boreq/eggplant/adapters/remote"
	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/remote"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/internal/logging"
	"github.com/boreq/errors"
)

type AccessContextProvider struct {
	app *application.Application
	log logging.Logger
}

func NewAccessContextProvider(app *application.Application) *AccessContextProvider {
	return &AccessContextProvider{
		app: app,
		log: logging.New("ports/http.AccessContextProvider"),
	}
}

func (h *AccessContextProvider) Get(r *http.Request) (accessctx.AccessContext, error) {
	if token, ok := h.getCookieAccessToken(r); ok {
		authCtx, err := h.app.Auth.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token})
		if err != nil {
			return nil, errors.Wrap(err, "could not check the access token")
		}
		return authCtx, nil
	}

	if token, ok := remoteadapter.GetAuthToken(r); ok {
		remoteCtx, err := h.app.Remote.CheckLocalAuthToken.Execute(remote.CheckLocalAuthToken{Token: token})
		if err != nil {
			return nil, errors.Wrap(err, "could not check the remote auth token")
		}
		return remoteCtx, nil
	}

	return accessctx.NewAnonymousAccessContext(), nil
}

func (h *AccessContextProvider) getCookieAccessToken(r *http.Request) (authdomain.AccessToken, bool) {
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
