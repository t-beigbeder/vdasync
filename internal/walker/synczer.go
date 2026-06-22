package walker

import (
	"fmt"
	"log/slog"
	"path"
	"regexp"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type SyncEntryStatus struct {
	relPath                  string
	targetDe                 *dssa.DataEntry
	sChecksum                string
	tChecksum                string
	IsDir                    bool
	Size                     int64
	AggregatedSize           int64
	AggregatedChildrenNumber int
	Created                  bool
	Updated                  bool
	Removed                  bool
	ModChanged               bool
	Error                    error
	AggregatedCreated        int
	AggregatedUpdated        int
	AggregatedRemoved        int
	AggregatedModChanged     int
	AggregatedError          int
	RemovedSize              int64
	RemovedChildrenNumber    int
}

type syncDataType struct {
	BaseDoerData
	syncOptions *config.SyncOptionsType
	sourceRoot  string
	targetDs    dssa.Dssa
	targetRoot  string
	excls       []*regexp.Regexp
	incls       []*regexp.Regexp
}

func compileRe(ss []string) ([]*regexp.Regexp, error) {
	rs := []*regexp.Regexp{}
	for _, s := range ss {
		r, err := regexp.Compile(s)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r)
	}
	return rs, nil
}

func reFromFile(path_ string, label string) ([]*regexp.Regexp, error) {
	if path_ == "" {
		return nil, nil
	}
	res, err := common.FileLines(path_)
	if err != nil {
		return nil, fmt.Errorf("RunSynchronizer: %s %v", label, err)
	}
	return compileRe(res)
}

func NewSynchronizer(
	lgr *slog.Logger, concurrency int,
	syncOptions *config.SyncOptionsType,
	sourceDs dssa.Dssa,
	targetDs dssa.Dssa, targetRoot string,
) (Walker, error) {
	excls, err := reFromFile(syncOptions.ExclListPath, "exclusion file")
	if err != nil {
		return nil, fmt.Errorf("RunSynchronizer: exclusion file regexps: %v", err)
	}
	incls, err := reFromFile(syncOptions.InclListPath, "inclusion file")
	if err != nil {
		return nil, fmt.Errorf("RunSynchronizer: inclusion file regexps: %v", err)
	}
	return MakeWalker(
		lgr,
		concurrency,
		sourceDs,
		onStartDirEntrySync,
		onStartNdirEntrySync,
		nil,
		nil,
		onDoneEntrySync,
		&syncDataType{
			syncOptions: syncOptions, targetDs: targetDs, targetRoot: targetRoot,
			excls: excls, incls: incls,
			BaseDoerData: BaseDoerData{DoerLabel: "sync"},
		},
	), nil
}

func RunSynchronizer(
	lgr *slog.Logger, concurrency int,
	syncOptions *config.SyncOptionsType,
	sourceDs dssa.Dssa, sourceRoot string,
	targetDs dssa.Dssa, targetRoot string,
) (Walker, error) {
	sde, err := sourceDs.Stat(sourceRoot)
	if err != nil {
		return nil, fmt.Errorf("RunSynchronizer: source %v", err)
	}
	if !sde.IsDir {
		return nil, fmt.Errorf("RunSynchronizer: source %s is not a dir", sourceRoot)
	}
	tde, err := targetDs.Stat(targetRoot)
	if err != nil {
		return nil, fmt.Errorf("RunSynchronizer: target %v", err)
	}
	if !tde.IsDir {
		return nil, fmt.Errorf("RunSynchronizer: target %s is not a dir", targetRoot)
	}
	wk, err := NewSynchronizer(lgr, concurrency, syncOptions, sourceDs, targetDs, targetRoot)
	if err != nil {
		return nil, err
	}
	lgr.Info("RunSynchronizer", "concurrency", concurrency, "source", sourceRoot, "target", targetRoot)
	return wk, wk.Run(sde)
}

func SyncResult(walker Walker) map[string]*SyncEntryStatus {
	result := map[string]*SyncEntryStatus{}
	walker.UserDataMap().Range(func(_, value any) bool {
		es, _ := value.(*SyncEntryStatus)
		if es != nil {
			result[es.relPath] = es
		}
		return true
	})
	return result
}

func syncData(pe *ProcessedEntry) *syncDataType {
	args := pe.Args_()
	if len(args) < 1 {
		return &syncDataType{}
	}
	sd, ok := args[0].(*syncDataType)
	if !ok {
		return &syncDataType{}
	}
	return sd
}

func syncOptions(pe *ProcessedEntry) *config.SyncOptionsType {
	return syncData(pe).syncOptions
}

func targetDs(pe *ProcessedEntry) dssa.Dssa {
	return syncData(pe).targetDs
}

func syncRelPath(pe *ProcessedEntry) string {
	return common.RelPath(pe.DataEntry.Path, syncData(pe).sourceRoot)
}

func targetPath(pe *ProcessedEntry) string {
	return path.Join(syncData(pe).targetRoot, syncRelPath(pe))
}

func syncRelTargetPath(pe *ProcessedEntry, tde *dssa.DataEntry) string {
	return common.RelPath(tde.Path, syncData(pe).targetRoot)
}

func sourcePath(pe *ProcessedEntry, tde *dssa.DataEntry) string {
	return path.Join(syncData(pe).sourceRoot, syncRelTargetPath(pe, tde))
}

func syncUserData(pe *ProcessedEntry) *SyncEntryStatus {
	es, _ := pe.wi.GetUserData(pe.DataEntry).(*SyncEntryStatus)
	return es
}

func setSyncError(pe *ProcessedEntry, message string, isTarget bool, err error) error {
	sot := "source"
	if isTarget {
		sot = "target"
	}
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "dss", sot, "de", syncRelPath(pe), "err", err)
	syncUserData(pe).Error = err
	return pe.Error
}

func dssInfoSync(pe *ProcessedEntry, isTarget bool, function string) {
	sot := "source"
	if isTarget {
		sot = "target"
	}
	pe.Lgr_().Info(fmt.Sprintf("running dss %s", function), "dss", sot, "de", syncRelPath(pe))
}

func syncEntryStatusInit(pe *ProcessedEntry) {
	es := &SyncEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Size = pe.DataEntry.Size
	es.Error = pe.Error
	es.relPath = syncRelPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func isIncluded(pe *ProcessedEntry) bool {
	sd := syncData(pe)
	rp := syncRelPath(pe)
	if len(sd.incls) == 0 {
		return false
	}
	for _, ire := range sd.incls {
		if ire.MatchString(rp) {
			return true
		}
	}
	return false
}

func isExcluded(pe *ProcessedEntry) bool {
	sd := syncData(pe)
	rp := syncRelPath(pe)
	for _, ere := range sd.excls {
		if ere.MatchString(rp) {
			ise := !isIncluded(pe)
			if ise {
				pe.Lgr_().Debug("isExcluded: excluded", "ere", ere, "pe", pe.DataEntry.Path, "rp", rp)
				return true
			} else {
				pe.Lgr_().Debug("isExcluded: excluded -> included", "ere", ere, "pe", pe.DataEntry.Path, "rp", rp)
			}
		}
	}
	return false
}

func onStartDirEntrySync(pe *ProcessedEntry, noLstatOnList bool) []*dssa.DataEntry {
	var (
		children []*dssa.DataEntry
		err      error
	)
	if pe.parent == nil && syncData(pe).sourceRoot == "" {
		sd := syncData(pe)
		sd.sourceRoot = pe.DataEntry.Path
	}
	if isExcluded(pe) {
		return nil
	}

	syncEntryStatusInit(pe)

	dssInfoSync(pe, false, "List")
	if children, err = DssList(pe.Dssa_(), pe.DataEntry.Path, pe.wi.noLstatOnList); err != nil {
		setSyncError(pe, "onStartDirEntrySync: source List", false, err)
		return nil
	}

	inclChildren := []*dssa.DataEntry{}
	for _, de := range children {
		if isExcluded(&ProcessedEntry{DataEntry: de, parent: pe, wi: pe.wi}) {
			continue
		}
		inclChildren = append(inclChildren, de)
	}

	if err = prepareTargetDirCreate(pe, inclChildren); err != nil {
		return nil
	}

	return inclChildren
}

func onStartNdirEntrySync(pe *ProcessedEntry) {
	if isExcluded(pe) {
		return
	}

	syncEntryStatusInit(pe)
	runNdirEntrySync(pe)
}

func computeDdeAggregates(pe *ProcessedEntry) {
	ddes, nddes := splitDndFrom(pe.children)
	var (
		agSz int64
		agCN int
		agC  int
		agU  int
		agR  int
		agM  int
		agE  int
	)
	for _, dde := range ddes {
		dud, _ := pe.wi.GetUserData(dde).(*SyncEntryStatus)
		agSz += dud.AggregatedSize
		agCN += dud.AggregatedChildrenNumber + 1
		agC += dud.AggregatedCreated
		agU += dud.AggregatedUpdated
		agR += dud.AggregatedRemoved
		agM += dud.AggregatedModChanged
		agE += dud.AggregatedError
	}
	for _, ndde := range nddes {
		dund, _ := pe.wi.GetUserData(ndde).(*SyncEntryStatus)
		agSz += dund.Size
		agCN += 1
		if dund.Created {
			agC += 1
		}
		if dund.Updated {
			agU += 1
		}
		if dund.Removed {
			agR += 1
		}
		if dund.ModChanged {
			agM += 1
		}
		if dund.Error != nil {
			agE += 1
		}
	}
	es := syncUserData(pe)
	es.AggregatedSize = agSz
	es.AggregatedChildrenNumber = agCN
	if es.Created {
		agC += 1
	}
	if es.Updated {
		agU += 1
	}
	if es.Removed {
		agR += 1
	}
	if es.ModChanged {
		agM += 1
	}
	if es.Error != nil {
		agE += 1
	}
	es.AggregatedCreated = agC
	es.AggregatedUpdated = agU
	es.AggregatedRemoved = agR
	es.AggregatedModChanged = agM
	es.AggregatedError = agE
}

func onDoneEntrySync(pe *ProcessedEntry) {
	if isExcluded(pe) {
		return
	}

	setEntryChanges(pe)
	es := syncUserData(pe)
	if !syncOptions(pe).Dryrun {
		if es.Created || es.Updated || es.ModChanged {
			runSetStatEntrySync(pe)
		}
	}
	if es.Created {
		es.Updated = false
		es.ModChanged = false
	}
	if es.Updated {
		es.ModChanged = false
	}
	if pe.DataEntry.IsDir {
		computeDdeAggregates(pe)
	}
}
