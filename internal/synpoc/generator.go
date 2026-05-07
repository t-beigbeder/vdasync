package synpoc

import (
	"fmt"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

func data_entry_generator(max int) chan *dssa.DataEntry {
	generator := make(chan *dssa.DataEntry)

	go func() {
		for count := 0; count < max; count++ {
			generator <- &dssa.DataEntry{
				IsDir: count%3 == 1,
				Path:  fmt.Sprintf("de%03d", count),
			}
		}
		close(generator)
	}()

	return generator
}
