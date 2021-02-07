package symwalk

import (
	"os"
	"path/filepath"
	"sort"

	"github.com/pkg/errors"
)

type WalkFunc func(path string, info os.FileInfo, err error) error

func Walk(root string, walkFn WalkFunc) error {
	info, err := os.Lstat(root)

	if err != nil {
		err = walkFn(root, nil, err)
	} else {
		err = walk(root, info, walkFn)
	}

	if err == filepath.SkipDir {
		return nil
	}

	return err

}

func walk(path string, info os.FileInfo, walkFn WalkFunc) error {
	if info.Mode()&os.ModeDir == 0 && info.Mode()&os.ModeSymlink == 0 {
		return walkFn(path, info, nil)
	}

	if info.Mode()&os.ModeSymlink != 0 {
		targetPath, err := filepath.EvalSymlinks(path)
		if err != nil {
			return errors.Wrap(err, "could not eval a symlink")
		}

		targetInfo, err := os.Lstat(targetPath)
		if err != nil {
			return errors.Wrap(err, "could not lstat a symlink target")
		}

		if targetInfo.Mode()&os.ModeDir == 0 {
			return walkFn(path, targetInfo, nil)
		}
	}

	names, err := readDirNames(path)
	err1 := walkFn(path, info, err)
	// If err != nil, walk can't walk into this directory.
	// err1 != nil means walkFn want walk to skip this directory or stop walking.
	// Therefore, if one of err and err1 isn't nil, walk will return.
	if err != nil || err1 != nil {
		// The caller's behavior is controlled by the return value, which is decided
		// by walkFn. walkFn may ignore err and return nil.
		// If walkFn returns SkipDir, it will be handled by the caller.
		// So walk should return whatever walkFn returns.
		return err1
	}

	for _, name := range names {
		filename := filepath.Join(path, name)
		fileInfo, err := os.Lstat(filename)
		if err != nil {
			if err := walkFn(filename, fileInfo, err); err != nil && err != filepath.SkipDir {
				return err
			}
		} else {
			err = walk(filename, fileInfo, walkFn)
			if err != nil {
				if !fileInfo.IsDir() || err != filepath.SkipDir {
					return err
				}
			}
		}
	}

	return nil
}

func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}

	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}

	sort.Strings(names)
	return names, nil
}
