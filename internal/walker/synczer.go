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
	IsDir          bool
	Size           int64
	AggregatedSize int64
	Created        bool
	Updated        bool
	Removed        bool
	ModChanged     bool
	Error          error
}

type syncDataType struct {
	syncOptions *config.SyncOptionsType
	sourceRoot  dssa.Path
	targetDs    dssa.Dssa
	targetRoot  dssa.Path
	syncResult  map[string]*SyncEntryStatus
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
		nil,
		&syncDataType{syncOptions: syncOptions, targetDs: targetDs, targetRoot: targetRoot},
	)
}

func SyncResult(walker Walker) map[string]*SyncEntryStatus {
	wi, ok := walker.(*walkerImpl)
	if !ok || len(wi.args) < 1 {
		return nil
	}
	syncData, ok := wi.args[0].(*syncDataType)
	if !ok {
		return nil
	}
	return syncData.syncResult
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

func targetPath(pe *ProcessedEntry) dssa.Path {
	sd := syncData(pe)
	tr := sd.targetRoot
	tp := make([]string, len(tr))
	copy(tp, tr)
	tp = append(tp, syncRelPath(pe)...)
	return tp
}

func setSyncResult(pe *ProcessedEntry, es *SyncEntryStatus) {
	if es == nil {
		es = &SyncEntryStatus{}
		es.IsDir = pe.DataEntry.IsDir
		es.Size = pe.DataEntry.Size
		es.Error = pe.Error
	}
	syncData(pe).syncResult[path.Join(pe.DataEntry.Path...)] = es
}

func setSyncError(pe *ProcessedEntry, message string, err error) error {
	pe.Error = fmt.Errorf("%s: %v", message, err)
	pe.Lgr_().Error(message, "relPath", syncRelPath(pe))
	setSyncResult(pe, nil)
	return pe.Error
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
		sd.syncResult = map[string]*SyncEntryStatus{}
	}

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
