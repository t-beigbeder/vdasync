package walker

import (
	"fmt"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type DoerEntryStatus struct {
	relPath string
	IsDir   bool
	Error   error
}

type doerDataType struct {
	dssAlias   string
	sourceRoot string
	doerLabel  string
}

func NewRecursiveDoer(
	lgr *slog.Logger, concurrency int,
	dss dssa.Dssa,
	dssAlias string,
	doerLabel string,
	onDoneEntry EntryProcessor,
) Walker {
	return MakeWalker(
		lgr,
		concurrency,
		dss,
		onStartDirEntryRDoer,
		onStartNdirEntryRDoer,
		nil,
		nil,
		onDoneEntry,
		&doerDataType{dssAlias: dssAlias, doerLabel: doerLabel},
	)
}

func DoerResult(walker Walker) map[string]*DoerEntryStatus {
	result := map[string]*DoerEntryStatus{}
	walker.UserDataMap().Range(func(key, value any) bool {
		es, _ := value.(*DoerEntryStatus)
		if es != nil {
			result[es.relPath] = es
		}
		return true
	})
	return result
}

func doerData(pe *ProcessedEntry) *doerDataType {
	args := pe.Args_()
	if len(args) < 1 {
		return &doerDataType{}
	}
	rd, ok := args[0].(*doerDataType)
	if !ok {
		return &doerDataType{}
	}
	return rd
}

func doerPeRelPath(pe *ProcessedEntry) string {
	return common.RelPath(pe.DataEntry.Path, doerData(pe).sourceRoot)
}

func doerUserData(pe *ProcessedEntry) *DoerEntryStatus {
	es, _ := pe.wi.GetUserData(pe.DataEntry).(*DoerEntryStatus)
	return es
}

func setDoerError(pe *ProcessedEntry, message string, err error) {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "dss", doerData(pe).dssAlias, "relPath", doerPeRelPath(pe), "err", err)
	doerUserData(pe).Error = err
}

func doerEntryStatusInit(pe *ProcessedEntry) {
	es := &DoerEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Error = pe.Error
	es.relPath = doerPeRelPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func dssInfoRDoer(pe *ProcessedEntry, function string) {
	pe.Lgr_().Info(fmt.Sprintf("running dss %s", function), "doer", doerData(pe).doerLabel, "alias", doerData(pe).dssAlias, "de", doerPeRelPath(pe))
}

func onStartDirEntryRDoer(pe *ProcessedEntry) []*dssa.DataEntry {
	if pe.parent == nil && doerData(pe).sourceRoot == "" {
		sd := doerData(pe)
		sd.sourceRoot = pe.DataEntry.Path
	}

	doerEntryStatusInit(pe)

	dssInfoRDoer(pe, "List")
	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		setDoerError(pe, "onStartDirEntryRDoer: List", err)
		return nil
	}
	return des
}

func onStartNdirEntryRDoer(pe *ProcessedEntry) {
	doerEntryStatusInit(pe)
}

func onDoneEntryChmodRW(pe *ProcessedEntry) {
	pe.DataEntry.UserRights.Write = true
	dssInfoRDoer(pe, "SetStat(\"+w\")")
	err := pe.Dssa_().SetStat(pe.DataEntry, false, true)
	if err != nil {
		setDoerError(pe, "onDoneEntryChmodRW: SetStat", err)
		return
	}
}

func onDoneEntryChmodRO(pe *ProcessedEntry) {
	pe.DataEntry.UserRights.Write = false
	pe.DataEntry.GroupRights.Write = false
	pe.DataEntry.OtherRights.Write = false
	dssInfoRDoer(pe, "SetStat(\"-w\")")
	err := pe.Dssa_().SetStat(pe.DataEntry, false, true)
	if err != nil {
		setDoerError(pe, "onDoneEntryChmodRO: SetStat", err)
		return
	}
}

func onDoneEntryMtime(pe *ProcessedEntry, mtime int64) {
	pe.DataEntry.Mtime = mtime
	dssInfoRDoer(pe, "SetStat(mtime)")
	err := pe.Dssa_().SetStat(pe.DataEntry, true, false)
	if err != nil {
		setDoerError(pe, "onDoneEntryMtime: SetStat", err)
		return
	}
}

func RecDoAll(
	lgr *slog.Logger, concurrency int, dss dssa.Dssa, path_ string, dssAlias string,
	doerLabel string, onDoneEntry EntryProcessor) (Walker, error) {
	walker := NewRecursiveDoer(lgr, concurrency, dss, dssAlias, doerLabel, onDoneEntry)
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
	for _, es := range DoerResult(walker) {
		if es.Error != nil {
			hasErrors = true
		}
	}
	if hasErrors {
		err = fmt.Errorf("RecDoAll: some children failed for %s", doerLabel)
	}
	return walker, err
}

func RecChmodRW(lgr *slog.Logger, concurrency int, dss dssa.Dssa, path_ string, dssAlias string) (Walker, error) {
	return RecDoAll(lgr, concurrency, dss, path_, dssAlias, "ChmodRW", onDoneEntryChmodRW)
}

func RecChmodRO(lgr *slog.Logger, concurrency int, dss dssa.Dssa, path_ string, dssAlias string) (Walker, error) {
	return RecDoAll(lgr, concurrency, dss, path_, dssAlias, "ChmodRO", onDoneEntryChmodRO)
}

func RecTouch(lgr *slog.Logger, concurrency int, dss dssa.Dssa, path_ string, dssAlias string, mtime int64) (Walker, error) {
	return RecDoAll(
		lgr, concurrency, dss, path_, dssAlias, "Touch Mtime",
		func(pe *ProcessedEntry) {
			onDoneEntryMtime(pe, mtime)
		},
	)
}
