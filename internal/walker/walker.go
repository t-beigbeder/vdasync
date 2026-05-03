package walker

import (
	"log/slog"
	"sync"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type Walker interface {
	Run(*dssa.DataEntry) error
}

type ProcessedEntry struct {
	DataEntry, parent *dssa.DataEntry
	UserData          interface{}
	wi                *walkerImpl
	children          []*dssa.DataEntry
	done              func()
}

type EntryLister func(*ProcessedEntry) []*dssa.DataEntry

type EntryProcessor func(*ProcessedEntry)

type walkerImpl struct {
	lgr         *slog.Logger
	concurrency int
	ds          dssa.Dssa

	onStartDirEntry  EntryLister
	onStartNdirEntry EntryProcessor
	onDoneDirs       EntryProcessor
	onDoneFiles      EntryProcessor
	onDoneEntry      EntryProcessor

	args []interface{}

	pq chan *ProcessedEntry
}

var _ Walker = &walkerImpl{}

func MakeWalker(
	lgr *slog.Logger, concurrency int, ds dssa.Dssa,
	onStartDirEntry EntryLister,
	onStartNdirEntry, onDoneDirs, onDoneFiles, onDoneEntry EntryProcessor,
	args ...interface{},
) Walker {
	wi := &walkerImpl{
		lgr:              lgr,
		concurrency:      concurrency,
		ds:               ds,
		onStartDirEntry:  onStartDirEntry,
		onStartNdirEntry: onStartNdirEntry,
		onDoneDirs:       onDoneDirs,
		onDoneFiles:      onDoneFiles,
		onDoneEntry:      onDoneEntry,
		args:             args,
	}
	return wi
}

func (wi *walkerImpl) Run(root *dssa.DataEntry) error {
	wi.lgr.Info("Run: starting", "ds", wi.ds, "args", wi.args)
	wi.pq = make(chan *ProcessedEntry, wi.concurrency)
	rootIsDone := make(chan bool)
	done := func() {
		wi.lgr.Debug("Run is done")
		rootIsDone <- true
	}
	go func() {
		wi.pq <- &ProcessedEntry{
			DataEntry: root,
			wi:        wi,
			done:      done,
		}
	}()

LOOP:
	for {
		select {
		case <-rootIsDone:
			wi.lgr.Info("Run: root is done")
			break LOOP
		case pe := <-wi.pq:
			wi.lgr.Info("Run: pulling", "path", pe.DataEntry.Path, "isDir", pe.DataEntry.IsDir)
			go wi.process(pe)
		}
	}
	wi.lgr.Info("Run: stopping")
	return nil
}

func (wi *walkerImpl) process(pe *ProcessedEntry) {
	isDir := pe.DataEntry.IsDir
	wi.lgr.Info("walker process starting", "entry", pe.DataEntry.Path, "isDir", isDir)
	if isDir {
		wi.processDde(pe)
	} else {
		wi.processNde(pe)
	}
	if wi.onDoneEntry != nil {
		wi.onDoneEntry(pe)
	}
	wi.lgr.Info("walker process done", "entry", pe.DataEntry.Path, "isDir", isDir)
	pe.done()
}

func (wi *walkerImpl) processNde(pe *ProcessedEntry) {
	if wi.onStartNdirEntry != nil {
		wi.onStartNdirEntry(pe)
	}
}

func (wi *walkerImpl) processDde(pe *ProcessedEntry) {
	if wi.onStartDirEntry != nil {
		pe.children = wi.onStartDirEntry(pe)
	} else {
		pe.children = []*dssa.DataEntry{}
	}
	ddes, nddes := splitDndFrom(pe.children)

	var wg sync.WaitGroup

	// processing all subdirs in //
	wg.Add(len(ddes))
	for _, dde := range ddes {
		go func() {
			ddone := func() {
				wg.Done()
			}
			wi.pq <- &ProcessedEntry{DataEntry: dde, parent: pe.DataEntry, wi: wi, done: ddone}
		}()
	}
	wg.Wait()
	if wi.onDoneDirs != nil {
		wi.onDoneDirs(pe)
	}

	// processing all files in //
	wg.Add(len(nddes))
	for _, ndde := range nddes {
		go func() {
			nddone := func() {
				wg.Done()
			}
			wi.pq <- &ProcessedEntry{DataEntry: ndde, parent: pe.DataEntry, wi: wi, done: nddone}
		}()
	}
	wg.Wait()
	if wi.onDoneFiles != nil {
		wi.onDoneFiles(pe)
	}

}

func splitDndFrom(des []*dssa.DataEntry) ([]*dssa.DataEntry, []*dssa.DataEntry) {
	ddes := []*dssa.DataEntry{}
	nddes := []*dssa.DataEntry{}
	for _, dde := range des {
		if dde.IsDir {
			ddes = append(ddes, dde)
		} else {
			nddes = append(nddes, dde)
		}
	}
	return ddes, nddes
}
