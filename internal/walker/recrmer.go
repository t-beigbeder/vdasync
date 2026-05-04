package walker

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type RmEntryStatus struct {
	relPath                  string
	IsDir                    bool
	Size                     int64
	AggregatedSize           int64
	AggregatedChildrenNumber int
	Error                    error
}

type rmDataType struct {
	dryRun     bool
	sourceRoot dssa.Path
}

func NewRecursiveRemover(
	lgr *slog.Logger, concurrency int,
	sourceDs dssa.Dssa,
	dryRun bool,
) Walker {
	return MakeWalker(
		lgr,
		concurrency,
		sourceDs,
		onStartDirEntryRRm,
		onStartNdirEntryRRm,
		nil,
		onDoneFilesRRm,
		onDoneEntryRRm,
		&rmDataType{dryRun: dryRun},
	)
}

func RmResult(walker Walker) map[string]*RmEntryStatus {
	result := map[string]*RmEntryStatus{}
	walker.UserDataMap().Range(func(key, value any) bool {
		es, _ := value.(*RmEntryStatus)
		if es != nil {
			result[es.relPath] = es
		}
		return true
	})
	return result
}

func rmData(pe *ProcessedEntry) *rmDataType {
	args := pe.Args_()
	if len(args) < 1 {
		return &rmDataType{}
	}
	rd, ok := args[0].(*rmDataType)
	if !ok {
		return &rmDataType{}
	}
	return rd
}

func rmDeRelSPath(pe *ProcessedEntry, de *dssa.DataEntry) string {
	rd := rmData(pe)
	sr := rd.sourceRoot
	sp := de.Path
	rp := make([]string, len(sp)-len(sr))
	copy(rp, sp[len(sr):])
	return path.Join(rp...)
}

func rmPeRelSPath(pe *ProcessedEntry) string {
	return rmDeRelSPath(pe, pe.DataEntry)
}

func rmUserData(pe *ProcessedEntry, de *dssa.DataEntry) *RmEntryStatus {
	if de == nil {
		de = pe.DataEntry
	}
	es, _ := pe.wi.GetUserData(de).(*RmEntryStatus)
	return es
}

func setRmError(pe *ProcessedEntry, message string, err error) error {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "relPath", rmPeRelSPath(pe))
	rmUserData(pe, nil).Error = err
	return pe.Error
}

func entryStatusInit(pe *ProcessedEntry) {
	es := &RmEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Size = pe.DataEntry.Size
	es.Error = pe.Error
	es.relPath = rmPeRelSPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func onStartDirEntryRRm(pe *ProcessedEntry) []*dssa.DataEntry {
	if pe.parent == nil && rmData(pe).sourceRoot == nil {
		sd := rmData(pe)
		sd.sourceRoot = pe.DataEntry.Path
	}

	entryStatusInit(pe)

	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		setRmError(pe, "onStartDirEntryRRm: List", err)
		return nil
	}
	ddes, nddes := splitDndFrom(des)
	pe.Lgr_().Debug("onStartDirEntryRRm", "rp", rmPeRelSPath(pe), "des", len(des), "ddes", len(ddes), "nddes", len(nddes))
	return des
}

func onStartNdirEntryRRm(pe *ProcessedEntry) {
	entryStatusInit(pe)
}

func onDoneFilesRRm(pe *ProcessedEntry) {
	ddes, nddes := splitDndFrom(pe.children)
	pe.Lgr_().Debug("onDoneFilesRRm", "rp", rmPeRelSPath(pe), "pe.children", len(pe.children), "ddes", len(ddes), "nddes", len(nddes))
	var (
		agSz int64
		agCN int
	)
	for _, dde := range ddes {
		pe.Lgr_().Debug("onDoneFilesRRm: 1", "rp", rmPeRelSPath(pe), "crp", rmDeRelSPath(pe, dde), "cSt", rmUserData(pe, dde))
		dud, _ := pe.wi.GetUserData(dde).(*RmEntryStatus)
		agSz += dud.AggregatedSize
		agCN += dud.AggregatedChildrenNumber
	}
	for _, ndde := range nddes {
		dund, _ := pe.wi.GetUserData(ndde).(*RmEntryStatus)
		agSz += dund.Size
	}
	es := rmUserData(pe, nil)
	es.AggregatedSize = agSz
	es.AggregatedChildrenNumber = agCN + len(nddes)
	pe.Lgr_().Debug("onDoneFilesRRm: 2", "rp", rmPeRelSPath(pe), "es", es)
}

func onDoneEntryRRm(pe *ProcessedEntry) {
	rmUserData(pe, nil).Size = pe.DataEntry.Size
	if !rmData(pe).dryRun {
		err := pe.Dssa_().Rm(pe.DataEntry.Path)
		if err != nil {
			setRmError(pe, "onDoneEntryRRm: Rm", err)
			return
		}
	}
}

func RemoveAll(lgr *slog.Logger, concurrency int, ds dssa.Dssa, path_ dssa.Path, dryRun bool) (Walker, error) {
	walker := NewRecursiveRemover(lgr, concurrency, ds, dryRun)
	de, err := ds.Stat(path_)
	if err != nil {
		return nil, err
	}
	err = walker.Run(de)
	if err != nil {
		return nil, err
	}
	if de.Error != nil {
		return nil, de.Error
	}
	return walker, nil
}
