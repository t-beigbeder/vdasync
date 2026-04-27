package localfiles

import (
	"io"
	"io/fs"
	"os"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

type localFiles struct {
}

func osPath(path_ dssa.Path) string {
	lp := path_
	if lp[0] == "" {
		lp[0] = "/"
	}
	return path.Join(path_...)
}

// List implements [dssa.Dssa].
func (d *localFiles) List(path_ dssa.Path) ([]*dssa.DataEntry, error) {
	des, err := os.ReadDir(osPath(path_))
	if err != nil {
		return nil, err
	}
	dtes := []*dssa.DataEntry{}
	for _, de := range des {
		cPath := append(path_, de.Name())
		dte, err := d.Stat(cPath)
		if err != nil {
			dte = &dssa.DataEntry{IsDir: de.IsDir(), Path: cPath, Error: err}
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

// Stat implements [dssa.Dssa].
func (d *localFiles) Stat(path_ dssa.Path) (*dssa.DataEntry, error) {
	fi, err := os.Lstat(osPath(path_))
	isSymlink := fi.Mode().Type()&fs.ModeSymlink != 0
	var linkTarget string
	if isSymlink {
		linkTarget, err = os.Readlink(osPath(path_))
		if err != nil {
			return nil, err
		}
	}
	if err != nil {
		return nil, err
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
func (d *localFiles) SetStat(de *dssa.DataEntry) error {
	path_ := path.Join(osPath(de.Path))
	if err := common.SetAccessRights(
		path_, [2]int{de.User, de.Group},
		[3]dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}); err != nil {
		return err
	}
	if err := common.Lutimes(path_, de.Mtime); err != nil {
		return err
	}
	return nil
}

// GetReader implements [dssa.Dssa].
func (d *localFiles) GetReadCloser(path_ dssa.Path) (io.ReadCloser, error) {
	return os.Open(osPath(path_))
}

// GetWriter implements [dssa.Dssa].
func (d *localFiles) GetWriteCloser(path_ dssa.Path) (io.WriteCloser, error) {
	return os.Create(osPath(path_))
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
