package opelogimpl

import (
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

// setNoOpIf sets stats for NoOp if entries has not been touched
func (ole *oplLogicalEntry) setNoOpIf() {
	stats := ole.source.getStats()
	if stats.IsSet() {
		return
	}
	var size int64
	if ole.source.se() != nil {
		size = ole.source.se().Size
	}
	ole.source.setStatsFor("no", size)
}

// childrenQueue provides to walker children merged both from source and target
func (ole *oplLogicalEntry) childrenQueue() []string {
	var mChildren []string
	sOse, tOse := ole.source, ole.target
	if sOse.childrenQueued {
		sOse.getState().DepCount = int32(len(sOse.se().Children))
		mChildren = slices.Clone(sOse.se().Children)
	}
	if tOse.childrenQueued {
		tOse.getState().DepCount = int32(len(tOse.se().Children))
		for _, child := range tOse.se().Children {
			if slices.Contains(sOse.se().Children, child) {
				continue
			}
			mChildren = append(mChildren, child)
		}
	}

	return mChildren
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
		tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
		return err
	}
	if tOse.isDir() {
		if tOse.endOrNoOpDirPossible() && tOse.enableWriteNeeded() && !tOse.owo().Force {
			err := common.ErrNeededWriteEnableForbidden
			tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
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
func (ole *oplLogicalEntry) checkForUpdate() bool {
	sOse, tOse := ole.source, ole.target
	if !sOse.isPresent() || !tOse.isPresent() {
		return false
	}
	if !tOse.se().Equal(sOse.se(), true, ole.owo().NoMtime, ole.owo().NoMtLink, true) {
		return true
	}
	if ole.owo().CsAlgos == "" {
		return false
	}
	eq, err := common.CompareTcss(sOse.getState().Tcss, tOse.getState().Tcss, ole.owo().CsAlgos)
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
	if !ole.checkForUpdate() {
		return nil
	}
	tOse := ole.target
	if tOse.isDir() {
		if err := tOse.updateDirOps(); err != nil {
			return err
		}
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
	if sOse.isPresent() || tOse.isAbsent() {
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

	// record events in the end as they need meaningful state: StoredEntry, Checksums
	defer ole.recordEvents()

	// initializes state for both stored entries
	// clear errors if possible, (re)start dir loading if possible
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

	// set NoOp stats if nothing done on any side
	ole.setNoOpIf()
	return nil
}
