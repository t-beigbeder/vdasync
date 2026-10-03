package opelogimpl

import (
	"errors"
	"slices"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for logical entries
// common services on dss are in opldss

// source and target entries processing updates their states until events can be logged along with their final state
func (ole *oplLogicalEntry) recordEvents() {
	ole.source().recordEvents()
	ole.target().recordEvents()
}

// childrenQueue provides to walker children merged both from source and target
func (ole *oplLogicalEntry) childrenQueue() []string {
	var mChildren []string
	sOse, tOse := ole.source(), ole.target()
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

// doTryRm makes target entry removal progress.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) doTryRm() error {
	tOse := ole.target()
	if !ole.owo().Rm && !ole.owo().Dryrun {
		err := common.ErrNeededRmForbidden
		tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
		return err
	}
	// ok, but won't do anything if entries error
	if ole.target().hasError() {
		return nil
	}
	if ole.source().hasError() {
		err := errors.New("error on source prevents remove")
		tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
		return err
	}
	if tOse.isDir() {
		// whatever the current state, launch rmdir
		tSt := tOse.getState()
		theEnd := (tSt.Stc == opelog.STC_DIR_RM && tSt.DepCount == 0)
		if tOse.rmDirPossible(theEnd) && tOse.enableWriteNeeded() && !tOse.owo().Force {
			err := common.ErrNeededWriteEnableForbidden
			tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
			return err
		}
		if err := tOse.rmDir(theEnd); err != nil {
			return err
		}
		return nil
	}
	if err := tOse.remove(); err != nil {
		return err
	}
	return nil
}

// tryRm makes target entry removal progress if possible.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryRm() error {
	if !ole.owi.impliesGoal("update") {
		return nil
	}
	tOse := ole.target()
	if tOse.isPresent() && ole.parentTSt != nil && ole.parentTSt.Stc == opelog.STC_DIR_RM {
		tSt := tOse.getState()
		if tSt.Stc == opelog.STC_DIR_RM && tSt.DepCount != 0 {
			// shouldn't occur, given the queue management logic
			return nil
		}
		if err := ole.doTryRm(); err != nil {
			return err
		}
	}
	if tOse.isPresent() && (!ole.seEqualType() || tOse.se().IsSymLink) {
		if err := ole.doTryRm(); err != nil {
			return err
		}
	}
	return nil
}

// checkForUpdate assumes s/t both present and checks if they differ
func (ole *oplLogicalEntry) checkForUpdate() bool {
	sOse, tOse := ole.source(), ole.target()
	tSt := tOse.getState()
	if tSt.Stc == opelog.STC_DIR_CHANGE && tSt.DepCount != 0 {
		// shouldn't occur, given the queue management logic
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

// doTryUpdate performs target entry update.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) doTryUpdate() error {
	sOse, tOse := ole.source(), ole.target()
	_ = sOse
	if tOse.isDir() {
		tSt := tOse.getState()
		if tSt.Stc == opelog.STC_DIR_CHANGE && tSt.DepCount != 0 {
			return nil
		}
	}
	return nil
}

// tryUpdate performs target entry update if possible.
//
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryUpdate() error {
	if !ole.owi.impliesGoal("update") {
		return nil
	}
	sOse, tOse := ole.source(), ole.target()
	_, _ = sOse, tOse
	if sOse.isPresent() && tOse.isPresent() && ole.checkForUpdate() {
		if err := ole.doTryUpdate(); err != nil {
			return err
		}
	}

	return nil
}

// tryCreate performs target entry creation if possible.
func (ole *oplLogicalEntry) tryCreate() error {
	if !ole.owi.impliesGoal("create") {
		return nil
	}
	sOse, tOse := ole.source(), ole.target()
	_, _ = sOse, tOse
	return errors.ErrUnsupported
}

// tryChange makes target entry change (create, remove, update) progress if possible.
//
// It takes place after errors have been cleared as far as possible.
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryChange() (err error) {
	sOse, tOse := ole.source(), ole.target()
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

func (ole *oplLogicalEntry) tryLoad() error {
	return errors.ErrUnsupported
}

// process evaluates current states of source and target wrt the walker's
// operational context and performs required actions.
//
// Only errors concerning the walker are notified, stored entry level errors are silenced.
func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")

	// record events in the end as they need meaningful state: StoredEntry, Checksums
	defer ole.recordEvents()

	// initializes state for both stored entries
	// clear errors if possible, (re)start dir loading if possible
	_ = ole.source().load()
	_ = ole.target().load()

	// try performing one or several "change" actions
	if err := ole.tryChange(); err != nil {
		return nil
	}

	// try performing loading actions
	if err := ole.tryLoad(); err != nil {
		return nil
	}
	return nil
}
