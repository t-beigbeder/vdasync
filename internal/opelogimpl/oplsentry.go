package opelogimpl

import (
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for stored entries
// common services on dss are in opldss

// entry processing updates its state until events can be logged along with their final state
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
	if se.UserRights.Write && (!se.IsDir || se.UserRights.Execute) {
		yes = true
	}
	return
}

// endOrNoOpDirPossible factorized test for DepCount nul (recursive op done) or empty dir
func (ose *oplStoredEntry) endOrNoOpDirPossible() bool {
	return ose.getState().DepCount <= 0 || len(ose.se().Children) == 0
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
	ose.setStatsFor("rm", -1)
	if ose.owo().Dryrun {
		return nil
	}
	if err := ose.dssRm(false); err != nil {
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
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
	if ose.enableWriteNeeded() {
		if err := ose.enableWrite(); err != nil {
			return err
		}
	}
	if !ose.endOrNoOpDirPossible() {
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
	if ose.owo().Dryrun {
		return nil
	}
	if err := ose.dssMkdir(ose.ole.source.se()); err != nil {
		err = fmt.Errorf("createDirOps: mkdir error %s", err)
		ose.setState(opelog.STC_SE_ERROR, err.Error(), nil, nil, 0)
		return err
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
	size := ose.ole.source.se().Size
	sTcss := ose.ole.source.getState().Tcss
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
		ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
	}()
	var (
		css   string
		tTcss [][]byte
		eq    bool
		se    *opelog.StoredEntry
	)
	css, err = ose.dssCopyFile()
	if err != nil {
		return
	}
	tTcss, err = common.Checksums2TypedChecksums(css)
	if err != nil {
		return
	}
	if ose.owo().Check {
		eq, err = common.CompareTcss(tTcss, sTcss, ose.ole.getCsAlgos(false))
		if err != nil {
			return
		}
		if !eq {
			err = errors.New("source/target checksums differ")
			return
		}
	}
	if se, err = ose.setMeta(); err != nil {
		return
	}
	ose.setState(opelog.STC_DONE_PRESENT, "", se, tTcss, 0)
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
	if !ose.isTarget && ose.isPresent() && ose.isRegularFile() && !ose.hasError() && ose.ole.owi.needInvCheck() && len(ose.getState().Tcss) == 0 {
		css, err := ose.dssRead()
		if err != nil {
			ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
			return err
		}
		tCss, err := common.Checksums2TypedChecksums(css)
		if err != nil {
			ose.setState(opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
			return err
		}
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
	if !ose.isDir() || len(ose.se().Children) == 0 {
		// all possible already done
		return nil
	}
	if ose.getState().DepCount != 0 {
		// another action is ongoing, will progress (including -1 meaning that side is done)
		return nil
	}
	// means change actions have been inhibited
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

// processChildrenDone loads children state, set current's accordingly, and propagate to parent
func (ose *oplStoredEntry) processChildrenDone() error {
	var err error
	se := ose.se()
	owi := ose.ole.owi
	errorsNum := 0
	stats := ose.getStats()
	for _, child := range se.Children {
		cle, ok := ose.ole.childrenLeCache[child]
		if !ok {
			cle, err = owi.oplm.GetLogicalEntry(path.Join(ose.ole.relPath, child))
			if err != nil {
				ose.setState(opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
				return err
			}
		}
		cSt := cle.GetState(owi.sessionTime, ose.isTarget)
		if cSt.Stc == opelog.STC_DESC_ERROR || cSt.Stc == opelog.STC_SE_ERROR {
			errorsNum++
		}
		cStats := cle.GetStats(owi.sessionTime)
		stats.SourceListOrStat.Number += cStats.SourceListOrStat.Number
		stats.TargetListOrStat.Number += cStats.TargetListOrStat.Number
		stats.Read.Number += cStats.Read.Number
		stats.Read.Size += cStats.Read.Size
		stats.Create.Number += cStats.Create.Number
		stats.Create.Size += cStats.Create.Size
		stats.Update.Number += cStats.Update.Number
		stats.Update.Size += cStats.Update.Size
		stats.Remove.Number += cStats.Remove.Number
		stats.Remove.Size += cStats.Remove.Size
		stats.MetaChange.Number += cStats.MetaChange.Number
		stats.NoOp.Number += cStats.NoOp.Number
		stats.NoOp.Size += cStats.NoOp.Size
		stats.Error.Number += cStats.Error.Number
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
	if st.DepCount == -1 && !ose.toolRestarted {
		// load children state and propagate to parent
		return ose.processChildrenDone()
	}
	if st.DepCount == -1 {
		// tool restarted implies new parent children cycle
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
