package localfiles

import (
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type localFiles struct {
}

// List implements [dssa.Dssa].
func (d *localFiles) List(path_ string) ([]*dssa.DataEntry, error) {
	des, err := os.ReadDir(path_)
	if err != nil {
		return nil, err
	}
	dtes := []*dssa.DataEntry{}
	for _, de := range des {
		cPath := path.Join(path_, de.Name())
		dte, err := d.Stat(cPath)
		if err != nil {
			dte = &dssa.DataEntry{IsDir: de.IsDir(), Path: cPath, Error: err}
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

// Mkdir implements [dssa.Dssa].
func (d *localFiles) Mkdir(de *dssa.DataEntry) error {
	ugor := []dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}
	err := os.Mkdir(de.Path, common.Rights2Mod([3]dssa.Rights(ugor)))
	return err
}

// Stat implements [dssa.Dssa].
func (d *localFiles) Stat(path_ string) (*dssa.DataEntry, error) {
	fi, err := os.Lstat(path_)
	if err != nil {
		return &dssa.DataEntry{Path: path_, Error: err, ErrNotExist: os.IsNotExist(err)}, err
	}
	isSymlink := fi.Mode().Type()&fs.ModeSymlink != 0
	var linkTarget string
	if isSymlink {
		linkTarget, err = os.Readlink(path_)
		if err != nil {
			return nil, err
		}
	}
	ugIds, ugoRights := common.GetAccessRights(fi)
	return &dssa.DataEntry{
		IsDir:         fi.IsDir(),
		Path:          path_,
		Size:          fi.Size(),
		Mtime:         fi.ModTime().Unix(),
		User:          ugIds[0],
		UserRights:    ugoRights[0],
		Group:         ugIds[1],
		GroupRights:   ugoRights[1],
		OtherRights:   ugoRights[2],
		IsSymLink:     isSymlink,
		SymLinkTarget: linkTarget,
	}, nil
}

// SetStat implements [dssa.Dssa].
func (d *localFiles) SetStat(de *dssa.DataEntry, noPerm, noMtime bool) error {
	if !noPerm && !de.IsSymLink {
		if err := common.SetAccessRights(
			de.Path, [2]int{de.User, de.Group},
			[3]dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}); err != nil {
			return err
		}
	}
	if !noMtime {
		if err := common.Lutimes(de.Path, de.Mtime); err != nil {
			return err
		}
	}
	return nil
}

// GetReader implements [dssa.Dssa].
func (d *localFiles) GetReadCloser(path_ string) (io.ReadCloser, error) {
	return os.Open(path_)
}

// GetWriter implements [dssa.Dssa].
func (d *localFiles) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	return os.Create(path_)
}

// Rm implements [dssa.Dssa].
func (d *localFiles) Rm(path_ string) error {
	return os.Remove(path_)
}

// Symlink implements [dssa.Dssa].
func (d *localFiles) Symlink(old string, new_ string) error {
	return os.Symlink(old, new_)
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
