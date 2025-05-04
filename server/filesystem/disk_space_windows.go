package filesystem

import (
	"os"
	"path/filepath"

	"emperror.dev/errors"
)

// Determines the directory size of a given location by running parallel tasks to iterate
// through all of the folders. Returns the size in bytes. This can be a fairly taxing operation
// on locations with tons of files, so it is recommended that you cache the output.
func (fs *Filesystem) DirectorySize(dir string) (int64, error) {
	d, err := fs.SafePath(dir)
	if err != nil {
		return 0, err
	}

	var size int64
	err = filepath.Walk(d, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})

	return size, errors.WrapIf(err, "server/filesystem: directorysize: failed to walk directory")
}
