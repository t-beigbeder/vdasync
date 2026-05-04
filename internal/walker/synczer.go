package walker

import (
	"fmt"
	"io"
	"log/slog"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

type SyncEntryStatus struct {
	relPath                  string
	IsDir                    bool
	Size                     int64
	AggregatedSize           int64
	AggregatedChildrenNumber int
	Created                  bool
	Updated                  bool
	Removed                  bool
	ModChanged               bool
	Error                    error
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
		onDoneFilesSync,
		nil,
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

func syncUserData(pe *ProcessedEntry, de *dssa.DataEntry) *SyncEntryStatus {
	if de == nil {
		de = pe.DataEntry
	}
	es, _ := pe.wi.GetUserData(de).(*SyncEntryStatus)
	return es
}

func setSyncError(pe *ProcessedEntry, message string, err error) error {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "relPath", syncRelSPath(pe))
	syncUserData(pe, nil).Error = err
	return pe.Error
}

func syncEntryStatusInit(pe *ProcessedEntry) {
	es := &SyncEntryStatus{}
	es.IsDir = pe.DataEntry.IsDir
	es.Size = pe.DataEntry.Size
	es.Error = pe.Error
	es.relPath = syncRelSPath(pe)
	pe.wi.SetUserData(pe.DataEntry, es)
}

func prepareTargetDir(pe *ProcessedEntry, sChildren []*dssa.DataEntry) error {
	tp := targetPath(pe)
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && !tde.ErrNotExist {
		return setSyncError(pe, "onStartDirEntrySync: target Stat", err)
	}

	if !syncOptions(pe).Dryrun {
		if tde.ErrNotExist {
			tde.UserRights = dssa.Rights{Read: true, Write: true, Execute: true}
			if err = targetDs(pe).Mkdir(tde); err != nil {
				return setSyncError(pe, "onStartDirEntrySync: target Mkdir", err)
			}
		} else {
			if !tde.UserRights.Write {
				wtde := *tde
				wtde.UserRights.Write = true
				if err := pe.Dssa_().SetStat(&wtde); err != nil {
					return setSyncError(pe, "onStartDirEntrySync: target SetStat", err)
				}
			}
		}
	}
	return nil
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

	if children, err = pe.Dssa_().List(pe.DataEntry.Path); err != nil {
		setSyncError(pe, "onStartDirEntrySync: source List", err)
		return nil
	}

	if err = prepareTargetDir(pe, children); err != nil {
		return nil
	}

	return children
}

func onStartNdirEntrySync(pe *ProcessedEntry) {
	syncEntryStatusInit(pe)

	tp := targetPath(pe)
	pe.Lgr_().Debug("onStartNdirEntrySync: sp, tp", "sp", pe.DataEntry.Path, "tp", tp)
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && !tde.ErrNotExist {
		pe.Error = fmt.Errorf("onStartNdirEntrySync: target Stat(%v): %v", tp, err)
		return
	}
	if !syncOptions(pe).Dryrun {
		rdr, err := pe.wi.ds.GetReadCloser(pe.DataEntry.Path)
		if err != nil {
			pe.Error = fmt.Errorf("onStartNdirEntrySync: GetReadCloser(%v): %v", pe.DataEntry.Path, err)
			return
		}
		defer rdr.Close()
		wrr, err := targetDs(pe).GetWriteCloser(targetPath(pe))
		if err != nil {
			pe.Error = fmt.Errorf("onStartNdirEntrySync: GetWriteCloser(%v): %v", targetPath(pe), err)
			return
		}
		defer wrr.Close()
		_, err = io.Copy(wrr, rdr)
		if err != nil {
			pe.Error = fmt.Errorf("onStartNdirEntrySync: Copy(%v): %v", targetPath(pe), err)
			return
		}
		err = wrr.Close()
		if err != nil {
			pe.Error = fmt.Errorf("onStartNdirEntrySync: Close(%v): %v", targetPath(pe), err)
			return
		}
	}
}

func onDoneFilesSync(pe *ProcessedEntry) {
	ddes, nddes := splitDndFrom(pe.children)
	var (
		agSz int64
		agCN int
	)
	for _, dde := range ddes {
		dud, _ := pe.wi.GetUserData(dde).(*SyncEntryStatus)
		agSz += dud.AggregatedSize
		agCN += dud.AggregatedChildrenNumber
	}
	for _, ndde := range nddes {
		dund, _ := pe.wi.GetUserData(ndde).(*SyncEntryStatus)
		agSz += dund.Size
	}
	es := syncUserData(pe, nil)
	es.AggregatedSize = agSz
	es.AggregatedChildrenNumber = agCN + len(nddes)
	pe.Lgr_().Debug("onDoneFilesRRm", "rp", rmPeRelSPath(pe), "es", es)
}
