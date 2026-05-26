package http

import (
	"net/http"

	"github.com/boreq/eggplant/application"
	"github.com/boreq/eggplant/application/auth"
	authdomain "github.com/boreq/eggplant/domain/auth"
	"github.com/boreq/errors"
)

type HttpAuthProvider struct {
	app *application.Application
}

func NewHttpAuthProvider(app *application.Application) *HttpAuthProvider {
	return &HttpAuthProvider{
		app: app,
	}
}

func (h *HttpAuthProvider) Get(r *http.Request) (*AuthenticatedUser, error) {
	token, ok := h.getToken(r)
	if !ok {
		return nil, nil
	}

	cmd := auth.CheckAccessToken{
		Token: token,
	}

	user, err := h.app.Auth.CheckAccessToken.Execute(cmd)
	if err != nil {
		if errors.Is(err, auth.ErrUnauthorized) {
			return nil, nil
		}
		return nil, errors.Wrap(err, "could not check the access token")
	}

	u := AuthenticatedUser{
		User:  *user,
		Token: token,
	}

	return &u, nil
}

func (h *HttpAuthProvider) getToken(r *http.Request) (authdomain.AccessToken, bool) {
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
