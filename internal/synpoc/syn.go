package synpoc

import (
	"log/slog"
	"sync"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

func process_dir(
	log *slog.Logger,
	gen chan *dssa.DataEntry,
	processing_list chan *dssa.DataEntry,
	done func(),
	de *dssa.DataEntry) {

	defer done()
	ddes := list(gen)
	wg := sync.WaitGroup

	for _, dde := range ddes {
		continue
	}
}
