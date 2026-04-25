package localfiles

import (
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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
		if err != nil {
			return nil, fmt.Errorf("List: Stat(%s) error: %s", de.Name(), err)
		}
		dte := &dssa.DataEntry{
			IsDir:       de.IsDir(),
			Name:        de.Name(),
			Size:        fi.Size(),
			Mtime:       fi.ModTime().Unix(),
			User:        -1,            // FIXME
			UserRights:  dssa.Rights{}, // FIXME
			Group:       -1,
			GroupRights: dssa.Rights{},
			OtherRights: dssa.Rights{},
			IsSymLink:   false,
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
