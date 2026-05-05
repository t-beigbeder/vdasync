package walker

import (
	"errors"
	"fmt"
	"io"
	"path"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
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
	tde := syncUserData(pe.parent).targetDe
	pud := syncUserData(pe.parent)
	_ = pud
	if !syncOptions(pe).Dryrun {
		if !tde.UserRights.Write {
			dssInfoSync(pe.parent, true, "SetStat(\"UserRights.Write\")")
			wtde := *tde
			wtde.UserRights.Write = true
			if err := targetDs(pe.parent).SetStat(&wtde); err != nil {
				return setSyncError(pe, "prepareTargetDirForUpdate: SetStat on parent", true, err)
			}
		}
	}
	return nil
}

func isTargetInSource(pe *ProcessedEntry, sChildren []*dssa.DataEntry, tde *dssa.DataEntry) bool {
	ssp := path.Join(sourcePath(pe, tde)...)
	for _, sChild := range sChildren {
		if path.Join(sChild.Path...) == ssp {
			if sChild.IsDir == tde.IsDir {
				return true
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
		if isTargetInSource(pe, sChildren, tde) {
			continue
		}
		if err = prepareTargetDirForUpdate(pe); err != nil {
			hasErrors = true
			continue
		}
		if tde.IsDir {
			// TODO: 1st pass with dryrun if removal limits set in options
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
		}
	}
	if hasErrors {
		return fmt.Errorf("purgeTargetDirChildren: some children removal failed")
	}
	return nil
}

func prepareTargetDirCreate(pe *ProcessedEntry, sChildren []*dssa.DataEntry) error {
	tp := targetPath(pe)
	dssInfoSync(pe, true, "Stat")
	tde, err := targetDs(pe).Stat(tp)
	if err != nil && !tde.ErrNotExist {
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
	if !syncOptions(pe).Dryrun {
		if err = prepareTargetDirForUpdate(pe); err != nil {
			return
		}
		dssInfoSync(pe, false, "GetReadCloser")
		rdr, err := pe.wi.ds.GetReadCloser(pe.DataEntry.Path)
		if err != nil {
			setSyncError(pe, "runNdirEntrySync: GetReadCloser", false, err)
			return
		}
		defer rdr.Close()
		dssInfoSync(pe, true, "GetWriteCloser")
		wrr, err := targetDs(pe).GetWriteCloser(targetPath(pe))
		if err != nil {
			setSyncError(pe, "runNdirEntrySync: GetWriteCloser", true, err)
			return
		}
		defer wrr.Close()
		dssInfoSync(pe, true, "Copy source data to target")
		_, err = io.Copy(wrr, rdr)
		if err != nil {
			setSyncError(pe, "runNdirEntrySync: Copy", true, err)
			return
		}
		dssInfoSync(pe, true, "Close target")
		err = wrr.Close()
		if err != nil {
			setSyncError(pe, "runNdirEntrySync: Close", true, err)
			return
		}
	}
}
