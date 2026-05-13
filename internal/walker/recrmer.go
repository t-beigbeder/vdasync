package walker

import (
	"fmt"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
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
	BaseDoerData
	dssAlias   string
	dryRun     bool
	sourceRoot string
}

func NewRecursiveRemover(
	lgr *slog.Logger, concurrency int,
	dss dssa.Dssa,
	dssAlias string,
	dryRun bool,
) Walker {
	return MakeWalker(
		lgr,
		concurrency,
		dss,
		onStartDirEntryRRm,
		onStartNdirEntryRRm,
		nil,
		onDoneFilesRRm,
		onDoneEntryRRm,
		&rmDataType{dssAlias: dssAlias, dryRun: dryRun, BaseDoerData: BaseDoerData{DoerLabel: "rRm"}},
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

func rmPeRelPath(pe *ProcessedEntry) string {
	return common.RelPath(pe.DataEntry.Path, rmData(pe).sourceRoot)
}

func rmUserData(pe *ProcessedEntry) *RmEntryStatus {
	es, _ := pe.wi.GetUserData(pe.DataEntry).(*RmEntryStatus)
	return es
}

func setRmError(pe *ProcessedEntry, message string, err error) {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "dss", rmData(pe).dssAlias, "relPath", rmPeRelPath(pe), "err", err)
	rmUserData(pe).Error = err
}

func rmEntryStatusInit(pe *ProcessedEntry) {
	es := &RmEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Size = pe.DataEntry.Size
	es.Error = pe.Error
	es.relPath = rmPeRelPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func dssInfoRRm(pe *ProcessedEntry, function string) {
	pe.Lgr_().Info(fmt.Sprintf("running dss %s", function), "alias", rmData(pe).dssAlias, "de", rmPeRelPath(pe))
}

func onStartDirEntryRRm(pe *ProcessedEntry) []*dssa.DataEntry {
	if pe.parent == nil && rmData(pe).sourceRoot == "" {
		sd := rmData(pe)
		sd.sourceRoot = pe.DataEntry.Path
	}

	rmEntryStatusInit(pe)

	dssInfoRRm(pe, "List")
	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		setRmError(pe, "onStartDirEntryRRm: List", err)
		return nil
	}
	return des
}

func onStartNdirEntryRRm(pe *ProcessedEntry) {
	rmEntryStatusInit(pe)
}

func onDoneFilesRRm(pe *ProcessedEntry) {
	ddes, nddes := splitDndFrom(pe.children)
	var (
		agSz int64
		agCN int
	)
	for _, dde := range ddes {
		dud, _ := pe.wi.GetUserData(dde).(*RmEntryStatus)
		agSz += dud.AggregatedSize
		agCN += dud.AggregatedChildrenNumber
	}
	for _, ndde := range nddes {
		dund, _ := pe.wi.GetUserData(ndde).(*RmEntryStatus)
		agSz += dund.Size
	}
	es := rmUserData(pe)
	es.AggregatedSize = agSz
	es.AggregatedChildrenNumber = agCN + len(nddes)
}

func onDoneEntryRRm(pe *ProcessedEntry) {
	rmUserData(pe).Size = pe.DataEntry.Size
	if !rmData(pe).dryRun {
		dssInfoRRm(pe, "Rm")
		err := pe.Dssa_().Rm(pe.DataEntry.Path)
		if err != nil {
			setRmError(pe, "onDoneEntryRRm: Rm", err)
			return
		}
	}
}

func RemoveAll(lgr *slog.Logger, concurrency int, dss dssa.Dssa, path_ string, dssAlias string, dryRun bool) (Walker, error) {
	walker := NewRecursiveRemover(lgr, concurrency, dss, dssAlias, dryRun)
	de, err := dss.Stat(path_)
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
	hasErrors := false
	for _, es := range RmResult(walker) {
		if es.Error != nil {
			hasErrors = true
		}
	}
	if hasErrors {
		err = fmt.Errorf("RemoveAll: some children removal failed")
	}
	return walker, err
}
