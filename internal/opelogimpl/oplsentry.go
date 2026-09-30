package opelogimpl

import (
	"errors"

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

func (ose *oplStoredEntry) happyWithSt(stc opelog.StateCode) (yes bool) {
	switch stc {
	case opelog.STC_DONE_ABSENT, opelog.STC_DONE_PRESENT, opelog.STC_DIR_LOAD:
		return true
	case opelog.STC_SE_ERROR:
		if !ose.owo().ClearErrors {
			return true
		}
	case opelog.STC_DIR_RM, opelog.STC_DIR_CHANGE, opelog.STC_DESC_ERROR:
		if !ose.toolRestarted {
			return true
		}
	}
	return
}

func (ose *oplStoredEntry) clearErrorNeeded(stc opelog.StateCode) (yes bool) {
	switch stc {
	case opelog.STC_SE_ERROR:
		if ose.owo().ClearErrors {
			return true
		}
	}
	return
}

func (ose *oplStoredEntry) reloadDirNeeded(stc opelog.StateCode) (yes bool) {
	switch stc {
	case opelog.STC_DIR_RM, opelog.STC_DIR_CHANGE, opelog.STC_DESC_ERROR:
		if ose.toolRestarted {
			return true
		}
	}
	return
}

// remove an entry or empty dir
func (ose *oplStoredEntry) remove() error {
	if err := ose.dssRm(); err != nil {
		ose.setState(false, opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
		return err
	}
	ose.setState(false, opelog.STC_DONE_ABSENT, "", nil, nil, 0)
	return nil
}

// rmDir initiates and/or concludes a recursive dir removal
func (ose *oplStoredEntry) rmDir(theEnd bool) (initiated bool, err error) {
	if !theEnd && len(ose.se().Children) != 0 {
		ose.setState(false, opelog.STC_DIR_RM, "", ose.se(), nil, len(ose.se().Children))
		initiated = true
		ose.childrenQueued = true
		return
	}
	if err = ose.remove(); err != nil {
		return
	}
	return
}

// doLoad is actual load from dss: Stat, and List for dirs
func (ose *oplStoredEntry) doLoad() error {
	se, err := ose.dssStatAndList()
	if err == nil {
		ose.setState(false, opelog.STC_DONE_PRESENT, "", se, nil, 0)
	} else {
		ose.setState(false, opelog.STC_SE_ERROR, err.Error(), se, nil, 0)
		return err
	}
	return nil
}

// load ensures stored entry is fetched with dss and its state is cached
// according to walker's operational context
func (ose *oplStoredEntry) load() error {
	st := ose.getState()
	if st == nil {
		return ose.doLoad()
	}
	if ose.happyWithSt(st.Stc) {
		return nil
	}
	if ose.clearErrorNeeded(st.Stc) {
		return ose.doLoad()
	}
	if ose.reloadDirNeeded(st.Stc) {
		return ose.doLoad()
	}
	return nil
}

func (ose *oplStoredEntry) checkInventory() error {
	return errors.ErrUnsupported
}
