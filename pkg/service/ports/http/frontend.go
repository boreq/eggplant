package http

import "net/http"

type frontendFileSystem struct {
	fs http.FileSystem
}

func newFrontendFileSystem(fs http.FileSystem) *frontendFileSystem {
	return &frontendFileSystem{
		fs: fs,
	}
}

func (f *frontendFileSystem) Open(name string) (http.File, error) {
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
