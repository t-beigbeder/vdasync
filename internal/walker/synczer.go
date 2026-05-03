package walker

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

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
		nil,
		&syncDataType{syncOptions: syncOptions, targetDs: targetDs, targetRoot: targetRoot},
	)
}

func syncData(pe *ProcessedEntry) *syncDataType {
	args := pe.Args_()
	if len(args) < 1 {
		return &syncDataType{}
	}
	st, ok := args[0].(*syncDataType)
	if !ok {
		return &syncDataType{}
	}
	return st
}

func syncOptions(pe *ProcessedEntry) *config.SyncOptionsType {
	return syncData(pe).syncOptions
}

func targetDs(pe *ProcessedEntry) dssa.Dssa {
	return syncData(pe).targetDs
}

func targetPath(pe *ProcessedEntry) dssa.Path {
	sd := syncData(pe)
	sr := sd.sourceRoot
	sp := pe.DataEntry.Path
	tr := sd.targetRoot
	tp := make([]string, len(tr))
	copy(tp, tr)
	tp = append(tp, sp[len(sr):]...)
	return tp
}

func onStartDirEntrySync(pe *ProcessedEntry) []*dssa.DataEntry {
	if pe.parent == nil && syncData(pe).sourceRoot == nil {
		syncData(pe).sourceRoot = pe.DataEntry.Path
	}
	tp := targetPath(pe)
	pe.Lgr_().Debug("onStartDirEntrySync: sp, tp", "sp", pe.DataEntry.Path, "tp", tp)
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && !tde.ErrNotExist {
		pe.Error = fmt.Errorf("onStartDirEntrySync: target Stat(%v): %v", tp, err)
		return nil
	}
	if !syncOptions(pe).Dryrun {
		if tde.ErrNotExist {
			tde.UserRights = dssa.Rights{Read: true, Write: true, Execute: true}
			if err = targetDs(pe).Mkdir(tde); err != nil {
				pe.Error = fmt.Errorf("onStartDirEntrySync: Mkdir(%s): %v", tde.Path, err)
				return nil
			}
		} else {
			if !tde.UserRights.Write {
				wtde := *tde
				wtde.UserRights.Write = true
				if err := pe.Dssa_().SetStat(&wtde); err != nil {
					pe.Error = fmt.Errorf("onStartDirEntrySync: SetStat: %v", err)
					return nil
				}
			}
		}
	}
	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		pe.Error = fmt.Errorf("onStartDirEntrySync: List(%v): %v", pe.DataEntry.Path, err)
		return nil
	}
	return des
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
