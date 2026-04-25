//go:build unix

package common

import (
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"os"
	"syscall"
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
