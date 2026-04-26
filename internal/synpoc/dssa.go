package synpoc

import "github.com/t-beigbeder/otvl_dtacsy/dssa"

func list(generator chan *dssa.DataEntry) []*dssa.DataEntry {
	count := 0
	des := []*dssa.DataEntry{}
	for de := range generator {
		des = append(des, de)
		count++
		if count >= 5 {
			break
		}
	}
	return des
}

func split_dnd_from(des []*dssa.DataEntry) ([]*dssa.DataEntry, []*dssa.DataEntry) {
	ddes := []*dssa.DataEntry{}
	nddes := []*dssa.DataEntry{}
	for _, dde := range ddes {
		if dde.IsDir {
			ddes = append(ddes, dde)
		} else {
			nddes = append(nddes, dde)
		}
	}
	return ddes, nddes
}
