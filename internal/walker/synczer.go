package walker

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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
	syncOptions *config.SyncOptionsType
	sourceRoot  dssa.Path
	targetDs    dssa.Dssa
	targetRoot  dssa.Path
}

func NewSynchronizer(
	lgr *slog.Logger, concurrency int,
	syncOptions *config.SyncOptionsType,
	sourceDs dssa.Dssa,
	targetDs dssa.Dssa, targetRoot dssa.Path,
) Walker {
	return MakeWalker(
		lgr,
		concurrency,
		sourceDs,
		onStartDirEntrySync,
		onStartNdirEntrySync,
		nil,
		nil,
		onDoneEntrySync,
		&syncDataType{syncOptions: syncOptions, targetDs: targetDs, targetRoot: targetRoot},
	)
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

func syncRelPath(pe *ProcessedEntry) dssa.Path {
	sd := syncData(pe)
	sr := sd.sourceRoot
	sp := pe.DataEntry.Path
	rp := make([]string, len(sp)-len(sr))
	copy(rp, sp[len(sr):])
	return rp
}

func syncRelSPath(pe *ProcessedEntry) string {
	return path.Join(syncRelPath(pe)...)
}

func targetPath(pe *ProcessedEntry) dssa.Path {
	sd := syncData(pe)
	tr := sd.targetRoot
	tp := make([]string, len(tr))
	copy(tp, tr)
	tp = append(tp, syncRelPath(pe)...)
	return tp
}

func syncRelTargetPath(pe *ProcessedEntry, tde *dssa.DataEntry) dssa.Path {
	sd := syncData(pe)
	tr := sd.targetRoot
	tp := tde.Path
	rp := make([]string, len(tp)-len(tr))
	copy(rp, tp[len(tr):])
	return rp
}

func sourcePath(pe *ProcessedEntry, tde *dssa.DataEntry) dssa.Path {
	sd := syncData(pe)
	sr := sd.sourceRoot
	sp := make([]string, len(sr))
	copy(sp, sr)
	sp = append(sp, syncRelTargetPath(pe, tde)...)
	return sp
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
	pe.Lgr_().Error(message, "dss", sot, "de", syncRelSPath(pe), "err", err)
	syncUserData(pe).Error = err
	return pe.Error
}

func dssInfoSync(pe *ProcessedEntry, isTarget bool, function string) {
	sot := "source"
	if isTarget {
		sot = "target"
	}
	pe.Lgr_().Info(fmt.Sprintf("running dss %s", function), "dss", sot, "de", syncRelSPath(pe))
}

func syncEntryStatusInit(pe *ProcessedEntry) {
	es := &SyncEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Size = pe.DataEntry.Size
	es.Error = pe.Error
	es.relPath = syncRelSPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func onStartDirEntrySync(pe *ProcessedEntry) []*dssa.DataEntry {
	var (
		children []*dssa.DataEntry
		err      error
	)
	if pe.parent == nil && syncData(pe).sourceRoot == nil {
		sd := syncData(pe)
		sd.sourceRoot = pe.DataEntry.Path
	}

	syncEntryStatusInit(pe)

	dssInfoSync(pe, false, "List")
	if children, err = pe.Dssa_().List(pe.DataEntry.Path); err != nil {
		setSyncError(pe, "onStartDirEntrySync: source List", false, err)
		return nil
	}

	if err = prepareTargetDirCreate(pe, children); err != nil {
		return nil
	}

	return children
}

func onStartNdirEntrySync(pe *ProcessedEntry) {
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
	es.AggregatedModChanged = agU
	es.AggregatedError = agE
}

func onDoneEntrySync(pe *ProcessedEntry) {
	if !syncOptions(pe).Dryrun {
		if syncUserData(pe).Created || syncUserData(pe).Updated {
			runSetStatEntrySync(pe)
		}
	}
	es := syncUserData(pe)
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
