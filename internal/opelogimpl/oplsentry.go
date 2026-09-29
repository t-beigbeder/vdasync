package opelogimpl

import (
	"errors"

	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for stored entries
// common services on dss are in opldss

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

func (ose *oplStoredEntry) doLoad() error {
	se, err := ose.dssStatAndList()
	if err == nil {
		ose.setState(false, opelog.STC_DONE_PRESENT, "", se, nil, 0)
	} else {
		ose.setState(false, opelog.STC_DONE_PRESENT, err.Error(), se, nil, 0)
		return err
	}
	return nil
}

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
