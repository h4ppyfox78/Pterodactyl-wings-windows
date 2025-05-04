package filesystem

import (
	"sync/atomic"
	"syscall"

	"emperror.dev/errors"
	"github.com/karrick/godirwalk"
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
	var st syscall.Stat_t

	err = godirwalk.Walk(d, &godirwalk.Options{
		Unsorted: true,
		Callback: func(p string, e *godirwalk.Dirent) error {
			// If this is a symlink then resolve the final destination of it before trying to continue walking
			// over its contents. If it resolves outside the server data directory just skip everything else for
			// it. Otherwise, allow it to continue.
			if e.IsSymlink() {
				if _, err := fs.SafePath(p); err != nil {
					if IsErrorCode(err, ErrCodePathResolution) {
						return godirwalk.SkipThis
					}

					return err
				}
			}

			if !e.IsDir() {
				syscall.Lstat(p, &st)
				atomic.AddInt64(&size, st.Size)
			}

			return nil
		},
	})

	return size, errors.WrapIf(err, "server/filesystem: directorysize: failed to walk directory")
}
