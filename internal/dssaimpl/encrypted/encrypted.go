package encrypted

import (
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
)

type EncryptedDssa interface {
	dssa.Dssa
	Underlying() dssa.Dssa
	Msts() metasts.MetaStorageSvc
}

// encryptedDssaImpl implements dssa.Dssa to store data files encrypted
// in underlying dssa
type encryptedDssaImpl struct {
	lgr        *slog.Logger
	underlying dssa.Dssa
	msts       metasts.MetaStorageSvc
}

// Msts implements [EncryptedDssa].
func (ed *encryptedDssaImpl) Msts() metasts.MetaStorageSvc {
	return ed.msts
}

// Underlying implements [EncryptedDssa].
func (ed *encryptedDssaImpl) Underlying() dssa.Dssa {
	return ed.underlying
}

func (ed *encryptedDssaImpl) getDe(path_ string) (*dssa.DataEntry, error) {
	ok, err := ed.msts.Exists(path_)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}
	return ed.msts.Get(path_)
}

func (ed *encryptedDssaImpl) actualPath(de *dssa.DataEntry) string {
	return common.Id2Path(de.Id)
}

// EndSession implements [dssa.Dssa].
func (ed *encryptedDssaImpl) EndSession() error {
	return ed.msts.EndSession()
}

// GetReadCloser implements [dssa.Dssa].
func (ed *encryptedDssaImpl) GetReadCloser(string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// GetWriteCloser implements [dssa.Dssa].
func (ed *encryptedDssaImpl) GetWriteCloser(string) (io.WriteCloser, error) {
	panic("unimplemented")
}

// List implements [dssa.Dssa].
func (ed *encryptedDssaImpl) List(path_ string) ([]*dssa.DataEntry, error) {
	return ed.msts.List(path_)
}

// Mkdir implements [dssa.Dssa].
func (ed *encryptedDssaImpl) Mkdir(de *dssa.DataEntry) error {
	de.IsDir = true // implicit for localfiles
	return ed.msts.Put(de)
}

// NewSession implements [dssa.Dssa].
func (ed *encryptedDssaImpl) NewSession() error {
	return ed.msts.NewSession()
}

// Rm implements [dssa.Dssa].
func (ed *encryptedDssaImpl) Rm(path_ string) error {
	ok, err := ed.msts.Exists(path_)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("encryptedDssaImpl.Rm: %s: no such file or directory", path_)
	}
	de, err := ed.msts.Get(path_)
	if err != nil {
		return err
	}
	if !de.IsDir && !de.IsSymLink {
		if err = ed.underlying.Rm(ed.actualPath(de)); err != nil { // FIXME
			return err
		}
	}
	return ed.msts.Del(path_)
}

// SetStat implements [dssa.Dssa].
func (ed *encryptedDssaImpl) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	ede, err := ed.getDe(de.Path)
	if err != nil {
		return err
	}
	cde := *de
	if ede != nil {
		if noPerm {
			cde.User, cde.UserRights = ede.User, ede.UserRights
			cde.Group, cde.GroupRights = ede.Group, ede.GroupRights
			cde.OtherRights = ede.OtherRights
		}
		if noMtime {
			cde.Mtime = ede.Mtime
		}
	}
	return ed.msts.Put(&cde)
}

// Stat implements [dssa.Dssa].
func (ed *encryptedDssaImpl) Stat(path_ string) (*dssa.DataEntry, error) {
	de, err := ed.getDe(path_)
	if err != nil {
		return nil, err
	}
	if de == nil {
		de = &dssa.DataEntry{Path: path_, Error: fs.ErrNotExist, ErrNotExist: true}
	}
	return de, de.Error
}

// Symlink implements [dssa.Dssa].
func (ed *encryptedDssaImpl) Symlink(old string, new_ string) error {
	de, err := ed.getDe(new_)
	if err != nil {
		return err
	}
	if de != nil {
		return fs.ErrExist
	}
	de = &dssa.DataEntry{
		Path:          new_,
		IsSymLink:     true,
		SymLinkTarget: old,
		Mtime:         time.Now().Unix(),
		User:          os.Getuid(),
		UserRights:    dssa.Rights{Read: true, Write: true, Execute: true},
	}
	return ed.msts.Put(de)
}

var _ dssa.Dssa = &encryptedDssaImpl{}
var _ EncryptedDssa = &encryptedDssaImpl{}
