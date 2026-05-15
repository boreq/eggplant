package domain

import (
	"io"
	"time"
)

type ConvertedFile struct {
	// Name is just a filename used for mimetype detection. It is here just to
	// check its extension type basically.
	Name string

	// Modtime is used to figure out if the content has changed.
	Modtime time.Time

	// Content must be closed by the caller.
	Content io.ReadSeekCloser
}
