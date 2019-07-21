package fileutils

import (
	"syscall"
	"time"
)

func (f *fileStat) createTime() time.Time {
	stat := f.Sys().(*syscall.Stat_t)

	return time.Unix(0, syscall.TimespecToNsec(stat.Mtim))
}

func (f *fileStat) changeTime() time.Time {
	stat := f.Sys().(*syscall.Stat_t)

	return time.Unix(0, syscall.TimespecToNsec(stat.Ctim))
}

func (f *fileStat) accessTime() time.Time {
	stat := f.Sys().(*syscall.Stat_t)

	return time.Unix(0, syscall.TimespecToNsec(stat.Atim))
}
