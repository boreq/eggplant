package http

import (
	"net/http"
	"strings"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/accessctx"
	"github.com/boreq/eggplant/application/auth"
	"github.com/boreq/eggplant/application/remote"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
	"github.com/boreq/errors"
)

type AccessContextProvider struct {
	app *application.Application
}

func NewAccessContextProvider(app *application.Application) *AccessContextProvider {
	return &AccessContextProvider{
		app: app,
	}
}

func (h *AccessContextProvider) Get(r *http.Request) (accessctx.AccessContext, error) {
	if token, ok := h.getCookieAccessToken(r); ok {
		authCtx, err := h.app.Auth.CheckAccessToken.Execute(auth.CheckAccessToken{Token: token})
		if err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				return accessctx.NewAnonymousAccessContext(), nil
			}
			return nil, errors.Wrap(err, "could not check the access token")
		}
		return authCtx, nil
	}

	if token, ok := h.getRemoteAuthToken(r); ok {
		remoteCtx, err := h.app.Remote.CheckLocalAuthToken.Execute(remote.CheckLocalAuthToken{Token: token})
		if err != nil {
			if errors.Is(err, remote.ErrUnauthorized) {
				return accessctx.NewAnonymousAccessContext(), nil
			}
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

func (h *AccessContextProvider) getRemoteAuthToken(r *http.Request) (remotedomain.AuthToken, bool) {
	const prefix = "Bearer "
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, prefix) {
		return remotedomain.AuthToken{}, false
	}
	b, err := crockford.Decode(strings.TrimPrefix(header, prefix))
	if err != nil {
		return remotedomain.AuthToken{}, false
	}
	token, err := remotedomain.NewAuthTokenFromBytes(b)
	if err != nil {
		return remotedomain.AuthToken{}, false
	}
	return token, true
}
