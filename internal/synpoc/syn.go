package synpoc

import (
	"log/slog"
	"sync"
	"time"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

func process_dir(
	log *slog.Logger,
	gen chan *dssa.DataEntry,
	processing_list chan *dssa.DataEntry,
	done func(),
	cde *dssa.DataEntry) {

	log.Info("process_dir", "name", cde.Name)
	processing_list <- true
	defer func() {
		defer done()
		<-processing_list
	}()

	ddes, nddes := split_dnd_from(list(gen))

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(ddes))
		wg.Wait()
		go func() {
			for _, dde := range ddes {
				continue
			}
		}()

	}()

	time.Sleep(10 * time.Millisecond)

}
