// Taken from https://github.com/otiai10/copy

package tpl

import (
	"bytes"
	"io"
	"os"

	"io/ioutil"
	"path/filepath"
)

// Copy copies src to dest, doesn't matter if src is a directory or a file
func Copy(src, dest string, fs *Fs) error {
	info, err := fs.Lstat(src)

	if err != nil {
		return err
	}
	return copy(src, dest, info, fs)
}

// copy dispatches copy-funcs according to the mode.
// Because this "copy" could be called recursively,
// "info" MUST be given here, NOT nil.
func copy(src, dest string, info os.FileInfo, fs *Fs) error {
	if info.Mode()&os.ModeSymlink != 0 {
		return lcopy(src, dest)
	}

	if info.IsDir() {
		return dcopy(src, dest, info, fs)
	}

	return fcopy(src, dest, info, fs)
}

// fcopy is for just a file,
// with considering existence of parent directory
// and file permission.
func fcopy(src, dest string, info os.FileInfo, fs *Fs) error {
	if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = os.Chmod(f.Name(), info.Mode()); err != nil {
		return err
	}

	b, err := fs.ReadFile(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(f, bytes.NewReader(b))
	return err
}

// dcopy is for a directory,
// with scanning contents inside the directory
// and pass everything to "copy" recursively.
func dcopy(srcdir, destdir string, info os.FileInfo, fs *Fs) error {

	if err := os.MkdirAll(destdir, info.Mode()); err != nil {
		return err
	}

	contents, err := ioutil.ReadDir(srcdir)
	if err != nil {
		return err
	}

	for _, content := range contents {
		cs, cd := filepath.Join(srcdir, content.Name()),
			filepath.Join(destdir, content.Name())

		if err := copy(cs, cd, content, fs); err != nil {
			// If any error, exit immediately
			return err
		}
	}
	return nil
}

// lcopy is for a symlink,
// with just creating a new symlink by replicating src symlink.
func lcopy(src, dest string) error {
	src, err := os.Readlink(src)

	if err != nil {
		return err
	}
	return os.Symlink(src, dest)
}
