package opelogimpl

import (
	"errors"
	"slices"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for logical entries
// common services on dss are in opldss

// tryChange makes target entry change (create, remove, update) progress if possible
//
// Errors are only returned when other or further actions are not possible
func (ole *oplLogicalEntry) tryChange() (bool, error) {
	updAble, creAble := ole.owi.impliesGoal("update"), ole.owi.impliesGoal("create")
	_ = creAble
	sOse, tOse := ole.source(), ole.target()

	if sOse.hasError() || tOse.hasError() {
		return false, nil
	}
	if sOse.isPresent() && tOse.isPresent() {
		if !ole.seEqualType() {
			if !updAble {
				return false, nil
			}
			if tOse.isDir() {
				if !ole.owo().Rm {
					err := common.ErrNeededRmDisabled
					tOse.setState(false, opelog.STC_SE_ERROR, err.Error(), tOse.se(), nil, 0)
					return false, err
				}
				// whatever the current state, launch rmdir
				initiated, err := tOse.rmDir(false)
				if err != nil {
					return false, err
				}
				if initiated {
					return true, nil
				}
				// can try next action, eg update from file
				return false, nil
			}
		}
	}
	return false, errors.ErrUnsupported
}

func (ole *oplLogicalEntry) tryCreate() (bool, error) {
	return false, errors.ErrUnsupported
}

func (ole *oplLogicalEntry) tryLoad() (bool, error) {
	return false, errors.ErrUnsupported
}

func (ole *oplLogicalEntry) recordEvents() {
	ole.source().recordEvents()
	ole.target().recordEvents()
}

// childrenQueue provides to walker children merged both from source and target
func (ole *oplLogicalEntry) childrenQueue() []string {
	var mChildren []string
	sOse, tOse := ole.source(), ole.target()
	if sOse.childrenQueued {
		mChildren = slices.Clone(sOse.se().Children)
	}
	if tOse.childrenQueued {
		for _, child := range tOse.se().Children {
			if slices.Contains(sOse.se().Children, child) {
				continue
			}
			mChildren = append(mChildren, child)
		}
	}
	return mChildren
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
	_ = ole.source().load()
	_ = ole.target().load()

	// try performing actions starting with the most rights' demanding
	for !done && err == nil {
		done, err = ole.tryChange()
		// can remove and then update
	}
	for !done && err == nil {
		done, err = ole.tryCreate()
		// can remove and then update
	}
	return nil
}
