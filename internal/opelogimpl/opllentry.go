package opelogimpl

import ()

// This file is about high order services for logical entries
// common services on dss are in opldss

// func (ole *oplLogicalEntry) load() (leErr *LeError, err error) {
// 	ole.detail("load: start")
// 	var (
// 		sErr, tErr error
// 	)
// 	// TODO: may need to read source for checksums verification
// 	// however this must only be done after ensuring no copy to target is needed
// 	sErr = ole.source().load()
// 	// TODO: source children can be created absent in target if target absent
// 	// children entries should only be written once
// 	tErr = ole.target().load()
// 	if sErr == nil {
// 		sErr = ole.source().checkInventory()
// 	}
// 	if sErr != nil || tErr != nil {
// 		leErr = &LeError{source: sErr, target: tErr}
// 	}
// 	return
// }

// func (ole *oplLogicalEntry) manageRestart() error {
// 	tSt := ole.target().getState()
// 	if tSt == nil {
// 		return nil
// 	}
// 	if tSt.ToolStartTime == ole.owi.toolStartTime {
// 		return nil
// 	}
// 	switch {
// 	case tSt.Stc == opelog.STC_DIR_CHANGE && ole.owi.impliesGoal("update"):
// 		return nil
// 	case tSt.Stc == opelog.STC_DIR_RM && ole.owi.impliesGoal("update"):
// 		return nil
// 	case tSt.Stc == opelog.STC_DIR_LOAD && ole.owi.impliesGoal("load"):
// 		return nil
// 	case tSt.Stc == opelog.STC_DIR_CHANGE && ole.owi.impliesGoal("create"):
// 		return nil
// 	default:
// 		return nil
// 	}
// }

func (ole *oplLogicalEntry) tryChange() (bool, error) {
	updAble, creAble := ole.owi.impliesGoal("update"), ole.owi.impliesGoal("create")
	if ole.source().hasError() || ole.target().hasError() {
		return false, nil
	}
	if ole.source().isPresent() && ole.target().isPresent() {
		if ole.equalType() {
			
		}
	}
}

// process evaluates current states of source and target wrt the context
// and performs required actions.
//
// Only errors concerning the walker are notified, stored entry level errors are silenced.
func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	_ = ole.source().load()
	_ = ole.target().load()
	done, err := ole.tryChange()
	if done || err != nil {
		return nil
	}
	return nil
}
