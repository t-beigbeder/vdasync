package localfiles

import (
	"os"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

type localFiles struct {
}

// List implements [dssa.Dssa].
func (d *localFiles) List(path_ dssa.Path) ([]*dssa.DataEntry, error) {
	des, err := os.ReadDir(path.Join(path_...))
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
	fi, err := os.Stat(path.Join(path_...))
	if err != nil {
		return nil, err
	}
	ugIds, ugoRights := common.GetAccessRights(fi)
	return &dssa.DataEntry{
		IsDir:       fi.IsDir(),
		Path:        path_,
		Size:        fi.Size(),
		Mtime:       fi.ModTime().Unix(),
		User:        ugIds[0],
		UserRights:  ugoRights[0],
		Group:       ugIds[1],
		GroupRights: ugoRights[1],
		OtherRights: ugoRights[2],
		IsSymLink:   false,
	}, nil
}

// SetStat implements [dssa.Dssa].
func (d *localFiles) SetStat(*dssa.DataEntry) error {
	panic("unimplemented")
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
