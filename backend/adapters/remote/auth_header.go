package remote

import (
	"net/http"
	"strings"

	"github.com/boreq/eggplant/domain/crockford"
	remotedomain "github.com/boreq/eggplant/domain/remote"
)

const authorizationHeaderPrefix = "Bearer "

func SetAuthToken(r *http.Request, token remotedomain.AuthToken) {
	r.Header.Set("Authorization", authorizationHeaderPrefix+crockford.Encode(token.Bytes()))
}

func GetAuthToken(r *http.Request) (remotedomain.AuthToken, bool) {
	header := r.Header.Get("Authorization")
	if !strings.HasPrefix(header, authorizationHeaderPrefix) {
		return remotedomain.AuthToken{}, false
	}
	b, err := crockford.Decode(strings.TrimPrefix(header, authorizationHeaderPrefix))
	if err != nil {
		return remotedomain.AuthToken{}, false
	}
	token, err := remotedomain.NewAuthTokenFromBytes(b)
	if err != nil {
		return remotedomain.AuthToken{}, false
	}
	return token, true
}
