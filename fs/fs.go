package fs

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/yaoapp/gou/connector"
	"github.com/yaoapp/gou/fs/system"
	"github.com/yaoapp/kun/exception"
)

// FileSystems Register filesystems
var FileSystems = map[string]FileSystem{
	"system": system.New(),
}

// RootFileSystems high-level filesystem
var RootFileSystems = map[string]FileSystem{}

// DownloadWhitelist the file system allow download
var DownloadWhitelist = map[string]bool{
	"pdf":    true,
	"ppt":    true,
	"pptx":   true,
	"xls":    true,
	"xlsx":   true,
	"doc":    true,
	"docx":   true,
	"png":    true,
	"jpg":    true,
	"jpeg":   true,
	"bmp":    true,
	"svg":    true,
	"tif":    true,
	"mp3":    true,
	"mid":    true,
	"wma":    true,
	"wav":    true,
	"mp4":    true,
	"swf":    true,
	"rm":     true,
	"rmvb":   true,
	"mpg":    true,
	"mpeg":   true,
	"avi":    true,
	"mov":    true,
	"wmv":    true,
	"rar":    true,
	"zip":    true,
	"tar":    true,
	"gz":     true,
	"tar.gz": true,
	"7z":     true,
	"pkg":    true,
	"dmg":    true,
	"dep":    true,
	"txt":    true,
	"json":   true,
	"jsonc":  true,
	"html":   true,
	"conf":   true,
	"css":    true,
	"js":     true,
	"htm":    true,
	"webp":   true,
}

// RegisterConnector register a fileSystem via connector
func RegisterConnector(c connector.Connector) error {
	// if c.Is(connector.DATABASE) {
	// 	FileSystems[c.ID()] = system.New() // xun.New(Connector)

	// } else if c.Is(connector.REDIS) {
	// 	FileSystems[c.ID()] = system.New() // redis.New(Connector)

	// } else if c.Is(connector.MONGO) {
	// 	FileSystems[c.ID()] = system.New() // mongo.New(Connector)
	// }
	return fmt.Errorf("connector %s does not support", c.ID())
}

// Register a FileSystem
func Register(id string, fs FileSystem) FileSystem {
	FileSystems[id] = fs
	return FileSystems[id]
}

// RootRegister Register a root FileSystem
func RootRegister(id string, fs FileSystem) FileSystem {
	RootFileSystems[id] = fs
	return RootFileSystems[id]
}

// Get pick a filesystem via the given name
func Get(name string) (FileSystem, error) {
	if fs, has := FileSystems[name]; has {
		return fs, nil
	}
	return nil, fmt.Errorf("%s does not registered", name)
}

// MustGet pick a filesystem via the given name
func MustGet(name string) FileSystem {
	fs, err := Get(name)
	if err != nil {
		exception.New(err.Error(), 400).Throw()
		return nil
	}
	return fs
}

// RootGet pick a filesystem via the given name (root first)
func RootGet(name string) (FileSystem, error) {

	if fs, has := RootFileSystems[name]; has {
		return fs, nil
	}

	if fs, has := FileSystems[name]; has {
		return fs, nil
	}

	return nil, fmt.Errorf("%s does not registered", name)
}

// MustRootGet pick a filesystem via the given name
func MustRootGet(name string) FileSystem {
	fs, err := RootGet(name)
	if err != nil {
		exception.New(err.Error(), 400).Throw()
		return nil
	}
	return fs
}

// ReadFile reads the named file and returns the contents.
// A successful call returns err == nil, not err == EOF. Because ReadFile reads the whole file, it does not treat an EOF from Read as an error to be reported.
func ReadFile(xfs FileSystem, file string) ([]byte, error) {
	return xfs.ReadFile(file)
}

// ReadCloser returns a ReadCloser to read the named file.
func ReadCloser(xfs FileSystem, file string) (io.ReadCloser, error) {
	return xfs.ReadCloser(file)
}

// WriteFile writes data to the named file, creating it if necessary.
//
//	If the file does not exist, WriteFile creates it with permissions perm (before umask); otherwise WriteFile truncates it before writing, without changing permissions.
func WriteFile(xfs FileSystem, file string, data []byte, perm uint32) (int, error) {
	return xfs.WriteFile(file, data, perm)
}

// WriteCloser returns a WriteCloser that writes to the named file.
func WriteCloser(xfs FileSystem, file string, perm uint32) (io.WriteCloser, error) {
	return xfs.WriteCloser(file, perm)
}

// Write writes the content of reader to the named file, creating it if necessary.
func Write(xfs FileSystem, file string, reader io.Reader, perm uint32) (int, error) {
	return xfs.Write(file, reader, perm)
}

// AppendFile Append writes data to the named file, creating it if necessary.
// If the file does not exist, AppendFile creates it with permissions perm (before umask); otherwise AppendFile truncates it before writing, without changing permissions.
func AppendFile(xfs FileSystem, file string, data []byte, perm uint32) (int, error) {
	return xfs.AppendFile(file, data, perm)
}

// Append Append writes data to the named file, creating it if necessary.
func Append(xfs FileSystem, file string, reader io.Reader, perm uint32) (int, error) {
	return xfs.Append(file, reader, perm)
}

// InsertFile Insert writes data to the named file, creating it if necessary.
// If the file does not exist, InsertFile creates it with permissions perm (before umask); otherwise InsertFile truncates it before writing, without changing permissions.
func InsertFile(xfs FileSystem, file string, offset int64, data []byte, perm uint32) (int, error) {
	return xfs.InsertFile(file, offset, data, perm)
}

// Insert Insert writes data to the named file, creating it if necessary.
func Insert(xfs FileSystem, file string, offset int64, reader io.Reader, perm uint32) (int, error) {
	return xfs.Insert(file, offset, reader, perm)
}

// ReadDir reads the named directory, returning all its directory entries sorted by filename.
// If an error occurs reading the directory, ReadDir returns the entries it was able to read before the error, along with the error.
func ReadDir(xfs FileSystem, dir string, recursive bool) ([]string, error) {
	return xfs.ReadDir(dir, recursive)
}

// Glob returns the names of all files matching pattern or nil if there is no matching file.
// The syntax of patterns is the same as in Match. The pattern may describe hierarchical names such as /usr/*/bin/ed (assuming the Separator is '/').
func Glob(xfs FileSystem, pattern string) (matches []string, err error) {
	return xfs.Glob(pattern)
}

// Mkdir creates a new directory with the specified name and permission bits (before umask).
// If there is an error, it will be of type *PathError.
func Mkdir(xfs FileSystem, dir string, perm uint32) error {
	return xfs.Mkdir(dir, perm)
}

// MkdirAll creates a directory named path, along with any necessary parents, and returns nil, or else returns an error.
// The permission bits perm (before umask) are used for all directories that MkdirAll creates. If path is already a directory, MkdirAll does nothing and returns nil.
func MkdirAll(xfs FileSystem, dir string, perm uint32) error {
	return xfs.MkdirAll(dir, perm)
}

// MkdirTemp creates a new temporary directory in the directory dir and returns the pathname of the new directory.
// The new directory's name is generated by adding a random string to the end of pattern.
// If pattern includes a "*", the random string replaces the last "*" instead. If dir is the empty string, MkdirTemp uses the default directory for temporary files, as returned by TempDir.
// Multiple programs or goroutines calling MkdirTemp simultaneously will not choose the same directory. It is the caller's responsibility to remove the directory when it is no longer needed.
func MkdirTemp(xfs FileSystem, dir string, pattern string) (string, error) {
	return xfs.MkdirTemp(dir, pattern)
}

// Chmod changes the mode of the named file to mode. If the file is a symbolic link, it changes the mode of the link's target. If there is an error, it will be of type *PathError.
// A different subset of the mode bits are used, depending on the operating system.
// On Unix, the mode's permission bits, ModeSetuid, ModeSetgid, and ModeSticky are used.
// On Windows, only the 0200 bit (owner writable) of mode is used; it controls whether the file's read-only attribute is set or cleared. The other bits are currently unused.
// For compatibility with Go 1.12 and earlier, use a non-zero mode. Use mode 0400 for a read-only file and 0600 for a readable+writable file.
// On Plan 9, the mode's permission bits, ModeAppend, ModeExclusive, and ModeTemporary are used.
func Chmod(xfs FileSystem, name string, mode uint32) error {
	return xfs.Chmod(name, mode)
}

// Remove removes the named file or (empty) directory. If there is an error, it will be of type *PathError.
func Remove(xfs FileSystem, name string) error {
	return xfs.Remove(name)
}

// RemoveAll removes path and any children it contains. It removes everything it can but returns the first error it encounters. If the path does not exist, RemoveAll returns nil (no error). If there is an error, it will be of type *PathError.
func RemoveAll(xfs FileSystem, name string) error {
	return xfs.RemoveAll(name)
}

// Move move from src to dst
func Move(xfs FileSystem, name string, dst string) error {
	return xfs.Move(name, dst)
}

// Copy copy from src to dst
func Copy(xfs FileSystem, name string, dst string) error {
	return xfs.Copy(name, dst)
}

// merge files into new file
func Merge(xfs FileSystem, fileList []string, dst string) error {
	return xfs.Merge(fileList, dst)
}

// Exists returns a boolean indicating whether the error is known to report that a file or directory already exists.
// It is satisfied by ErrExist as well as some syscall errors.
func Exists(xfs FileSystem, name string) (bool, error) {
	return xfs.Exists(name)
}

// Size return the length in bytes for regular files; system-dependent for others
func Size(xfs FileSystem, name string) (int, error) {
	return xfs.Size(name)
}

// Mode return the file mode bits
func Mode(xfs FileSystem, name string) (uint32, error) {
	return xfs.Mode(name)
}

// ModTime return the file modification time
func ModTime(xfs FileSystem, name string) (time.Time, error) {
	return xfs.ModTime(name)
}

// IsDir check the given path is dir
func IsDir(xfs FileSystem, name string) bool {
	return xfs.IsDir(name)
}

// IsFile check the given path is file
func IsFile(xfs FileSystem, name string) bool {
	return xfs.IsFile(name)
}

// IsLink check the given path is symbolic link
func IsLink(xfs FileSystem, name string) bool {
	return xfs.IsLink(name)
}

// MimeType return the MimeType
func MimeType(xfs FileSystem, name string) (string, error) {
	return xfs.MimeType(name)
}

// MoveAppend move the file from src to dest and append the content
func MoveAppend(xfs FileSystem, src string, dst string) error {

	// check the src file exists
	if has, _ := xfs.Exists(src); !has {
		return fmt.Errorf("%s does not exists", src)
	}

	// Check the src file is a file
	if (xfs.IsFile(src)) == false {
		return fmt.Errorf("%s is not a file", src)
	}

	data, err := xfs.ReadFile(src)
	if err != nil {
		return err
	}

	// Get file pterm from src
	pterm := uint32(os.ModePerm)
	if has, _ := xfs.Exists(dst); has {

		// Check the dst file is a file
		if (xfs.IsFile(dst)) == false {
			return fmt.Errorf("%s is not a file", dst)
		}

		pterm, err = xfs.Mode(dst)
		if err != nil {
			return err
		}
	}

	// Append the file
	_, err = xfs.AppendFile(dst, data, pterm)
	if err != nil {
		return err
	}

	// remove the src file
	return xfs.Remove(src)
}

// MoveInsert move the file from src to dest and insert the content
func MoveInsert(xfs FileSystem, src string, dst string, offset int64) error {

	// check the src file exists
	if has, _ := xfs.Exists(src); !has {
		return fmt.Errorf("%s does not exists", src)
	}

	// Check the src file is a file
	if (xfs.IsFile(src)) == false {
		return fmt.Errorf("%s is not a file", src)
	}

	data, err := xfs.ReadFile(src)
	if err != nil {
		return err
	}

	// Get file pterm from src
	pterm := uint32(os.ModePerm)
	if has, _ := xfs.Exists(dst); has {

		// Check the dst file is a file
		if (xfs.IsFile(dst)) == false {
			return fmt.Errorf("%s is not a file", dst)
		}

		pterm, err = xfs.Mode(dst)
		if err != nil {
			return err
		}
	}

	// Insert the file
	_, err = xfs.InsertFile(dst, offset, data, pterm)
	if err != nil {
		return err
	}

	// remove the src file
	return xfs.Remove(src)
}

// Zip zip the dirs
func Zip(xfs FileSystem, name string, target string) error {

	if has, _ := xfs.Exists(name); !has {
		return fmt.Errorf("%s does not exists", name)
	}

	if (xfs.IsDir(name)) == false {
		return fmt.Errorf("%s is not a dir", name)
	}

	root := xfs.Root()
	absPath := filepath.Join(root, name)
	// basePath := filepath.Base(absPath)
	targetPath := filepath.Join(root, target)

	targetWriter, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer targetWriter.Close()

	writer := zip.NewWriter(targetWriter)
	defer writer.Close()

	return filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		zipPath := strings.TrimPrefix(path, absPath)
		if zipPath == "" {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		zipFile, err := writer.Create(zipPath)
		if err != nil {
			return err
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFile, file)
		if err != nil {
			return err
		}
		return nil
	})

}

// Unzip unzip the file and return the file list
func Unzip(xfs FileSystem, name string, target string) ([]string, error) {
	if (strings.HasSuffix(name, ".zip") || strings.HasSuffix(name, ".ZIP")) == false {
		return nil, fmt.Errorf("%s is not a zip file", name)
	}

	if has, _ := xfs.Exists(name); !has {
		return nil, fmt.Errorf("%s does not exists", name)
	}

	root := xfs.Root()
	absPath := filepath.Join(root, name)
	reader, err := zip.OpenReader(absPath)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	if has, _ := xfs.Exists(target); !has {
		err := xfs.MkdirAll(target, 0755)
		if err != nil {
			return nil, err
		}
	}

	if (xfs.IsDir(target)) == false {
		return nil, fmt.Errorf("%s is not a dir", target)
	}

	files := []string{}
	for _, file := range reader.File {
		name := filepath.Join(root, target, file.Name)
		if file.FileInfo().IsDir() {
			continue
		}

		if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
			return nil, err
		}

		zipFile, err := file.Open()
		if err != nil {
			return nil, err
		}
		defer zipFile.Close()

		// 创建目标文件
		targetFile, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		defer targetFile.Close()

		_, err = io.Copy(targetFile, zipFile)
		if err != nil {
			return nil, err
		}
		files = append(files, filepath.Join(target, file.Name))
	}

	return files, nil

}

// BaseName return the base name
func BaseName(name string) string {
	return filepath.Base(name)
}

// DirName return the dir name
func DirName(name string) string {
	return filepath.Dir(name)
}

// ExtName return the extension name
func ExtName(name string) string {
	return strings.TrimPrefix(filepath.Ext(name), ".")
}

// AbsPath return the absolute path
func AbsPath(xfs FileSystem, name string) (string, error) {
	return xfs.AbsPath(name)
}
