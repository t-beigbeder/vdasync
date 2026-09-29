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

func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	sErr := ole.source().load()
	tErr := ole.target().load()
	_, _ = sErr, tErr
	return nil
}
