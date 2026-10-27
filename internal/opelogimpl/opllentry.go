package opelogimpl

import (
	"path"
	"slices"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for logical entries
// common services on dss are in opldss

// source and target entries processing updates their states until events can be logged along with their final state
func (ole *oplLogicalEntry) recordEvents() {
	ole.source.recordEvents()
	ole.target.recordEvents()
}

// setChildrenQueue provides to walker children merged both from source and target
func (ole *oplLogicalEntry) setChildrenQueue() []string {
	var mChildren []string
	sOse, tOse := ole.source, ole.target
	if sOse.childrenQueued {
		ole.hasChanges = true
		sOse.getState().DepCount = int32(len(sOse.se().Children))
		mChildren = slices.Clone(sOse.se().Children)
	}
	if tOse.childrenQueued {
		ole.hasChanges = true
		tOse.getState().DepCount = int32(len(tOse.se().Children))
		for _, child := range tOse.se().Children {
			if sOse.se() != nil && slices.Contains(sOse.se().Children, child) {
				continue
			}
			mChildren = append(mChildren, child)
		}
	}

	return mChildren
}

func (ole *oplLogicalEntry) loadClientCache(child string) (cle *opelog.LogicalEntry, err error) {
	var ok bool
	if cle, ok = ole.childrenLeCache[child]; ok {
		return cle, nil
	}
	if cle, err = ole.owi.oplm.GetLogicalEntry(path.Join(ole.relPath, child)); cle == nil || err != nil {
		return
	}
	ole.childrenLeCache[child] = cle

	return
}

// propagateChildrenStats loads children stats, propagate to current's accordingly
func (ole *oplLogicalEntry) propagateChildrenStats() error {
	var (
		cle *opelog.LogicalEntry
		err error
	)
	sChildren, tChildren := []string{}, []string{}
	stats := ole.getStats()
	statsInited := ole.statsInited
	propagate := false
	if !statsInited {
		stats = opelog.NewComputedStats()
	}
	if ole.source.se() != nil {
		sChildren = ole.source.se().Children
	}
	if ole.target.se() != nil {
		tChildren = ole.target.se().Children
	}

	for _, child := range sChildren {
		if cle, err = ole.loadClientCache(child); err != nil {
			return err
		}
		if cle == nil {
			continue
		}
		cStats := cle.GetStats(ole.owi.sessionTime)
		if cStats == nil {
			continue
		}
		stats.SourceListOrStat.Number += cStats.SourceListOrStat.Number
		stats.Read.Number += cStats.Read.Number
		stats.Read.Size += cStats.Read.Size
		stats.Error.Number += cStats.Error.Number
		stats.NoOp.Number += cStats.NoOp.Number
		stats.NoOp.Size += cStats.NoOp.Size
		propagate = true
	}

	for _, child := range tChildren {
		if cle, err = ole.loadClientCache(child); err != nil {
			return err
		}
		if cle == nil {
			continue
		}
		cStats := cle.GetStats(ole.owi.sessionTime)
		if cStats == nil {
			continue
		}
		stats.TargetListOrStat.Number += cStats.TargetListOrStat.Number
		stats.Create.Number += cStats.Create.Number
		stats.Create.Size += cStats.Create.Size
		stats.Update.Number += cStats.Update.Number
		stats.Update.Size += cStats.Update.Size
		stats.Remove.Number += cStats.Remove.Number
		stats.Remove.Size += cStats.Remove.Size
		stats.MetaChange.Number += cStats.MetaChange.Number
		if !slices.Contains(sChildren, child) {
			stats.Error.Number += cStats.Error.Number
			stats.NoOp.Number += cStats.NoOp.Number
			stats.NoOp.Size += cStats.NoOp.Size
		}
		propagate = true
	}
	if propagate {
		// dirs stats always up-to-date
		ole.hasChanges = true
	}
	if !stats.IsSet(true) {
		stats.NoOp.Number += 1
	}
	if !statsInited {
		ole.le.SetStats(ole.owi.sessionTime, stats)
	}
	return nil
}

// checkForRemove checks if source absent/target present of if their types differ
func (ole *oplLogicalEntry) checkForRemove() bool {
	sOse := ole.source
	tOse := ole.target
	if sOse.isAbsent() && tOse.isPresent() {
		return true
	}
	if sOse.isPresent() && tOse.isPresent() && !ole.seEqualType() {
		return true
	}
	return false
}

// tryRm makes target entry removal progress if possible.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryRm() error {
	if !ole.owi.impliesGoal("update") {
		return nil
	}
	if !ole.checkForRemove() {
		return nil
	}
	tOse := ole.target
	if !ole.owo().Rm && !ole.owo().Dryrun {
		err := common.ErrNeededRmForbidden
		tOse.setState(opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
		return err
	}
	if tOse.isDir() {
		if tOse.endOrNoOpDirPossible() && tOse.enableWriteNeeded() && !tOse.owo().Force {
			err := common.ErrNeededWriteEnableForbidden
			tOse.setState(opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
			return err
		}
		if err := tOse.rmDir(); err != nil {
			return err
		}
		return nil
	}
	if err := tOse.remove(); err != nil {
		return err
	}
	return nil
}

// checkForUpdate checks if s/t both present and if they differ, incl. checksums if requested
func (ole *oplLogicalEntry) checkForUpdate(noRights bool) bool {
	sOse, tOse, owo := ole.source, ole.target, ole.owo()
	if !sOse.isPresent() || !tOse.isPresent() {
		return false
	}
	seEqual := tOse.se().Equal(sOse.se(), owo.NoEqMtime, owo.NoMtime, owo.NoMtLink, noRights, false)
	if sOse.isDir() || sOse.isSymLink() {
		return !seEqual
	}
	if !seEqual {
		return true
	}
	csAlgos := ole.getCsAlgos(false)
	if csAlgos == "" {
		return false
	}
	eq, err := common.CompareTcss(sOse.getState().Tcss, tOse.getState().Tcss, csAlgos)
	if err != nil {
		return true
	}
	return !eq
}

// tryUpdate performs target entry update if possible.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryUpdate() error {
	if !ole.owi.impliesGoal("update") {
		return nil
	}
	tOse := ole.target
	if !ole.checkForUpdate(false) && !tOse.isDir() {
		return nil
	}
	if tOse.isDir() {
		if err := tOse.updateDirOps(); err != nil {
			return err
		}
		return nil
	}
	if !ole.checkForUpdate(true) {
		se, err := tOse.setMeta()
		if err != nil {
			return err
		}
		tOse.setState(opelog.STC_DONE_PRESENT, "", se, tOse.getState().Tcss, 0)
		return nil
	}
	if tOse.isRegularFile() {
		if err := tOse.copyFile(false); err != nil {
			return err
		}
		return nil
	}
	if tOse.isSymLink() {
		if err := tOse.cloneSymLink(false); err != nil {
			return err
		}
		return nil
	}
	return nil
}

// checkForCreate checks if source present and target absent
func (ole *oplLogicalEntry) checkForCreate() bool {
	sOse, tOse := ole.source, ole.target
	if !sOse.isPresent() || !tOse.isAbsent() {
		return false
	}
	return true
}

// tryCreate performs target entry creation if possible.
func (ole *oplLogicalEntry) tryCreate() error {
	if !ole.owi.impliesGoal("create") {
		return nil
	}
	if !ole.checkForCreate() {
		return nil
	}
	sOse := ole.source
	tOse := ole.target
	if sOse.isDir() {
		// next call will match an update
		if err := tOse.createDirOps(); err != nil {
			return err
		}
		return nil
	}
	if sOse.isRegularFile() {
		if err := tOse.copyFile(true); err != nil {
			return err
		}
		return nil
	}
	if sOse.isSymLink() {
		if err := tOse.cloneSymLink(true); err != nil {
			return err
		}
		return nil
	}
	return nil

}

// tryChange makes target entry change (create, remove, update) progress if possible.
//
// It takes place after errors have been cleared as far as possible.
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryChange() (err error) {
	sOse, tOse := ole.source, ole.target
	if sOse.hasError() || tOse.hasError() {
		// no change allowed, even dryrun uninteresting in that case
		return
	}
	if err = ole.tryRm(); err != nil {
		return
	}
	if err = ole.tryUpdate(); err != nil {
		return
	}
	if sOse.isPresent() && !tOse.isPresent() {
		if err = ole.tryCreate(); err != nil {
			return
		}
	}
	return nil
}

// tryLoad makes target entry load progress if possible.
func (ole *oplLogicalEntry) tryLoad() error {
	if !ole.owi.impliesGoal("load") {
		return nil
	}
	sOse, tOse := ole.source, ole.target
	sErr := sOse.tryLoad()
	tErr := tOse.tryLoad()
	if sErr != nil || tErr != nil {
		return &LeError{source: sErr, target: tErr}
	}
	return nil
}

// process evaluates current states of source and target wrt the walker's
// operational context and performs required actions.
//
// Only errors concerning the walker are notified, stored entry level errors are silenced.
func (ole *oplLogicalEntry) process() error {
	ole.lgr.Debug("process: start")
	if !ole.isIncluded() {
		ole.detail("logical entry is ignored")
		ole.le.IsIgnored = true
		ole.hasChanges = true
		return nil
	}
	if ole.owi.needInvCheck() {
		iSt := ole.le.GetState(ole.owi.invTime, false)
		if iSt == nil || iSt.Se == nil {
			ole.source.setState(opelog.STC_SE_ERROR, "not in inventory", nil, nil, 0)
			ole.hasChanges = true
			return nil
		}
	}

	// record events in the end as they need meaningful state: StoredEntry, Checksums
	defer ole.recordEvents()

	// initializes state for both stored entries
	// clear errors if possible, (re)start dir loading if necessary
	_ = ole.source.load()
	_ = ole.target.load()

	// try performing one or several "change" actions
	if err := ole.tryChange(); err != nil {
		return nil
	}

	// try performing loading actions
	if err := ole.tryLoad(); err != nil {
		return nil
	}

	// if nothing (incl. stat or list) done on any side, set NoOp stats
	if !ole.getStats().IsSet(true) && ole.source.se() != nil {
		size := ole.source.se().Size
		if ole.source.isDir() {
			size = 0
		}
		ole.source.setStatsFor("no", size)
	}

	return nil
}
