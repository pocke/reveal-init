package main

import (
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
)

// CopyFile copies a file from src to dst. If src and dst files exist, and are
// the same, then return success.  Copy the file contents from src to dst.
// See. http://stackoverflow.com/questions/21060945/simple-way-to-copy-a-file-in-golang
func CopyFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return errors.WithStack(err)
	}
	if !sfi.Mode().IsRegular() {
		// cannot copy non-regular files (e.g., directories,
		// symlinks, devices, etc.)
		return errors.Errorf("CopyFile: non-regular source file %s (%q)", sfi.Name(), sfi.Mode().String())
	}
	dfi, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return errors.WithStack(err)
		}
	} else {
		if !(dfi.Mode().IsRegular()) {
			return errors.Errorf("CopyFile: non-regular destination file %s (%q)", dfi.Name(), dfi.Mode().String())
		}
		if os.SameFile(sfi, dfi) {
			return errors.WithStack(err)
		}
	}

	// Create dst directories
	dstDir := filepath.Dir(dst)
	if !Exists(dstDir) {
		err := os.MkdirAll(dstDir, 0777)
		if err != nil {
			return errors.Wrap(err, "creating dst directory is failed")
		}
	}

	return copyFileContents(src, dst)
}

// copyFileContents copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file.
func copyFileContents(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		err = errors.WithStack(err)
		return
	}
	defer func() {
		cerr := out.Close()
		if err == nil {
			err = errors.WithStack(cerr)
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		err = errors.WithStack(err)
		return
	}
	err = errors.WithStack(out.Sync())
	return
}
