//go:build unix

package common

import (
	"fmt"
	"os"
	"syscall"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"golang.org/x/sys/unix"
)

func GetAccessRights(fi os.FileInfo) ([2]int, [3]dssa.Rights) {
	st := fi.Sys().(*syscall.Stat_t)
	perm := fi.Mode().Perm()
	ugIds := [2]int{int(st.Uid), int(st.Gid)}
	ugoRights := [3]dssa.Rights{
		{Read: perm&(1<<8) != 0, Write: perm&(1<<7) != 0, Execute: perm&(1<<6) != 0},
		{Read: perm&(1<<5) != 0, Write: perm&(1<<4) != 0, Execute: perm&(1<<3) != 0},
		{Read: perm&(1<<2) != 0, Write: perm&(1<<1) != 0, Execute: perm&(1) != 0},
	}
	return ugIds, ugoRights
}

func SetAccessRights(path string, ugIds [2]int, ugoRights [3]dssa.Rights) error {
	var (
		mode os.FileMode
		err  error
	)
	ur, gr, or := ugoRights[0], ugoRights[1], ugoRights[2]
	if ur.Read {
		mode |= 1 << 8
	}
	if ur.Write {
		mode |= 1 << 7
	}
	if ur.Execute {
		mode |= 1 << 6
	}
	if gr.Read {
		mode |= 1 << 5
	}
	if gr.Write {
		mode |= 1 << 4
	}
	if gr.Execute {
		mode |= 1 << 3
	}
	if or.Read {
		mode |= 1 << 2
	}
	if or.Write {
		mode |= 1 << 1
	}
	if or.Execute {
		mode |= 1
	}
	if err = os.Chmod(path, mode); err != nil {
		return fmt.Errorf("in SetAccessRights: %v", err)
	}
	uid, gid := ugIds[0], ugIds[1]
	if os.Geteuid() == 0 {
		if err = os.Lchown(path, uid, gid); err != nil {
			return fmt.Errorf("in SetAccessRights: %v", err)
		}
	}
	return nil
}

func Lutimes(path string, mtime int64) error {
	return unix.Lutimes(path, []unix.Timeval{unix.NsecToTimeval(time.Now().UnixNano()), unix.NsecToTimeval(mtime * 1e9)})
}
