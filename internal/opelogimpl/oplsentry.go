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
	return ose.getState().DepCount == 0 || len(ose.se().Children) == 0
}

// enables an entry to be written if needed, just update stats in dryrun
func (ose *oplStoredEntry) enableWrite() error {
	if !ose.enableWriteNeeded() {
		return nil
	}
	se := ose.se()
	ose.setStatsFor("incMc", 1, 0)
	if ose.owo().Dryrun {
		return nil
	}
	se = se.Clone()
	se.UserRights.Write = true
	if se.IsDir {
		se.UserRights.Execute = true
	}
	if err := ose.dssSetStat(se, false, true); err != nil {
		ose.setState(false, opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
	}
	return nil
}

// remove an entry or empty dir, just update stats in dryrun
func (ose *oplStoredEntry) remove() error {
	ose.setStatsFor("setRm", 1, -1)
	if ose.owo().Dryrun {
		return nil
	}
	if err := ose.dssRm(); err != nil {
		ose.setState(false, opelog.STC_SE_ERROR, err.Error(), ose.se(), nil, 0)
		return err
	}
	ose.setState(false, opelog.STC_DONE_ABSENT, "", nil, nil, 0)
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

// updateDirOps initiates and/or concludes a recursive dir update
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
	// if err := ose.setMeta(); err != nil {
	// 	return err
	// }
	return nil
}

// doLoad is actual load from dss: Stat, and List for dirs
func (ose *oplStoredEntry) doLoad() error {
	se, err := ose.dssStatAndList()
	if err == nil {
		if ose.isTarget {
			ose.setStatsFor("setTls", 1, 0)
		} else {
			ose.setStatsFor("setSls", 1, 0)
		}
		// further processing will mark it with any required STC_DIR_
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
	if st.Stc == opelog.STC_SE_ERROR || st.Stc == opelog.STC_DESC_ERROR {
		if ose.toolRestarted && ose.owo().ClearErrors {
			return ose.doLoad()
		}
	}
	return nil
}

func (ose *oplStoredEntry) checkInventory() error {
	return errors.ErrUnsupported
}
