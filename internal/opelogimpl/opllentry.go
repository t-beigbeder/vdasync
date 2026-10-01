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

// tryRm makes target entry removal progress if possible.
//
// Returns true if the change is fully done, if not a remove may be followed by a creation.
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryRm() (bool, error) {
	if !ole.owi.impliesGoal("update") {
		return false, nil
	}
	tOse := ole.target()
	if tOse.isDir() {
		if !ole.owo().Rm {
			err := common.ErrNeededRmDisabled
			tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
			return false, err
		}
		// whatever the current state, launch rmdir
		tSt := tOse.getState()
		initiated, err := tOse.rmDir(tSt.Stc == opelog.STC_DIR_RM && tSt.DepCount == 0)
		if err != nil {
			return false, err
		}
		if initiated {
			return true, nil
		}
		// can try next action, eg update from file
		return false, nil
	}
	if err := tOse.remove(); err != nil {
		return false, err
	}
	return false, nil
}

// tryCreate performs target entry update if possible.
func (ole *oplLogicalEntry) tryUpdate() (bool, error) {
	if !ole.owi.impliesGoal("update") {
		return false, nil
	}
	sOse, tOse := ole.source(), ole.target()
	_, _ = sOse, tOse
	return false, errors.ErrUnsupported
}

// tryCreate performs target entry creation if possible.
func (ole *oplLogicalEntry) tryCreate() (bool, error) {
	if !ole.owi.impliesGoal("create") {
		return false, nil
	}
	sOse, tOse := ole.source(), ole.target()
	_, _ = sOse, tOse
	return false, errors.ErrUnsupported
}

// tryChange makes target entry change (create, remove, update) progress if possible.
//
// It takes place after possible errors have been cleared.
// Returns true if the change is fully done, if not a remove may be followed by a creation.
// Errors are only returned when other or further actions are not possible.
func (ole *oplLogicalEntry) tryChange() (bool, error) {
	sOse, tOse := ole.source(), ole.target()
	if sOse.hasError() || tOse.hasError() {
		return false, nil
	}
	if sOse.isPresent() && tOse.isPresent() {
		if ole.parentTSt != nil && ole.parentTSt.Stc == opelog.STC_DIR_RM {
			return ole.tryRm()
		}
		if !ole.seEqualType() {
			return ole.tryRm()
		}
		return ole.tryUpdate()
	}
	if !sOse.isPresent() {
		return false, nil
	}
	return ole.tryCreate()
}

func (ole *oplLogicalEntry) tryLoad() (bool, error) {
	return false, errors.ErrUnsupported
}

// process evaluates current states of source and target wrt the walker's
// operational context and performs required actions.
//
// Only errors concerning the walker are notified, stored entry level errors are silenced.
func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	var (
		done bool
		err  error
	)
	// record events in the end as they need meaningful state: StoredEntry, Checksums
	defer ole.recordEvents()

	// initializes state for both stored entries
	// clear errors if possible, (re)start dir loading if possible
	_ = ole.source().load()
	_ = ole.target().load()

	// try performing "change" actions, starting with the most rights' demanding
	for !done && err == nil {
		done, err = ole.tryChange()
		// looping because
		// can remove and then update
		// can create and then change mode
	}

	// if no change done, try to perform loading actions
	for !done && err == nil {
		done, err = ole.tryLoad()
		// possible?
	}
	return nil
}
