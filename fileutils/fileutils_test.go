package fileutils

import (
	"fmt"
	"os"
	"testing"
)

func TestTouch(t *testing.T) {
	if err := os.MkdirAll("test", 755); err != nil {
		panic(err)
	}
	if err := Touch("test/a", "test/b", "test/c", "test/d"); err != nil {
		panic(err)
	}
}

func TestIsExist(t *testing.T) {
	fmt.Println(IsExist("test/a"))
	fmt.Println(IsExist("test"))
	fmt.Println(IsExist("a"))
}

func TestIsFile(t *testing.T) {
	fmt.Println(IsFile("test/a"))
	fmt.Println(IsFile("test"))
	fmt.Println(IsFile("a"))
}

func TestIsDir(t *testing.T) {
	fmt.Println(IsDir("test/a"))
	fmt.Println(IsDir("test"))
	fmt.Println(IsDir("a"))
}

func TestCopy(t *testing.T) {
	if err := os.MkdirAll("./test/dir1", 0755); err != nil {
		panic(err)
	}

	if err := Copy("./test/dir2", "./test/dir1"); err != nil {
		panic(err)
	}
	if err := Copy("./test/dir2", "./test/dir1"); err != nil {
		panic(err)
	}

	if err := Copy("./test/dir1", "./test/a"); err != nil {
		panic(err)
	}
	if err := Copy("./test/dir1/a.copy", "./test/a"); err != nil {
		panic(err)
	}

	if err := Copy("./test/dir2", "./test/dir1"); err != nil {
		panic(err)
	}

	if err := Copy("./test/dir2", "./test/a", "./test/b", "./test/c", "./test/d"); err != nil {
		panic(err)
	}

	if err := Copy("./test/dir1", "./test/dir2"); err != nil {
		panic(err)
	}
}

func TestMove(t *testing.T) {
	if err := Move("test/dir3", "test/dir2"); err != nil {
		panic(err)
	}
	if err := Move("test/d", "test/c"); err != nil {
		panic(err)
	}
}

func TestRemove(t *testing.T) {
	if err := os.RemoveAll("test"); err != nil {
		panic(err)
	}
}

func TestSCopy(t *testing.T) {

	scpFile("username:password@ipaddress", "")
	scpFile("username:password@ipaddress:123", "")
	scpFile("username:password@ipaddress:123:", "")
	scpFile("username:password@ipaddress:123:/", "")
	scpFile("username:password@ipaddress:", "")

	scpFile("username:password@ipaddress:/", "")
	scpFile("username:password@ipaddress:/path", "")
	scpFile("username:password@ipaddress::/path", "")
	scpFile("username:password@ipaddress:123:/path", "")

}
