package fileutils

import (
	"crypto/md5"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

const Separator = string(os.PathSeparator)
const ListSeparator = string(os.PathListSeparator)

type fileErrors []error

func (err fileErrors) Error() string {
	if err == nil || len(err) == 0 {
		return "<nil>"
	}

	fmt.Println()
	var strerr string
	for i, _ := range err {
		if i != 0 {
			strerr += "\n"
		}
		strerr += err[i].Error()
	}

	return strerr
}

func NewNoSuchError(op, filename string) error {
	return errors.New(op + ": cannot stat '" + filename + "': No such file or directory")
}

func NewNoCopyError(op, types, info string) error {
	return errors.New(op + ": cannot copy a " + types + ", " + info)
}

// 常用文件操作，实现全部linux下文件操作

func IsExist(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return true
}

func IsPermission(path string) bool {
	_, err := os.Stat(path)
	if err == nil && !os.IsPermission(err) {
		return true
	}

	return false
}

func IsDir(path string) bool {
	stat, err := os.Stat(path)
	if err == nil && stat.IsDir() {
		return true
	}

	return false
}

func IsFile(path string) bool {
	stat, err := os.Stat(path)
	if err == nil && !stat.IsDir() {
		return true
	}

	return false
}

// overwrite dst file.
func copyFile(dst, src string) error {
	var err error
	if dst, err = filepath.Abs(dst); err != nil {
		return err
	}
	if src, err = filepath.Abs(src); err != nil {
		return err
	}
	if IsExist(dst) && IsDir(dst) {
		dst = filepath.Clean(dst) + Separator + filepath.Base(src)
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()
	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()
	_, err = io.Copy(dstFile, srcFile)

	return err
}

// overwrite dst files.
func copyFiles(dst string, src ...string) error {
	switch len(src) {
	case 0:
		return NewNoSuchError("copy files", "")
	case 1:
		return copyFile(dst, src[0])
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	var err error
	var errs fileErrors
	for i, _ := range src {
		if err = copyFile(filepath.Clean(dst)+Separator+src[i], src[i]); err != nil {
			errs = append(errs, err)
		}
	}

	if errs == nil {
		return nil
	}

	return errs
}

// overwrite dst dir.
func copyDir(dst, src string) error {
	var err error
	if dst, err = filepath.Abs(dst); err != nil {
		return err
	}
	if src, err = filepath.Abs(src); err != nil {
		return err
	}
	if !IsExist(src) {
		return NewNoSuchError("copy dir", src)
	}
	if IsFile(src) {
		return copyFile(dst, src)
	}
	if IsExist(dst) {
		dst = filepath.Clean(dst) + Separator + filepath.Base(src)
	}

	if strings.HasPrefix(filepath.Clean(dst), filepath.Clean(src)) {
		return NewNoCopyError("copy dir", "directory", "'"+src+"', into itself, '"+dst+"'")
	}
	if err = os.MkdirAll(dst, 0777); err != nil {
		return err
	}
	files, err := ioutil.ReadDir(src)
	if err != nil {
		return err
	}
	var errs fileErrors
	for i, _ := range files {
		if err = copyDir(dst, filepath.Clean(src)+Separator+files[i].Name()); err != nil {
			errs = append(errs, err)
		}
	}

	if errs == nil {
		return nil
	}

	return errs
}

// overwrite dst dirs.
func copyDirs(dst string, src ...string) error {
	switch len(src) {
	case 0:
		return NewNoSuchError("copy dirs", "")
	case 1:
		return copyDir(dst, src[0])
	}
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	var err error
	var errs fileErrors
	if dst, err = filepath.Abs(dst); err != nil {
		return err
	}
	for i, _ := range src {
		if src[i], err = filepath.Abs(src[i]); err != nil {
			errs = append(errs, err)
			continue
		}

		if err = copyDir(dst+strings.TrimPrefix(src[i], filepath.Dir(dst)), src[i]); err != nil {
			errs = append(errs, err)
		}
	}

	if errs == nil {
		return nil
	}

	return errs
}

// overwrite dst file.
func moveFile(dst, src string) error {
	if err := copyFile(dst, src); err != nil {
		return err
	}

	return os.Remove(src)
}

// overwrite dst files.
func moveFiles(dst string, src ...string) error {
	switch len(src) {
	case 0:
		return NewNoSuchError("move files", "")
	case 1:
		return moveFile(dst, src[0])
	}
	if err := os.Mkdir(dst, 0755); err != nil {
		return err
	}
	var err error
	var errs fileErrors
	for i, _ := range src {
		if err = moveFile(filepath.Clean(dst)+Separator+src[i], src[i]); err != nil {
			errs = append(errs, err)
		}
	}

	if errs == nil {
		return nil
	}

	return errs
}

// overwrite dst dir.
func moveDir(dst, src string) error {
	if err := copyDir(dst, src); err != nil {
		return err
	}

	return os.RemoveAll(src)
}

// overwrite dst dirs.
func moveDirs(dst string, src ...string) error {
	switch len(src) {
	case 0:
		return NewNoSuchError("move dirs", "")
	case 1:
		return moveDir(dst, src[0])
	}
	if err := os.Mkdir(dst, 0755); err != nil {
		return err
	}
	var err error
	var errs fileErrors
	for i, _ := range src {
		if err = moveDir(filepath.Clean(dst)+Separator+src[i], src[i]); err != nil {
			errs = append(errs, err)
		}
	}

	if errs == nil {
		return nil
	}

	return errs
}

// username:password@ipaddress
// username:password@ipaddress:123
// username:password@ipaddress:123:
// username:password@ipaddress:123:/
// username:password@ipaddress:
// username:password@ipaddress:/
// username:password@ipaddress:/path
// username:password@ipaddress::/path
// username:password@ipaddress:123:/path
func scpFile(dst, src string) error {
	var (
		username = "root"
		password = ""
		address  = "127.0.0.1:22"
		filepath = "/"
	)
	if !strings.Contains(dst, "@") {
		return errors.New("invalid argument")
	}
	info := strings.SplitN(dst, "@", 2)
	username = info[0]
	if strings.Contains(info[0], ":") {
		user := strings.SplitN(info[0], ":", 2)
		username = user[0]
		password = user[1]
	}
	address = info[1]
	if strings.Contains(info[1], ":") {
		link := strings.SplitN(info[1], ":", 2)
		address = link[0]
		if link[1] == "" {

		} else if strings.HasPrefix(link[1], "/") {
			filepath = link[1]
		} else if strings.HasPrefix(link[1], ":") {
			if strings.HasPrefix(link[1][1:], "/") {
				filepath = link[1][1:]
			} else {
				filepath = "/" + link[1][1:]
			}
		} else if !strings.Contains(link[1], ":") {
			address += ":" + link[1]
		} else {
			path := strings.SplitN(link[1], ":", 2)
			if path[0] != "" {
				address += ":" + path[0]
			}
			if strings.HasPrefix(path[1], "/") {
				filepath = path[1]
			} else {
				filepath = "/" + path[1]
			}
		}
	}

	fmt.Println(username)
	fmt.Println(password)
	fmt.Println(address)
	fmt.Println(filepath)
	fmt.Println(fmt.Sprintf("%s:%s@%s:%s", username, password, address, filepath))
	fmt.Println()

	//s
	//ssh.ClientConfig{}
	//c,err := ssh.Dial()
	//c.SendRequest()

	return nil
}

//func scpFiles(dst string, src ...string) error {
//
//}
//
//func scpDir(dst, src string) error {
//
//}
//
//func scpDirs(dst string, src ...string) error {
//
//}

// create new null file.
func Touch(filename ...string) error {
	var file *os.File
	var err error
	var errs fileErrors
	for i, _ := range filename {
		file, err = os.OpenFile(filename[i], os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil && !os.IsExist(err) {
			errs = append(errs, err)
		}
		file.Close()
	}

	if errs == nil {
		return nil
	}

	return errs
}

// CopyFile copy source file to destination.
// If the source file is a directory, copy any children it contains.
// If the destination file exist, the destination file is truncated.
// If the destination file not exist, the destination file is created.
// If the destination file is directory, create a destination file under the directory.
func Copy(dst string, src ...string) error {
	return copyDirs(dst, src...)
}

// CopyFile move source file to destination.
// Returns error if the source file is a directory.
// If the destination file exist, the destination file is truncated.
// If the destination file not exist, the destination file is created.
// If the destination file is directory, create a destination file under the directory.
// Finally delete the source file.
func Move(dst string, src ...string) error {
	return moveDirs(dst, src...)
}

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error
// it encounters. If the path does not exist, RemoveAll
// returns nil (no error).
func Remove(files ...string) error {
	var err error
	var errs fileErrors
	for _, v := range files {
		if err = os.RemoveAll(v); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func SCopy(dst string, src ...string) error {
	return nil
	//return scpDirs(dst, src...)
}

// Md5sum calculate the md5 of the file.
// Returns the MD5 hex string of the file.
// If the file not exist, returns error.
func Md5sum(file string) (string, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%0x", md5.Sum(data)), nil
}
