package walker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func parentUpdated(pe *ProcessedEntry) {
	if pe.parent == nil || syncUserData(pe.parent).Updated {
		return
	}
	pe.Lgr_().Debug("lock", "pe", pe.DataEntry.Path)
	pe.parent.mx4child.Lock()
	defer pe.Lgr_().Debug("unlock 1", "pe", pe.DataEntry.Path)
	defer pe.parent.mx4child.Unlock()
	defer pe.Lgr_().Debug("unlock 0", "pe", pe.DataEntry.Path)
	syncUserData(pe.parent).Updated = true
}

func prepareTargetDirForUpdate(pe *ProcessedEntry) error {
	if pe.parent == nil {
		return nil
	}
	pTde := syncUserData(pe.parent).targetDe
	if !syncOptions(pe).Dryrun {
		if !pTde.UserRights.Write {
			dssInfoSync(pe.parent, true, "SetStat(\"UserRights.Write\")")
			wtde := *pTde
			wtde.UserRights.Write = true
			if err := targetDs(pe.parent).SetStat(&wtde, false, true); err != nil {
				return setSyncError(pe, "prepareTargetDirForUpdate: SetStat on parent", true, err)
			}
		}
	}
	return nil
}

func isTargetSameKindInSource(pe *ProcessedEntry, sChildren []*dssa.DataEntry, tde *dssa.DataEntry) bool {
	sp := sourcePath(pe, tde)
	for _, sChild := range sChildren {
		if sChild.Path == sp {
			if sChild.IsDir == tde.IsDir {
				if sChild.IsDir {
					return true
				}
				if sChild.IsSymLink == tde.IsSymLink {
					return true
				}
			} else {
				return false
			}
		}
	}
	return false
}

func purgeTargetDirChildren(pe *ProcessedEntry, sChildren []*dssa.DataEntry) error {
	tp := targetPath(pe)
	dssInfoSync(pe, true, "List")
	tdes, err := targetDs(pe).List(tp)
	hasErrors := false
	if err != nil {
		return setSyncError(pe, "purgeTargetDirChildren: List", true, err)
	}
	for _, tde := range tdes {
		if isTargetSameKindInSource(pe, sChildren, tde) {
			continue
		}
		// TODO: 1st pass with dryrun if removal limits set in options
		if syncData(pe).syncOptions.Dryrun {
			continue
		}
		if !syncData(pe).syncOptions.Rm {
			rp := syncRelTargetPath(pe, tde)
			pe.Lgr_().Error("purgeTargetDirChildren: needed rm forbidden", "dss", "target", "de", rp, "err", err)
			hasErrors = true
			continue
		}
		if err = prepareTargetDirForUpdate(pe); err != nil {
			hasErrors = true
			continue
		}
		if tde.IsDir {
			walker, err := RemoveAll(pe.Lgr_(), pe.wi.concurrency/2, targetDs(pe), tde.Path, "target", syncData(pe).syncOptions.Dryrun)
			if err != nil {
				hasErrors = true
				continue
			}
			if walker == nil {
				continue
			}
			ses := syncUserData(pe)
			for _, rmEs := range RmResult(walker) {
				ses.RemovedSize += rmEs.AggregatedSize
				ses.RemovedChildrenNumber += rmEs.AggregatedChildrenNumber
			}
		} else {
			rp := syncRelTargetPath(pe, tde)
			pe.Lgr_().Debug("running dss Rm", "dss", "target", "de", rp)
			if err := targetDs(pe).Rm(tde.Path); err != nil {
				pe.Lgr_().Error("purgeTargetDirChildren: Rm error", "dss", "target", "de", rp, "err", err)
				hasErrors = true
				continue
			}
			ses := syncUserData(pe)
			ses.RemovedSize += tde.Size
			ses.RemovedChildrenNumber += 1
		}
	}
	if hasErrors {
		return fmt.Errorf("purgeTargetDirChildren: some children removal failed")
	}
	return nil
}

func prepareTargetDirCreate(pe *ProcessedEntry, sChildren []*dssa.DataEntry) error {
	tp := targetPath(pe)
	// TODO: optimization if parent has no dte in dryrun
	dssInfoSync(pe, true, "Stat")
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && (!tde.ErrNotExist || pe.parent == nil) {
		return setSyncError(pe, "prepareTargetDirCreate: Stat", true, err)
	}

	if tde.ErrNotExist {
		parentUpdated(pe)
		syncUserData(pe).Created = true
		if err = prepareTargetDirForUpdate(pe); err != nil {
			return err
		}
		if !syncOptions(pe).Dryrun {
			dssInfoSync(pe, true, "Mkdir")
			tde.UserRights = dssa.Rights{Read: true, Write: true, Execute: true}
			if err = targetDs(pe).Mkdir(tde); err != nil {
				return setSyncError(pe, "prepareTargetDirCreate: Mkdir", true, err)
			}
			syncUserData(pe).targetDe = tde
		}
	} else if tde.IsDir {
		syncUserData(pe).targetDe = tde
		if err = purgeTargetDirChildren(pe, sChildren); err != nil {
			return err
		}
	} else {
		// may occur in dryrun, if not should have been removed by parent or else error
		parentUpdated(pe)
		syncUserData(pe).Updated = true
		if !syncOptions(pe).Dryrun {
			return setSyncError(pe, "prepareTargetDirCreate: inconsistent state", true, errors.New("target is not a dir"))
		}
	}
	return nil
}

func fileHasChanges(pe *ProcessedEntry, tde *dssa.DataEntry) (hasChanges bool) {
	defer func() {
		if !hasChanges {
			return
		}
		if pe.Lgr_().Enabled(context.TODO(), slog.LevelDebug) {
			pe.Lgr_().Debug("fileHasChanges", "sde", pe.DataEntry, "tde", tde, "sud", syncUserData(pe))
		}
	}()

	if pe.DataEntry.IsSymLink != tde.IsSymLink {
		hasChanges = true
		return
	}
	if pe.DataEntry.Size != tde.Size {
		hasChanges = true
		return
	}
	if !syncData(pe).syncOptions.NoMtime && pe.DataEntry.Mtime != tde.Mtime {
		hasChanges = true
		return
	}
	if syncData(pe).syncOptions.NoMtime && pe.DataEntry.Mtime > tde.Mtime {
		hasChanges = true
		return
	}
	if pe.DataEntry.SymLinkTarget != tde.SymLinkTarget {
		hasChanges = true
		return
	}
	if !syncData(pe).syncOptions.Check || pe.DataEntry.IsSymLink {
		return
	}

	srdr, err := pe.wi.ds.GetReadCloser(pe.DataEntry.Path)
	if err != nil {
		setSyncError(pe, "fileHasChanges: GetReadCloser", false, err)
		hasChanges = true
		return
	}
	defer srdr.Close()
	syncUserData(pe).sChecksum, err = common.ReaderSha256(srdr)
	if err != nil {
		setSyncError(pe, "fileHasChanges: ReaderSha256", false, err)
		hasChanges = true
		return
	}

	trdr, err := targetDs(pe).GetReadCloser(targetPath(pe))
	if err != nil {
		setSyncError(pe, "fileHasChanges: GetReadCloser", true, err)
		hasChanges = true
		return
	}
	defer trdr.Close()
	syncUserData(pe).tChecksum, err = common.ReaderSha256(trdr)
	if err != nil {
		setSyncError(pe, "fileHasChanges: ReaderSha256", true, err)
		hasChanges = true
		return
	}

	hasChanges = syncUserData(pe).sChecksum != syncUserData(pe).tChecksum
	return
}

func prepareTargetFile(pe *ProcessedEntry, tde *dssa.DataEntry) error {
	if tde.IsSymLink {
		syncUserData(pe).Removed = true
	}
	if pe.DataEntry.IsSymLink {
		syncUserData(pe).Removed = true
	}
	if syncUserData(pe).Removed && !syncData(pe).syncOptions.Dryrun {
		dssInfoSync(pe, true, "Rm")
		if err := targetDs(pe).Rm(tde.Path); err != nil {
			return setSyncError(pe, "prepareTargetFile: Rm", true, err)
		}
	}
	if !syncUserData(pe).Removed && !syncData(pe).syncOptions.Dryrun {
		if !tde.UserRights.Write {
			dssInfoSync(pe, true, "SetStat(\"UserRights.Write\")")
			wtde := *tde
			wtde.UserRights.Write = true
			if err := targetDs(pe).SetStat(&wtde, false, true); err != nil {
				return setSyncError(pe, "prepareTargetFile: SetStat on parent", true, err)
			}
		}
	}
	return nil
}

func runFileEntrySync(pe *ProcessedEntry) error {
	dssInfoSync(pe, false, "GetReadCloser")
	rdr, err := pe.wi.ds.GetReadCloser(pe.DataEntry.Path)
	if err != nil {
		return setSyncError(pe, "runNdirEntrySync: GetReadCloser", false, err)
	}
	defer rdr.Close()
	dssInfoSync(pe, true, "GetWriteCloser")
	wrr, err := targetDs(pe).GetWriteCloser(targetPath(pe))
	if err != nil {
		return setSyncError(pe, "runNdirEntrySync: GetWriteCloser", true, err)
	}
	defer wrr.Close()
	dssInfoSync(pe, true, "Copy source data to target")
	if _, err = io.Copy(wrr, rdr); err != nil {
		return setSyncError(pe, "runNdirEntrySync: Copy", true, err)
	}
	dssInfoSync(pe, true, "Close target")
	err = wrr.Close()
	if err != nil {
		return setSyncError(pe, "runNdirEntrySync: Close", true, err)
	}
	return nil
}

func runSymlinkEntrySync(pe *ProcessedEntry) error {
	dssInfoSync(pe, true, "Symlink")
	if err := targetDs(pe).Symlink(pe.DataEntry.SymLinkTarget, targetPath(pe)); err != nil {
		return setSyncError(pe, "runSymlinkEntrySync: Symlink", true, err)
	}
	return nil
}

func runSetStatEntrySync(pe *ProcessedEntry) error {
	var tde = &dssa.DataEntry{}
	*tde = *pe.DataEntry
	tde.Path = targetPath(pe)
	pe.Lgr_().Debug("runSetStatEntrySync", "sde", pe.DataEntry, "tde", tde)
	dssInfoSync(pe, true, "SetStat")
	if err := targetDs(pe).SetStat(tde, syncData(pe).syncOptions.NoPerm, syncData(pe).syncOptions.NoMtime); err != nil {
		return setSyncError(pe, "runSetStatEntrySync: SetStat", true, err)
	}
	return nil
}

func runNdirEntrySync(pe *ProcessedEntry) {
	tp := targetPath(pe)
	dssInfoSync(pe, true, "Stat")
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && !tde.ErrNotExist {
		setSyncError(pe, "prepareTargetDirCreate: Stat", true, err)
		return
	}
	if tde.IsDir {
		// may occur in dryrun, if not should have been removed by parent or else error
		parentUpdated(pe)
		syncUserData(pe).Updated = true
		if !syncOptions(pe).Dryrun {
			setSyncError(pe, "runNdirEntrySync: inconsistent state", true, errors.New("target is a dir"))
			return
		}
	}

	if tde.ErrNotExist {
		syncUserData(pe).Created = true
	} else {
		syncUserData(pe).targetDe = tde
		if !fileHasChanges(pe, tde) {
			return
		} else {
			if err = prepareTargetFile(pe, tde); err != nil {
				return
			}
			syncUserData(pe).Updated = true
		}
	}
	if !syncOptions(pe).Dryrun {
		if err = prepareTargetDirForUpdate(pe); err != nil {
			return
		}
		if !pe.DataEntry.IsSymLink {
			err = runFileEntrySync(pe)
		} else {
			err = runSymlinkEntrySync(pe)
		}
		if err != nil {
			return
		}
	}
}

func setEntryChanges(pe *ProcessedEntry) {
	es := syncUserData(pe)
	defer func() {
		if !es.ModChanged {
			return
		}
		if pe.Lgr_().Enabled(context.TODO(), slog.LevelDebug) {
			pe.Lgr_().Debug("setEntryChanges", "sde", pe.DataEntry, "tde", es.targetDe, "sud", syncUserData(pe))
		}
	}()
	if es.Error != nil || es.Created || es.Updated || es.ModChanged {
		return
	}
	hasChanges := false
	tde := es.targetDe
	sde := pe.DataEntry
	if !hasChanges && !syncData(pe).syncOptions.NoPerm {
		hasChanges = tde.UserRights != sde.UserRights ||
			tde.GroupRights != sde.GroupRights ||
			tde.OtherRights != sde.OtherRights
	}
	if !hasChanges && !syncData(pe).syncOptions.NoMtime && tde.Mtime != sde.Mtime {
		hasChanges = true
	}
	es.ModChanged = hasChanges
}
