package localfiles

import (
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"os"
	"path"
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
		dte := &dssa.DataEntry{
			IsDir: de.IsDir(),
			Name:  de.Name(),
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
