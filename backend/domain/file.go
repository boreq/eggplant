package domain

import (
	"errors"
	"path/filepath"
	"strings"
)

type FilePath struct {
	value string
}

func NewFilePath(s string) (FilePath, error) {
	if s == "" {
		return FilePath{}, errors.New("file path must not be empty")
	}
	return FilePath{value: s}, nil
}

func (p FilePath) String() string {
	return p.value
}

func (p FilePath) HasExtension(ext FileExtension) bool {
	return strings.EqualFold(filepath.Ext(p.value), ext.value)
}

type FileName struct {
	value string
}

func NewFileName(s string) (FileName, error) {
	if s == "" {
		return FileName{}, errors.New("file name must not be empty")
	}
	return FileName{value: s}, nil
}

func NewFileNameFromFilePath(p FilePath) (FileName, error) {
	return NewFileName(filepath.Base(p.String()))
}

func (f FileName) String() string {
	return f.value
}

type FileExtension struct {
	value string
}

func NewFileExtension(s string) (FileExtension, error) {
	if s == "" {
		return FileExtension{}, errors.New("file extension must not be empty")
	}

	if !strings.HasPrefix(s, ".") {
		return FileExtension{}, errors.New("file extension must start with '.'")
	}

	if strings.Count(s, ".") > 1 {
		return FileExtension{}, errors.New("file extension must not contain more than one dot")
	}

	return FileExtension{value: s}, nil
}

func (e FileExtension) String() string {
	return e.value
}
