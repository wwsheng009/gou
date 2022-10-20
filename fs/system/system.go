package system

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/yaoapp/kun/log"
)

// File the file
type File struct {
	root      string   // the root path
	allowlist []string // the pattern list https://pkg.go.dev/path/filepath#Match
	denylist  []string // the pattern list https://pkg.go.dev/path/filepath#Match
}

// New create a new file struct
func New(root ...string) *File {
	f := &File{allowlist: []string{}, denylist: []string{}}
	if len(root) > 0 {
		f.root = root[0]
	}
	return f
}

// Allow allow rel path
func (f *File) Allow(patterns ...string) *File {
	if f.root != "" {
		for i := range patterns {
			patterns[i] = filepath.Join(f.root, patterns[i])
		}
	}
	f.allowlist = append(f.allowlist, patterns...)
	return f
}

// Deny deny rel path
func (f *File) Deny(patterns ...string) *File {
	if f.root != "" {
		for i := range patterns {
			patterns[i] = filepath.Join(f.root, patterns[i])
		}
	}
	f.denylist = append(f.denylist, patterns...)
	return f
}

// AllowAbs allow abs path
func (f *File) AllowAbs(patterns ...string) *File {
	f.allowlist = append(f.allowlist, patterns...)
	return f
}

// DenyAbs deny abs path
func (f *File) DenyAbs(patterns ...string) *File {
	f.denylist = append(f.denylist, patterns...)
	return f
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
func (f *File) ReadFile(file string) ([]byte, error) {
	file, err := f.absPath(file)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(file)
}

// WriteFile writes data to the named file, creating it if necessary.
//
//	If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions.
func (f *File) WriteFile(file string, data []byte, pterm int) (int, error) {
	file, err := f.absPath(file)
	if err != nil {
		return 0, err
	}

	dir := filepath.Dir(file)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil && !os.IsExist(err) {
		return 0, err
	}

	err = os.WriteFile(file, data, fs.FileMode(pterm))
	if err != nil {
		return 0, err
	}

	return len(data), err
}

// ReadDir reads the named directory, returning all its directory entries sorted by filename.
// If an error occurs reading the directory, ReadDir returns the entries it was able to read before the error, along with the error.
func (f *File) ReadDir(dir string, recursive bool) ([]string, error) {

	dir, err := f.absPath(dir)
	if err != nil {
		return nil, err
	}

	dirs := []string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		file := filepath.Join(dir, entry.Name())
		dirs = append(dirs, f.relPath(file))
		if recursive && entry.IsDir() {
			subdirs, err := f.ReadDir(f.relPath(file), true)
			if err != nil {
				return nil, err
			}
			dirs = append(dirs, subdirs...)
		}
	}

	return dirs, nil
}

// Mkdir creates a new directory with the specified name and permission bits (before umask).
// If there is an error, it will be of type *PathError.
func (f *File) Mkdir(dir string, pterm int) error {
	dir, err := f.absPath(dir)
	if err != nil {
		return err
	}
	return os.Mkdir(dir, fs.FileMode(pterm))
}

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error.
// The permission bits perm (before umask) are used for all directories that MkdirAll creates. If path is already a directory, MkdirAll does nothing and returns nil.
func (f *File) MkdirAll(dir string, pterm int) error {
	dir, err := f.absPath(dir)
	if err != nil {
		return err
	}
	return os.MkdirAll(dir, fs.FileMode(pterm))
}

// MkdirTemp creates a new temporary directory in the directory dir and returns the pathname of the new directory.
// The new directory's name is generated by adding a random string to the end of pattern.
// If pattern includes a "*", the random string replaces the last "*" instead. If dir is the empty string, MkdirTemp uses the default directory for temporary files, as returned by TempDir.
// Multiple programs or goroutines calling MkdirTemp simultaneously will not choose the same directory. It is the caller's responsibility to remove the directory when it is no longer needed.
func (f *File) MkdirTemp(dir string, pattern string) (string, error) {

	var err error = nil
	if dir != "" {
		dir, err = f.absPath(dir)
		if err != nil {
			return "", err
		}

		err = os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	path, err := os.MkdirTemp(dir, pattern)
	if err != nil {
		return "", err
	}

	if dir != "" {
		path = f.relPath(path)
	}

	return path, err
}

// Remove removes the named file or (empty) directory. If there is an error, it will be of type *PathError.
func (f *File) Remove(name string) error {
	name, err := f.absPath(name)
	if err != nil {
		return err
	}

	err = os.Remove(name)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		log.Warn("[Remove] %s no such file or directory", name)
	}
	return nil
}

// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters. If the path does not exist, RemoveAll returns nil (no error). If there is an error, it will be of type *PathError.
func (f *File) RemoveAll(name string) error {
	name, err := f.absPath(name)
	if err != nil {
		return err
	}
	return os.RemoveAll(name)
}

// Exists returns a boolean indicating whether the error is known to report that a file or directory already exists.
// It is satisfied by ErrExist as well as some syscall errors.
func (f *File) Exists(name string) (bool, error) {

	name, err := f.absPath(name)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(name)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// Size return the length in bytes for regular files; system-dependent for others
func (f *File) Size(name string) (int, error) {

	name, err := f.absPath(name)
	if err != nil {
		return 0, err
	}

	info, err := os.Stat(name)
	if err != nil {
		return 0, err
	}
	return int(info.Size()), nil
}

// Mode return the file mode bits
func (f *File) Mode(name string) (int, error) {
	name, err := f.absPath(name)
	if err != nil {
		return 0, err
	}

	info, err := os.Stat(name)
	if err != nil {
		return 0, err
	}
	return int(info.Mode().Perm()), nil
}

// Chmod changes the mode of the named file to mode. If the file is a symbolic link, it changes the mode of the link's target. If there is an error, it will be of type *PathError.
// A different subset of the mode bits are used, depending on the operating system.
// On Unix, the mode's permission bits, ModeSetuid, ModeSetgid, and ModeSticky are used.
// On Windows, only the 0200 bit (owner writable) of mode is used; it controls whether the file's read-only attribute is set or cleared. The other bits are currently unused.
// For compatibility with Go 1.12 and earlier, use a non-zero mode. Use mode 0400 for a read-only file and 0600 for a readable+writable file.
// On Plan 9, the mode's permission bits, ModeAppend, ModeExclusive, and ModeTemporary are used.
func (f *File) Chmod(name string, mode int) error {
	name, err := f.absPath(name)
	if err != nil {
		return err
	}

	return os.Chmod(name, fs.FileMode(mode))
}

// ModTime return the file modification time
func (f *File) ModTime(name string) (time.Time, error) {
	name, err := f.absPath(name)
	if err != nil {
		return time.Time{}, err
	}

	info, err := os.Stat(name)
	if err != nil {
		return time.Now(), err
	}
	return info.ModTime(), nil
}

// IsDir check the given path is dir
func (f *File) IsDir(name string) bool {
	name, err := f.absPath(name)
	if err != nil {
		log.Warn("[IsDir] %s %s", name, err.Error())
		return false
	}

	info, err := os.Stat(name)
	if err != nil {
		log.Warn("[IsDir] %s %s", name, err.Error())
		return false
	}
	return info.IsDir()
}

// IsFile check the given path is file
func (f *File) IsFile(name string) bool {
	name, err := f.absPath(name)
	if err != nil {
		log.Warn("[IsFile] %s %s", name, err.Error())
		return false
	}

	info, err := os.Stat(name)
	if err != nil {
		log.Warn("[IsFile] %s %s", name, err.Error())
		return false
	}
	return !info.IsDir()
}

// IsLink check the given path is symbolic link
func (f *File) IsLink(name string) bool {
	name, err := f.absPath(name)
	if err != nil {
		log.Warn("[IsLink] %s %s", name, err.Error())
		return false
	}
	info, err := os.Stat(name)
	if err != nil {
		log.Warn("[IsLink] %s %s", name, err.Error())
		return false
	}
	return info.Mode()&os.ModeSymlink != 0
}

// Move move from oldpath to newpath
func (f *File) Move(oldpath string, newpath string) error {

	oldpath, err := f.absPath(oldpath)
	if err != nil {
		return err
	}

	newpath, err = f.absPath(newpath)
	if err != nil {
		return err
	}

	err = os.Rename(oldpath, newpath)
	if err != nil && strings.Contains(err.Error(), "invalid cross-device link") {
		return f.copyRemove(f.relPath(oldpath), f.relPath(newpath))
	}
	return err
}

// Copy copy from src to dst
func (f *File) Copy(src string, dest string) error {

	src, err := f.absPath(src)
	if err != nil {
		return err
	}

	dest, err = f.absPath(dest)
	if err != nil {
		return err
	}

	stat, err := os.Stat(src)
	if err != nil {
		return err
	}

	// Copy Link
	if stat.Mode()&os.ModeSymlink != 0 {
		return f.copyLink(f.relPath(src), f.relPath(dest))
	}

	// Copy File
	if !stat.IsDir() {
		return f.copyFile(f.relPath(src), f.relPath(dest))
	}

	// Copy Dir
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		sourcePath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())
		if err := f.Copy(f.relPath(sourcePath), f.relPath(destPath)); err != nil {
			return err
		}

	}
	return nil
}

// MimeType return the MimeType
func (f *File) MimeType(name string) (string, error) {
	name, err := f.absPath(name)
	if err != nil {
		return "", err
	}

	mtype, err := mimetype.DetectFile(name)
	if err != nil {
		return "", err
	}
	return mtype.String(), nil
}

func (f *File) isTemp(path string) bool {
	return strings.HasPrefix(path, os.TempDir())
}

// absPath returns an absolute representation of path
func (f *File) absPath(path string) (string, error) {
	if f.root != "" {
		if !f.isTemp(path) {
			path = filepath.Join(f.root, path)
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	if err := f.validate(absPath); err != nil {
		return "", err
	}

	return absPath, nil
}

// relPath returns an relative representation of path
func (f *File) relPath(path string) string {
	if f.root == "" {
		return path
	}
	return strings.TrimPrefix(path, strings.TrimRight(f.root, string(os.PathSeparator)))
}

func (f *File) validate(absPath string) error {

	// check the allow list
	for _, pattern := range f.allowlist {
		match, err := filepath.Match(pattern, absPath)
		if err != nil {
			return fmt.Errorf("%s checking allowlist error %s (%s)", f.relPath(absPath), err.Error(), pattern)
		}

		if match {
			return nil
		}
	}

	// check the deny list
	for _, pattern := range f.denylist {
		match, err := filepath.Match(pattern, absPath)
		if err != nil {
			return err
		}

		if match {
			return fmt.Errorf("%s is denied (%s)", f.relPath(absPath), pattern)
		}
	}

	return nil
}

func (f *File) copyFile(src string, dest string) error {
	src, err := f.absPath(src)
	if err != nil {
		return err
	}

	dest, err = f.absPath(dest)
	if err != nil {
		return err
	}

	dir := filepath.Dir(dest)
	err = os.MkdirAll(dir, fs.ModePerm)
	if err != nil && !os.IsExist(err) {
		return err
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer out.Close()

	in, err := os.Open(src)
	defer in.Close()
	if err != nil {
		return err
	}

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return nil
}

func (f *File) copyLink(src string, dest string) error {

	src, err := f.absPath(src)
	if err != nil {
		return err
	}

	dest, err = f.absPath(dest)
	if err != nil {
		return err
	}

	link, err := os.Readlink(src)
	if err != nil {
		return err
	}
	return os.Symlink(link, dest)
}

// copyRemove copy oldpath to newpath then remove oldpath
func (f *File) copyRemove(oldpath string, newpath string) error {
	err := f.Copy(oldpath, newpath)
	if err != nil {
		return err
	}

	return f.RemoveAll(oldpath)
}
