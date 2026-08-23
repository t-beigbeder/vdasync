package dssa

import (
	"log/slog"
	"runtime"
	"sync"
)

type DssBaseDoerData struct {
	DoerLabel string
}

func (bdd *DssBaseDoerData) Label() string { return bdd.DoerLabel }

type DssBaseDoerDataItf interface {
	Label() string
}

type DssWalker interface {
	Run(*DataEntry) error
	SetUserData(*DataEntry, any)
	GetUserData(*DataEntry) any
	UserDataMap() *sync.Map
	SetResult(any)
	GetResult() any
}

type DssProcessedEntry struct {
	DataEntry *DataEntry
	Error     error
	wi        *walkerImpl
	parent    *DssProcessedEntry
	children  []*DataEntry
	mx4child  sync.Mutex
	done      func()
}

func (pe *DssProcessedEntry) Lgr_() *slog.Logger {
	return pe.wi.implLgr
}

func (pe *DssProcessedEntry) Dssa_() Dssa {
	return pe.wi.ds
}

func (pe *DssProcessedEntry) Args_() []any {
	return pe.wi.args
}

type EntryLister func(*DssProcessedEntry, bool) []*DataEntry

type EntryProcessor func(*DssProcessedEntry)

type walkerImpl struct {
	lgr           *slog.Logger
	implLgr       *slog.Logger
	concurrency   int
	ds            Dssa
	noLstatOnList bool

	onStartDirEntry  EntryLister
	onStartNdirEntry EntryProcessor
	onDoneDirs       EntryProcessor
	onDoneFiles      EntryProcessor
	onDoneEntry      EntryProcessor

	args []any

	pq  chan *DssProcessedEntry
	udm *sync.Map

	result any
}

var _ DssWalker = &walkerImpl{}

func MakeWalker(
	lgr *slog.Logger, concurrency int, ds Dssa,
	onStartDirEntry EntryLister,
	onStartNdirEntry, onDoneDirs, onDoneFiles, onDoneEntry EntryProcessor,
	args ...any,
) DssWalker {
	dLabel := "undefined"
	if len(args) > 0 {
		wd, ok := args[0].(DssBaseDoerDataItf)
		if ok {
			dLabel = wd.Label()
		}
	}
	wi := &walkerImpl{
		lgr:              lgr.With("walker", true),
		implLgr:          lgr.With("walkerImpl", dLabel),
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

func (wi *walkerImpl) Run(root *DataEntry) error {
	wi.lgr.Info("Run: starting", "ds", wi.ds, "args", wi.args)
	wi.pq = make(chan *DssProcessedEntry, wi.concurrency+1) // leave room for next demand
	wi.udm = &sync.Map{}

	count := 0
	m := runtime.MemStats{}
	runtime.ReadMemStats(&m)
	wi.lgr.Info("Run: starting", "HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024)

	rootIsDone := make(chan bool)
	tokens := make(chan bool, wi.concurrency+1)
	done := func() {
		wi.lgr.Debug("Run root is done")
		rootIsDone <- true
	}
	go func() {
		wi.pq <- &DssProcessedEntry{
			DataEntry: root,
			wi:        wi,
			done:      done,
		}
	}()

LOOP:
	for {
		select {
		case <-rootIsDone:
			runtime.ReadMemStats(&m)
			wi.lgr.Info("Run: root is done", "count", count, "HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024)
			break LOOP
		case pe := <-wi.pq:
			wi.lgr.Debug("Run: pulling", "path", pe.DataEntry.Path, "isDir", pe.DataEntry.IsDir)
			count++
			if count%1000 == 0 {
				runtime.ReadMemStats(&m)
				wi.lgr.Info("Run: processed...", "count", count, "HeapInuse", m.HeapInuse/1024, "HeapAlloc", m.HeapAlloc/1024, "StackInuse", m.StackInuse/1024)
			}
			tokens <- true
			go func(pe *DssProcessedEntry) {
				wi.lgr.Debug("Concurrent run starting", "pe", pe.DataEntry.Path)
				wi.process(pe)
				wi.lgr.Debug("Concurrent run done", "pe", pe.DataEntry.Path)
				<-tokens
			}(pe)
		}
	}
	wi.lgr.Info("Run: stopping")
	return nil
}

func (wi *walkerImpl) SetUserData(de *DataEntry, ud any) {
	wi.udm.Store(de.Path, ud)
}

func (wi *walkerImpl) GetUserData(de *DataEntry) any {
	ud, _ := wi.udm.Load(de.Path)
	return ud
}

func (wi *walkerImpl) UserDataMap() *sync.Map {
	return wi.udm
}

func (wi *walkerImpl) SetResult(result any) { wi.result = result }

func (wi *walkerImpl) GetResult() any { return wi.result }

func (wi *walkerImpl) process(pe *DssProcessedEntry) {
	isDir := pe.DataEntry.IsDir
	wi.lgr.Debug("walker process starting", "entry", pe.DataEntry.Path, "isDir", isDir)
	if isDir {
		wi.processDde(pe)
	} else {
		wi.processNde(pe)
	}
}

func (wi *walkerImpl) processNde(pe *DssProcessedEntry) {
	if wi.onStartNdirEntry != nil {
		wi.onStartNdirEntry(pe)
	}
	if wi.onDoneEntry != nil {
		wi.onDoneEntry(pe)
	}
	wi.lgr.Debug("walker processNde done", "entry", pe.DataEntry.Path)
	pe.done()
}

func (wi *walkerImpl) processDde(pe *DssProcessedEntry) {
	if wi.onStartDirEntry != nil {
		pe.children = wi.onStartDirEntry(pe, wi.noLstatOnList)
	} else {
		pe.children = []*DataEntry{}
	}
	ddes, nddes := splitDndFrom(pe.children)
	go wi.batchProcessDde(pe, ddes, nddes)
}

func (wi *walkerImpl) batchProcessDde(pe *DssProcessedEntry, ddes, nddes []*DataEntry) {
	var wg sync.WaitGroup

	wg.Add(len(ddes))
	for _, dde := range ddes {
		wi.pq <- &DssProcessedEntry{DataEntry: dde, parent: pe, wi: wi, done: func() {
			wg.Done()
		}}
	}
	wg.Wait()
	if wi.onDoneDirs != nil {
		wi.onDoneDirs(pe)
	}

	wg.Add(len(nddes))
	for _, ndde := range nddes {
		wi.pq <- &DssProcessedEntry{DataEntry: ndde, parent: pe, wi: wi, done: func() {
			wg.Done()
		}}
	}
	wg.Wait()
	if wi.onDoneFiles != nil {
		wi.onDoneFiles(pe)
	}

	if wi.onDoneEntry != nil {
		wi.onDoneEntry(pe)
	}
	wi.lgr.Debug("walker batchProcessDde done", "entry", pe.DataEntry.Path)
	pe.done()

}

func splitDndFrom(des []*DataEntry) ([]*DataEntry, []*DataEntry) {
	ddes := []*DataEntry{}
	nddes := []*DataEntry{}
	for _, dde := range des {
		if dde.IsDir {
			ddes = append(ddes, dde)
		} else {
			nddes = append(nddes, dde)
		}
	}
	return ddes, nddes
}
