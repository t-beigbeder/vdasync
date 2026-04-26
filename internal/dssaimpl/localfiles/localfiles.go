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
		fi, err := os.Stat(de.Name())
		var dte *dssa.DataEntry
		if err == nil {
			ugIds, ugoRights := common.GetAccessRights(fi)
			dte = &dssa.DataEntry{
				IsDir:       de.IsDir(),
				Name:        de.Name(),
				Size:        fi.Size(),
				Mtime:       fi.ModTime().Unix(),
				User:        ugIds[0],
				UserRights:  ugoRights[0],
				Group:       ugIds[1],
				GroupRights: ugoRights[1],
				OtherRights: ugoRights[2],
				IsSymLink:   false,
			}
		} else {
			dte = &dssa.DataEntry{
				IsDir: de.IsDir(),
				Name:  de.Name(),
				Error: err,
			}
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
