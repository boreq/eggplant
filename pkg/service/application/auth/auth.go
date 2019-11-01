package auth

import "errors"

type UserRepository interface {
	// RegisterInitial should create a new user. If there are existing
	// users then this method should return an error.
	RegisterInitial(username, password string) error

	// Login should return an access token which can later be used to
	// retrieve a user by calling CheckAccessToken. If the credentials are
	// invalid then ErrUnauthorized is returned.
	Login(username, password string) (AccessToken, error)

	// CheckAccessToken should check whether the access token is valid and
	// return the user account for this access token. If the access token
	// is invalid then ErrUnauthorized is returned.
	CheckAccessToken(token AccessToken) (User, error)

	// Logout should invalidate the provided access token.
	Logout(token AccessToken) error

	// List should return a list of all users.
	List() ([]User, error)
}

var ErrUnauthorized = errors.New("unauthorized")

type AccessToken string

type User struct {
	Username      string `json:"username"`
	Administrator bool   `json:"administrator"`
}
