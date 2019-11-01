package frontend

import (
	"net/http"

	"github.com/boreq/eggplant/errors"
	_ "github.com/boreq/eggplant/statik"
	"github.com/rakyll/statik/fs"
)

type FrontendFileSystem struct {
	fs http.FileSystem
}

func NewFrontendFileSystem() (*FrontendFileSystem, error) {
	statikFS, err := fs.New()
	if err != nil {
		return nil, errors.Wrap(err, "opening statik failed")
	}

	return &FrontendFileSystem{
		fs: statikFS,
	}, nil
}

func (f *FrontendFileSystem) Open(name string) (http.File, error) {
	file, err := f.fs.Open(name)
	if err != nil {
		file, err := f.fs.Open("/index.html")
		if err != nil {
			return nil, err
		}
		return file, nil
	}
	return file, nil
}
