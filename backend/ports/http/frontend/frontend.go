//go:build !withfrontend

package frontend

import (
	"net/http"

	"github.com/boreq/errors"
)

type FrontendFileSystem struct{}

func NewFrontendFileSystem() (*FrontendFileSystem, error) {
	return &FrontendFileSystem{}, nil
}

func (f *FrontendFileSystem) Open(name string) (http.File, error) {
	return nil, errors.New("this version of the program wasn't compiled with the frontend in it")
}
