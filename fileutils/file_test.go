package fileutils

import (
	"fmt"
	"log"
	"runtime"
	"testing"
	"time"
)

func TestReadDir(t *testing.T) {
	dir := "./folder"
	list, err := ReadDir(dir)
	if err != nil {
		panic(err)
	}

	const format = "2006-01-02 15:04:05"

	fmt.Println("Permissions Size       IsDir              Name              CreateTime          AccessTime          ModifyTime          ChangeTime")
	for _, file := range list {
		fmt.Printf("%s  %010d %5v %30s ", file.Mode(), file.Size(), file.IsDir(), file.Name())
		fmt.Printf("%s %s %s %s\n", file.CreateTime().Format(format), file.AccessTime().Format(format), file.ModifyTime().Format(format), file.ChangeTime().Format(format))
	}
}

func testFile(name string) {
	fi, err := Stat(name)
	if err != nil {
		log.Println(name, "error:", err.Error())
	} else {
		fmt.Println(name, ":", fi.Name(), "location:", fi.Location())
	}
}

func TestStat(t *testing.T) {
	testFile("")
	testFile(".")
	testFile("./")
	testFile("../")
	testFile("../../")
	testFile("/")
	testFile("f:")
	testFile("f:/")
	testFile("f:\\")

	return
}

func TestStat2(t *testing.T) {
	//testFile("")
	//testFile(".")
	//testFile("./")
	//testFile("../")
	//testFile("../../")
	//testFile("/")
	//testFile("f:")
	//testFile("f:/")
	//testFile("f:\\")
	//
	//return
	var dir string
	switch runtime.GOOS {
	case "windows":
		dir = "f:/"
	default:
		dir = "/mnt/f/"
	}
	fi, err := Stat(dir)
	if err != nil {
		panic(err)
	}
	fmt.Println("name:", fi.Name())
	fmt.Println("location:", fi.Location())
	fmt.Println("size:", fi.Size())

	start := time.Now()

	files, err := fi.Files()
	if err != nil {
		panic(err)
	}
	fmt.Println("files:", files)

	dirs, err := fi.Dirs()
	if err != nil {
		panic(err)
	}
	fmt.Println("dirs:", dirs)

	size, err := fi.SizeAll()
	if err != nil {
		panic(err)
	}
	fmt.Println("size:", size-fi.Size())

	end := time.Now()
	fmt.Println("use time:", end.UnixNano()-start.UnixNano(), end.Unix()-start.Unix())

	fmt.Println("----------------  FDS  -----------------")
	start = time.Now()
	f, d, s, err := fi.FDS()
	if err != nil {
		panic(err)
	}

	fmt.Println("files:", f)
	fmt.Println("dirs:", d)
	fmt.Println("size:", s-fi.Size())

	end = time.Now()
	fmt.Println("use time:", end.UnixNano()-start.UnixNano(), end.Unix()-start.Unix())
}

func TestStat3(t *testing.T) {
	testFile("")
	testFile(".")
	testFile("./")
	testFile("../")
	testFile("../../")
	testFile("/")
	testFile("f:")
	testFile("f:/")
	testFile("f:\\")
}

func TestFileStat_SizeAll(t *testing.T) {
	files, err := ReadDir("C:/")
	if err != nil {
		panic(err)
	}

	for _, v := range files {
		fmt.Println(v.Name(), v.Size(), v.IsDir())
	}
}

func TestMd5sum(t *testing.T) {
	md5, err := Md5sum("./file_test.go")
	if err != nil {
		panic(err)
	}

	fmt.Println(md5)
}
