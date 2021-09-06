package frontend

import (
	"embed"
	"net/http"
)

//go:embed css/* js/* img/* index.html favicon.ico
var content embed.FS

type FrontendFileSystem struct {
	fs http.FileSystem
}

func NewFrontendFileSystem() (*FrontendFileSystem, error) {
	return &FrontendFileSystem{
		fs: http.FS(content),
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
