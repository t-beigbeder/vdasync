package synpoc

import (
	"log/slog"
	"sync"
	"time"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type process_entry struct {
	de   *dssa.DataEntry
	done func()
}

func process_dnde(log *slog.Logger, gen chan *dssa.DataEntry, pq chan *process_entry, pe *process_entry) {
	isDir := pe.de.IsDir
	if isDir {
		process_dde(log, gen, pq, pe)
	} else {
		process_nde(log, pe)
	}
}

func process_dde(log *slog.Logger, gen chan *dssa.DataEntry, pq chan *process_entry, pe *process_entry) {
	log.Info("process_dde starting", "name", pe.de.Path[0])
	time.Sleep(40 * time.Millisecond)
	ddes, nddes := split_dnd_from(list(gen))
	var wg sync.WaitGroup
	// processing all subdirs in //
	wg.Add(len(ddes))
	for _, dde := range ddes {
		go func() {
			ddone := func() {
				log.Debug("ddone", "ndde", dde.Path[0])
				wg.Done()
			}
			pq <- &process_entry{de: dde, done: ddone}
		}()
	}
	wg.Wait()
	// processing all files in //
	wg.Add(len(nddes))
	for _, ndde := range nddes {
		go func() {
			nddone := func() {
				log.Debug("nddone", "ndde", ndde.Path[0])
				wg.Done()
			}
			pq <- &process_entry{de: ndde, done: nddone}
		}()
	}
	wg.Wait()

	log.Info("GREP process_dde done", "name", pe.de.Path[0])
	time.Sleep(20 * time.Millisecond)
	pe.done()

}

func process_nde(log *slog.Logger, pe *process_entry) {
	log.Info("process_nde starting", "name", pe.de.Path[0])
	time.Sleep(60 * time.Millisecond)
	log.Info("GREP process_nde done", "name", pe.de.Path[0])
	pe.done()
}
