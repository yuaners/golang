package fileutils

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	fileExistError        = errors.New("the file existed")
	fileNotExistError     = errors.New("the file not existed")
	fileNotDirectoryError = errors.New("the file not directory")
	fileIsDirectoryError  = errors.New("the file is directory")
)

type FileInfo interface {
	os.FileInfo
	SizeAll() (int64, error)       // the file return this self size. the dir return all child files and dirs total size
	Files() (int64, error)         // file number
	Dirs() (int64, error)          // directory number
	Location() string              // location
	ModifyTime() time.Time         // modification time
	CreateTime() time.Time         // create time, windows only
	ChangeTime() time.Time         // change time, linux only
	AccessTime() time.Time         // last access time
	FDS() (f, d, s int64, e error) // files/dirs/size total, recommended use
}

type fileStat struct {
	location    string // location
	os.FileInfo        // os file info
}

func (f *fileStat) Files() (int64, error) {
	if !f.IsDir() {
		return 1, nil
	}

	fs, err := ReadDir(f.Location())
	if err != nil {
		if os.IsPermission(err) {
			return 0, nil
		}
		return 0, err
	}

	var files int64 = 0
	for _, f := range fs {
		num, err := f.Files()
		if err != nil {
			return 0, err
		}
		files += num
	}

	return files, nil
}

func (f *fileStat) Dirs() (int64, error) {
	if !f.IsDir() {
		return 0, nil
	}

	fs, err := ReadDir(f.Location())
	if err != nil {
		if os.IsPermission(err) {
			return 0, nil
		}
		return 0, err
	}

	var dirs int64 = 0
	for _, f := range fs {
		num, err := f.Dirs()
		if err != nil {
			return 0, err
		}
		dirs += num
	}

	return dirs + 1, nil
}

func (f *fileStat) SizeAll() (int64, error) {
	if !f.IsDir() {
		return f.Size(), nil
	}

	files, err := ReadDir(f.Location())
	if err != nil {
		if os.IsPermission(err) {
			return 0, nil
		}
		return 0, err
	}

	var size int64 = 0
	for _, file := range files {
		s, err := file.SizeAll()
		if err != nil {
			return 0, err
		}
		size += s
	}

	return size + f.Size(), nil
}

func (f *fileStat) ModifyTime() time.Time {
	return f.FileInfo.ModTime()
}

func (f *fileStat) CreateTime() time.Time {
	return f.createTime()
}

func (f *fileStat) ChangeTime() time.Time {
	return f.changeTime()
}

func (f *fileStat) AccessTime() time.Time {
	return f.accessTime()
}

func (f *fileStat) Location() string {
	return f.location
}

func (f *fileStat) FDS() (files, dirs, size int64, err error) {
	if !f.IsDir() {
		return 1, 0, f.Size(), nil
	}

	fi, err := ReadDir(f.Location())
	if err != nil {
		if os.IsPermission(err) {
			return 0, 0, 0, nil
		}
		return 0, 0, 0, err
	}

	for _, file := range fi {
		f, d, s, err := file.FDS()
		if err != nil {
			return 0, 0, 0, err
		}

		files += f
		dirs += d
		size += s
	}

	return files, dirs + 1, size + f.Size(), nil
}

// ReadDir reads the directory named by dirname and returns
// a list of directory entries sorted by filename.
func ReadDir(name string) ([]FileInfo, error) {
	if !filepath.IsAbs(name) {
		n, err := filepath.Abs(name)
		if err != nil {
			return nil, err
		}
		name = n
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	list, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	var fs []FileInfo
	for _, fi := range list {
		fs = append(fs, &fileStat{location: strings.Replace(name, "\\", "/", -1) + "/" + fi.Name(), FileInfo: fi})
	}

	return fs, nil
}

// Stat returns a FileInfo describing the named file.
// If there is an error, it will be of type *PathError.
func Stat(name string) (FileInfo, error) {
	name, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	}

	return &fileStat{location: strings.Replace(name, "\\", "/", -1), FileInfo: fi}, nil
}

// Lstat returns a FileInfo describing the named file.
// If the file is a symbolic link, the returned FileInfo
// describes the symbolic link. Lstat makes no attempt to follow the link.
// If there is an error, it will be of type *PathError.
func LStat(name string) (FileInfo, error) {
	name, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}

	fi, err := os.Lstat(name)
	if err != nil {
		return nil, err
	}

	return &fileStat{location: strings.Replace(name, "\\", "/", -1), FileInfo: fi}, nil
}

// SameFile reports whether fi1 and fi2 describe the same file.
// For example, on Unix this means that the device and inode fields
// of the two underlying structures are identical; on other systems
// the decision may be based on the path names.
// SameFile only applies to results returned by this package's Stat.
// It returns false in other cases.
func SameFile(f1, f2 FileInfo) bool {
	fs1, ok1 := f1.(*fileStat)
	fs2, ok2 := f2.(*fileStat)
	if !ok1 || !ok2 {
		return false
	}

	return os.SameFile(fs1, fs2)
}
