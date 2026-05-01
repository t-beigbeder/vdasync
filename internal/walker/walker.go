package walker

import (
	"log/slog"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type Walker interface {
	Run() error
}

type walkerImpl struct {
	lgr *slog.Logger
	concurrency int

	ds dssa.Dssa
	path_ dssa.Path
	args []interface{}
}

var _ Walker = &walkerImpl{}

type ProcessedEntry struct {
	Dssa dssa.Dssa
	Lgr *slog.Logger
	DataEntry   *dssa.DataEntry
	Done func()
}

type Processor func(ProcessedEntry)
func (wi *walkerImpl)Run() error {
	wi.lgr.Info("Run: starting","ds", wi.ds, "args", wi.args)
	pq := make(chan *processEntry, wi.concurrency)
	rootIsDone := make(chan bool)
	done := func() {
		wi.lgr.Debug("test is done")
		rootIsDone <- true
	}
	go func ()  {
		pq <- &processEntry{
			de: &dssa.DataEntry{Path: []}
		}
	}()
	wi.lgr.Info("Run: stopping")
	return nil
}

func MakeWalker(lgr *slog.Logger, concurrency int, ds dssa.Dssa, path_ dssa.Path, args ...interface{}) Walker {
	wi := &walkerImpl{lgr: lgr, ds: ds, args: args}
	return wi
}