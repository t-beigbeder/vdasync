//go:build !unix

package common

import (
	"fmt"
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

func SetAccessRights(path string, ugIds [2]int, ugoRights [3]dssa.Rights) error {
	var (
		mode os.FileMode
		err  error
	)
	ur := ugoRights[0]
	if ur.Read {
		mode |= 1 << 8
	}
	if ur.Write {
		mode |= 1 << 7
	}
	if ur.Execute {
		mode |= 1 << 6
	}
	if err = os.Chmod(path, mode); err != nil {
		return fmt.Errorf("in SetAccessRights: %v", err)
	}
	return nil
}

func Lutimes(path string, mtime int64) error {
	return nil
}
