package localfiles

import (
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
	for _, dte := range dtes {
		dte = &dssa.DataEntry{
			IsDir: dte.IsDir,
			Name: dte.Name,
			Size: dte.Size,
			Mtime: dte.Mtime,
			User: dte.User,
			UserRights: dssa.Rights{}, // FIXME
			Group: dte.Group,
			GroupRights: dssa.Rights{}, // FIXME
			OtherRights: dssa.Rights{}, // FIXME
			IsSymLink: dte.IsSymLink,
			SymLinkTarget: "toBeImplemented", // FIXME
		}
		dtes = append(dtes, dte)
	}
	return dtes, nil
}

func MakeLocalFilesDssa() dssa.Dssa {
	return &localFiles{}
}
