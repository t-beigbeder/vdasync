//go:build !unix

package common

import (
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"os"
	"runtime"
)

func GetAccessRights(fi os.FileInfo) ([2]int, [3]dssa.Rights) {
	if runtime.GOOS != "windows" {
		panic("GetAccessRights not uxlike was only tested on windows")
	}
	perm := fi.Mode().Perm()
	ugIds := [2]int{os.Getuid(), os.Getgid()}
	rg := dssa.Rights{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0}
	ugoRights := [3]dssa.Rights{rg, rg, rg}
	return ugIds, ugoRights
}
