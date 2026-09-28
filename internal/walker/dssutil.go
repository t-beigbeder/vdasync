package walker

import "github.com/t-beigbeder/vdasync/dssa"

// noLstatOnList can be activated for clients that don't need Lstat information
// when Dss implementation doesn't provide it (currentcly sftpc)
func DssList(dss dssa.Dssa, path_ string, noLstatOnList bool) ([]*dssa.DataEntry, error) {
	des, err := dss.List(path_)
	if err != nil {
		return nil, err
	}
	if noLstatOnList {
		return des, err
	}
	upDes := make([]*dssa.DataEntry, len(des))
	for i, de := range des {
		if de.NoLStat {
			upDe, err := dss.Stat(de.Path)
			if err != nil {
				return nil, err
			}
			upDes[i] = upDe
		} else {
			upDes[i] = de
		}
	}
	return upDes, nil
}
