package dir

import (
	"github.com/karrick/godirwalk"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"time"
)

type LocalDir struct {
	path string
}

func NewLocalDir(path string) (*LocalDir, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, err
	}

	return &LocalDir{path: path}, nil
}

func (ld LocalDir) GetTypeName() string {
	return "local"
}

func (ld LocalDir) Walk(Callback WalkFunc, PostChildrenCallback PostWalkFunc) error {
	return godirwalk.Walk(ld.path, &godirwalk.Options{
		Callback: func(osPathname string, de *godirwalk.Dirent) error {
			stat, err := os.Stat(osPathname)
			if err != nil {
				return errors.Wrap(err, "stating")
			}
			relPath, err := filepath.Rel(ld.path, osPathname)
			if err != nil {
				return err
			}
			isDir, err := de.IsDirOrSymlinkToDir()
			if err != nil {
				return err
			}
			return Callback(relPath, stat.Size(), stat.ModTime(), isDir)
		},
		PostChildrenCallback: func(osPathname string, de *godirwalk.Dirent) error {
			return PostChildrenCallback(osPathname)
		},
		Unsorted:             true,
		FollowSymbolicLinks:  true,
	})
}

func (ld LocalDir) GetFile(path string) (time.Time, ReadSeekCloser, error) {
	fullPath := filepath.Join(ld.path, path)

	// get file meta-data for date, also makes sure that the file exists
	stat, err := os.Stat(fullPath)
	if err != nil {
		return time.Time{}, nil, errors.Wrap(err, "Couldn't stat file")
	}

	// read the actual file
	file, err := os.Open(fullPath)
	if err != nil {
		return time.Time{}, nil, errors.Wrap(err, "Couldn't read file")
	}

	return stat.ModTime(), file, nil
}