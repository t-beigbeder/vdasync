package walker

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type RmEntryStatus struct {
	IsDir                    bool
	Size                     int64
	AggregatedSize           int64
	AggregatedChildrenNumber int
	Error                    error
}

type rmDataType struct {
	dryRun     bool
	sourceRoot dssa.Path
	rmResult   map[string]*RmEntryStatus
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
		nil,
		nil,
		onDoneFilesRRm,
		onDoneEntryRRm,
		&rmDataType{dryRun: dryRun},
	)
}

func RmResult(walker Walker) map[string]*RmEntryStatus {
	wi, ok := walker.(*walkerImpl)
	if !ok || len(wi.args) < 1 {
		return nil
	}
	rmData, ok := wi.args[0].(*rmDataType)
	if !ok {
		return nil
	}
	return rmData.rmResult
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

func rmDeRelPath(pe *ProcessedEntry, de *dssa.DataEntry) dssa.Path {
	rd := rmData(pe)
	sr := rd.sourceRoot
	sp := de.Path
	rp := make([]string, len(sp)-len(sr))
	copy(rp, sp[len(sr):])
	return rp
}

func rmPeRelPath(pe *ProcessedEntry) dssa.Path {
	return rmDeRelPath(pe, pe.DataEntry)
}

func setRmResult(pe *ProcessedEntry, es *RmEntryStatus) *RmEntryStatus {
	if es == nil {
		es = &RmEntryStatus{}
		es.IsDir = pe.DataEntry.IsDir
		es.Size = pe.DataEntry.Size
		es.Error = pe.Error
	}
	rmData(pe).rmResult[path.Join(rmPeRelPath(pe)...)] = es
	return es
}

func setRmError(pe *ProcessedEntry, message string, err error) error {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "relPath", rmPeRelPath(pe))
	setRmResult(pe, nil)
	return pe.Error
}

func onStartDirEntryRRm(pe *ProcessedEntry) []*dssa.DataEntry {
	if pe.parent == nil && rmData(pe).sourceRoot == nil {
		sd := rmData(pe)
		sd.sourceRoot = pe.DataEntry.Path
		sd.rmResult = map[string]*RmEntryStatus{}
	}

	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		setRmError(pe, "onStartDirEntryRRm: List", err)
		return nil
	}
	return des
}

func onDoneFilesRRm(pe *ProcessedEntry) {
	ddes, nddes := splitDndFrom(pe.children)
	var (
		agSz int64
		agCN int
	)
	rr := rmData(pe).rmResult
	for _, dde := range ddes {
		pe.Lgr_().Debug("onDoneFilesRRm: 0", "rp", rmPeRelPath(pe), "crp", rmDeRelPath(pe, dde), "cSt", rr[path.Join(rmDeRelPath(pe, dde)...)])
		agSz += rr[path.Join(rmDeRelPath(pe, dde)...)].AggregatedSize
		agCN += rr[path.Join(rmDeRelPath(pe, dde)...)].AggregatedChildrenNumber
	}
	for _, ndde := range nddes {
		agSz += rr[path.Join(rmDeRelPath(pe, ndde)...)].Size
	}
	es := setRmResult(pe, nil)
	es.AggregatedSize = agSz
	es.AggregatedChildrenNumber = agCN + len(nddes)
	pe.Lgr_().Debug("onDoneFilesRRm: 1", "rp", rmPeRelPath(pe), "es", es)
}

func onDoneEntryRRm(pe *ProcessedEntry) {
	setRmResult(pe, nil).Size = pe.DataEntry.Size
	if rmData(pe).dryRun {
		err := pe.Dssa_().Rm(pe.DataEntry.Path)
		if err != nil {
			setRmError(pe, "onDoneEntryRRm: Rm", err)
			return
		}
	}
}

func RemoveAll(lgr *slog.Logger, concurrency int, ds dssa.Dssa, path_ dssa.Path, dryRun bool) error {
	walker := NewRecursiveRemover(lgr, concurrency, ds, dryRun)
	de, err := ds.Stat(path_)
	if err != nil {
		return err
	}
	err = walker.Run(de)
	if err != nil {
		return err
	}
	if de.Error != nil {
		return de.Error
	}
	return nil
}
