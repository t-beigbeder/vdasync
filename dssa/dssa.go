package dssa

import (
	"fmt"
	"io"
)

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

func (rights *Rights) String() string {
	rs := "-"
	if rights.Read {
		rs = "r"
	}
	ws := "-"
	if rights.Write {
		ws = "w"
	}
	xs := "-"
	if rights.Execute {
		xs = "x"
	}
	return rs + ws + xs
}

type DataEntry struct {
	IsDir       bool
	Path        string
	Size        int64
	Mtime       int64
	User        int
	UserRights  Rights
	Group       int
	GroupRights Rights
	OtherRights Rights
	// some implementations (sftpc) list entries w/o Lstat information
	// explicit call to Stat (that implements Lstat) on each entry is then required
	// for most operations
	NoLStat       bool
	IsSymLink     bool
	SymLinkTarget string
	Error         error
	ErrNotExist   bool
	Id            string
	AddMeta       []byte
}

func (de *DataEntry) ModString() string {
	tp := "-"
	if de.IsDir {
		tp = "d"
	} else if de.IsSymLink {
		tp = "l"
	}
	return fmt.Sprintf("%s%s%s%s", tp, de.UserRights.String(), de.GroupRights.String(), de.OtherRights.String())
}

type Dssa interface {
	NewSession() error
	EndSession() error
	List(string) ([]*DataEntry, error)
	Mkdir(*DataEntry) error
	Stat(string) (*DataEntry, error)
	SetStat(_ *DataEntry, noPerm, noMtime bool) error
	Checksum(algos, path_ string) (string, error)
	GetReadCloser(string) (io.ReadCloser, error)
	GetWriteCloser(string) (io.WriteCloser, error)
	Rm(string) error
	Symlink(old, new_ string) error
}

type DssaMaker interface {
	MakeDssa(args ...any) (Dssa, error)
}
