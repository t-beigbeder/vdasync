package opelogimpl

import (
	"errors"
	"fmt"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for stored entries
// common services on dss are in opldss

// stCheckNeeded checks if regular file has to be read to compute checksums
//
// if checksum is already present, a previous operation (create/update) already did it,
// if not, a full read is needed (again, if new session)
func (ose *oplStoredEntry) stCheckNeeded() bool {
	css, _ :=	common.TypedChecksums2Checksums(ose.getState().Tcss)
	if ose.owo().Check && !common.HasAlgos(css, ose.ole.getCsAlgos(true)) {
		return true
	}
	return false
}

// recordEvents ensures entry processing updates its state until events can be logged along with their final state
func (ose *oplStoredEntry) recordEvents() {
	if ose.loadTime != 0 {
		ose.createEvent(ose.loadTime, opelog.EVC_LOADED, ose.getState().Se, ose.getState().Tcss)
	}
	if ose.removeTime != 0 {
		ose.createEvent(ose.removeTime, opelog.EVC_REMOVED, ose.getState().Se, ose.getState().Tcss)
	}
	if ose.createTime != 0 {
		ose.createEvent(ose.createTime, opelog.EVC_CREATED, ose.getState().Se, ose.getState().Tcss)
	}
	if ose.updateTime != 0 {
		ose.createEvent(ose.updateTime, opelog.EVC_UPDATED, ose.getState().Se, ose.getState().Tcss)
	}
	if ose.metaChangeTime != 0 {
		ose.createEvent(ose.metaChangeTime, opelog.EVC_META_CHANGED, ose.getState().Se, ose.getState().Tcss)
	}
}

// enableWriteNeeded assumes cache state loaded and checks if write (and execute for dirs) access is enabled
func (ose *oplStoredEntry) enableWriteNeeded() (yes bool) {
	se := ose.se()
	if !se.UserRights.Write || (se.IsDir && !se.UserRights.Execute) {
		yes = true
	}
	return
}

// endOrNoOpDirPossible factorized test for DepCount nul (recursive op done) or empty dir
func (ose *oplStoredEntry) endOrNoOpDirPossible() bool {
	// FIXME: if DepCount has been reset, shall we loop in all circumstances?
	return (ose.getState().DepCount <= 0 && !ose.toolRestarted) || len(ose.se().Children) == 0
}

// enableWrite enables an entry to be written if needed
func (ose *oplStoredEntry) enableWrite() error {
	if !ose.enableWriteNeeded() {
		return nil
	}
	if ose.owo().Dryrun {
		return nil
	}
	se := ose.se().Clone()
	se.UserRights.Write = true
	if se.IsDir {
		se.UserRights.Execute = true
	}
	if err := ose.dssSetStat(se, false, true, true); err != nil {
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
	}
	return nil
}

// remove an entry or empty dir, just update stats in dryrun
func (ose *oplStoredEntry) remove() error {
	se := ose.se()
	size := se.Size
	if se.IsDir {
		size = 0
	}
	ose.setStatsFor("rm", size)
	if ose.owo().Dryrun {
		return nil
	}
	if err := ose.dssRm(false); err != nil {
		ose.setState(opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
		return err
	}
	ose.setState(opelog.STC_DONE_ABSENT, "", nil, nil, 0)
	return nil
}

// rmDir initiates and/or concludes a recursive dir removal
func (ose *oplStoredEntry) rmDir() error {
	if ose.enableWriteNeeded() {
		if err := ose.enableWrite(); err != nil {
			return err
		}
	}
	if !ose.endOrNoOpDirPossible() {
		ose.childrenQueued = true
		return nil
	}
	if err := ose.remove(); err != nil {
		return err
	}
	return nil
}

// setMeta sets an entry metadata from its source
func (ose *oplStoredEntry) setMeta() (*opelog.StoredEntry, error) {
	// TODO: set stats "mc" if end of dir update or simple meta adjust
	if ose.owo().Dryrun {
		return nil, nil
	}
	se := ose.ole.source.se().Clone()
	owo := ose.owo()
	noMtime := owo.NoMtime || (se.IsSymLink && owo.NoMtLink)
	if err := ose.dssSetStat(se, owo.NoPerm, noMtime, false); err != nil {
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
		return nil, err
	}
	if noMtime {
		se.Mtime = time.Now().Unix()
	}
	if owo.NoPerm {
		tSe := ose.se()
		se.User = tSe.User
		se.UserRights = tSe.UserRights.Clone()
		se.Group = tSe.Group
		se.GroupRights = tSe.GroupRights.Clone()
		se.OtherRights = tSe.OtherRights.Clone()
	}
	return se, nil
}

// updateDirOps initiates and/or concludes a recursive dir update
//
// when actually doing also sets meta and updates state
func (ose *oplStoredEntry) updateDirOps() error {
	if !ose.endOrNoOpDirPossible() {
		if err := ose.enableWrite(); err != nil {
			return err
		}
		ose.childrenQueued = true
		return nil
	}
	if ose.owo().Dryrun {
		return nil
	}
	se, err := ose.setMeta()
	if err != nil {
		return err
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, nil, 0)
	return nil
}

// createDirOps initiates a recursive dir creation, conclusion by updateDirOps
//
// when creating also sets meta and updates state
func (ose *oplStoredEntry) createDirOps() error {
	ose.setStatsFor("cr", 0)
	if ose.owo().Dryrun {
		return nil
	}
	rwSe := ose.ole.source.se().Clone()
	rwSe.Mtime = time.Now().Unix()
	if ose.ole.source.enableWriteNeeded() && len(rwSe.Children) > 0 {
		rwSe.UserRights.Write = true
		rwSe.UserRights.Execute = true
	}
	if err := ose.dssMkdir(rwSe); err != nil {
		err = fmt.Errorf("createDirOps: mkdir error %s", err)
		ose.setState(opelog.STC_SE_ERROR, err.Error(), nil, nil, 0)
		return err
	}
	if len(rwSe.Children) > 0 {
		ose.setState(opelog.STC_DONE_PRESENT, "", rwSe, nil, 0)
		return nil
	}
	se, err := ose.setMeta()
	if err != nil {
		return err
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, nil, 0)
	return nil
}

// copyFile copies file data from its source, just update stats in dryrun
//
// when copying also sets meta and updates state
func (ose *oplStoredEntry) copyFile(isCreated bool) (err error) {
	if !isCreated && !ose.owo().Dryrun && ose.enableWriteNeeded() {
		if err := ose.enableWrite(); err != nil {
			return err
		}
	}
	size := ose.ole.source.se().Size
	sTcss := ose.ole.source.getState().Tcss
	var tCss [][]byte

	ose.setStatsFor("rd", size)
	if !isCreated {
		ose.setStatsFor("up", size)
	} else {
		ose.setStatsFor("cr", size)
	}
	if ose.owo().Dryrun {
		return nil
	}

	defer func() {
		if err == nil {
			return
		}
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), tCss, 0)
	}()
	var (
		css string
		eq  bool
		se  *opelog.StoredEntry
	)
	css, err = ose.dssCopyFile()
	if err != nil {
		return
	}
	tCss, err = common.Checksums2TypedChecksums(css)
	if err != nil {
		return
	}
	if ose.owo().Check {
		if len(sTcss) > 0 {
			eq, err = common.CompareTcss(tCss, sTcss, ose.ole.getCsAlgos(false))
			if err != nil {
				return
			}
			if !eq {
				err = errors.New("source/target checksums differ")
				return
			}
		}
		if ose.ole.owi.needInvCheck() {
			iSt := ose.ole.le.GetState(ose.ole.owi.invTime, false)
			eq, err = common.CompareTcss(tCss, iSt.Tcss, ose.ole.owi.owo.InvCsAlgos)
			if err != nil {
				return
			}
			if !eq {
				err = errors.New("inventory/source checksums differ")
				return
			}
		}
		sOse := ose.ole.source
		if len(sTcss) == 0 {
			sOse.setState(opelog.STC_DONE_PRESENT, "", sOse.se(), tCss, 0)
		}
	}
	if se, err = ose.setMeta(); err != nil {
		return
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, tCss, 0)
	return
}

// cloneSymLink clones symlink from its source, just update stats in dryrun
func (ose *oplStoredEntry) cloneSymLink(isCreated bool) error {
	ose.setStatsFor("mc", 0)
	if ose.owo().Dryrun {
		return nil
	}
	if !isCreated {
		if err := ose.dssRm(true); err != nil {
			err = fmt.Errorf("cloneSymLink: rm error %s", err)
			ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
		}
	}
	if err := ose.dssSymLink(ose.ole.source.se().SymLinkTarget); err != nil {
		err = fmt.Errorf("cloneSymLink: symlink error %s", err)
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
		return err
	}
	se, err := ose.setMeta()
	if err != nil {
		return err
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, nil, 0)
	return nil
}

// tryLoad makes source and/or target entry load progress if possible.
func (ose *oplStoredEntry) tryLoad() error {
	if !ose.isPresent() {
		return nil
	}
	if ose.isRegularFile() && !ose.hasError() && ose.stCheckNeeded() {
		css, err := ose.dssRead()
		if err != nil {
			ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
			return err
		}
		ose.setStatsFor("rd", ose.getState().Se.Size)
		tCss, err := common.Checksums2TypedChecksums(css)
		if err != nil {
			ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
			return err
		}
		if ose.ole.owi.needInvCheck() && !ose.isTarget {
			iSt := ose.ole.le.GetState(ose.ole.owi.invTime, false)
			eq, err := common.CompareTcss(tCss, iSt.Tcss, ose.ole.owi.owo.InvCsAlgos)
			if err != nil {
				ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
				return err
			}
			if !eq {
				err = errors.New("inventory/source checksums differ")
				ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
				return err
			}
		}
		if ose.isTarget {
			sOse := ose.ole.source
			if sOse != nil && sOse.isPresent() && !sOse.hasError() {
				sSt := sOse.getState()
				eq, err := common.CompareTcss(tCss, sSt.Tcss, ose.ole.getCsAlgos(false))
				if !eq {
					err = errors.New("source/target checksums differ")
					ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
					return err
				}
			}
		}
		ose.setState(opelog.STC_DONE_PRESENT, "", ose.se(), tCss, 0)
		return nil
	}
	if !ose.isDir() || len(ose.se().Children) == 0 {
		// all possible already done
		return nil
	}
	if ose.getState().DepCount != 0 {
		// another action is ongoing, will progress (including -1 meaning that side is done)
		return nil
	}
	// DepCount nul means change actions have been inhibited
	// (initial state, or reset after tool (re)start)
	ose.childrenQueued = true
	return nil
}

// doLoad is actual load from dss: Stat, and List for dirs
func (ose *oplStoredEntry) doLoad() error {
	se, err := ose.dssStatAndList(false)
	if err == nil {
		if ose.isTarget {
			ose.setStatsFor("tls", 0)
		} else {
			ose.setStatsFor("sls", 0)
		}
		if se == nil {
			ose.setState(opelog.STC_DONE_ABSENT, "", se, nil, 0)
			return nil
		}
		owi := ose.ole.owi
		if !ose.isTarget && owi.needInvCheck() {
			iSt := ose.ole.le.GetState(owi.invTime, false)
			iSe := iSt.Se
			if !iSe.Equal(se, false, owi.owo.NoMtime, owi.owo.NoMtLink, true, true) {
				err = errors.New("inventory metadata differ")
				ose.setState(opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
				return err
			}
		}
		ose.setState(opelog.STC_DONE_PRESENT, "", se, nil, 0)
	} else {
		ose.setState(opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
		return err
	}
	return nil
}

func (ose *oplStoredEntry) processChildrenDone() error {
	var (
		cle *opelog.LogicalEntry
		err error
	)
	se := ose.se()
	owi := ose.ole.owi
	errorsNum := 0
	for _, child := range se.Children {
		if cle, err = ose.ole.loadClientCache(child); err != nil {
			ose.setState(opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
			return err
		}
		cSt := cle.GetState(owi.sessionTime, ose.isTarget)
		if cSt.Stc == opelog.STC_DESC_ERROR || cSt.Stc == opelog.STC_SE_ERROR {
			errorsNum++
		}
	}
	if errorsNum > 0 {
		ose.setState(opelog.STC_DESC_ERROR, "", se, nil, -1)
		return nil
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, nil, -1)
	return nil
}

// load ensures stored entry is fetched with dss and its state is cached
// according to walker's operational context
func (ose *oplStoredEntry) load() error {
	st := ose.getState()
	if st == nil {
		return ose.doLoad()
	}
	if ose.toolRestarted && !ose.hasError() && !ose.owo().KeepCached {
		// by default full rescan of dirs/files (not recommended in case of errors)
		return ose.doLoad()
	}
	if st.DepCount == -1 && !ose.toolRestarted {
		// load children state and propagate to current
		return ose.processChildrenDone()
	}
	if st.DepCount == -1 {
		// tool restarted implies new parent children cycle
		// resetting DepCount on non-empty dir is a signal to walk again
		// cf se.tryLoad
		st.DepCount = 0
		ose.ole.hasChanges = true
	}
	if st.Stc == opelog.STC_SE_ERROR || st.Stc == opelog.STC_DESC_ERROR {
		if ose.toolRestarted && ose.owo().ClearErrors {
			return ose.doLoad()
		}
	}
	return nil
}
