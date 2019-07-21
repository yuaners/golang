package fileutils

import (
	"syscall"
	"time"
)

func (f *fileStat) createTime() time.Time {
	stat := f.Sys().(*syscall.Win32FileAttributeData)
	ns := stat.CreationTime.Nanoseconds()

	return time.Unix(0, ns)
}

func (f *fileStat) changeTime() time.Time {
	stat := f.Sys().(*syscall.Win32FileAttributeData)
	ns := stat.LastWriteTime.Nanoseconds()

	return time.Unix(0, ns)
}

func (f *fileStat) accessTime() time.Time {
	stat := f.Sys().(*syscall.Win32FileAttributeData)
	ns := stat.LastAccessTime.Nanoseconds()

	return time.Unix(0, ns)
}
