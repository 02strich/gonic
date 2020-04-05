package dir

import (
	"io"
	"time"
)

type WalkFunc func(relPath string, fileSize int64, modTime time.Time, isDirectory bool) error
type PostWalkFunc func(relPath string) error

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type Dir interface {
	GetTypeName() string
	Walk(Callback WalkFunc, PostChildrenCallback PostWalkFunc) error
	GetFile(path string) (time.Time, ReadSeekCloser, error)
}
