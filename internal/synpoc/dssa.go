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
