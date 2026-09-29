package opelogimpl

import (
	"errors"
	"strings"
)

// This file is about high order services for logical entries

func (ole *oplLogicalEntry) load() (leErr *LeError, err error) {
	ole.detail("load: start")
	var (
		sErr, tErr error
	)
	sErr = ole.source().load()
	tErr = ole.target().load()
	if sErr == nil {
		sErr = ole.source().checkInventory()
	}
	if sErr != nil || tErr != nil {
		leErr = &LeError{source: sErr, target: tErr}
	}
	return
}

func (ole *oplLogicalEntry) process() error {
	var (
		leErr, err error
	)
	ole.lgr().Debug("process: start")
	// TODO: manage restart on source and target state
	for goal := range strings.SplitSeq("load,create,update,verify", ",") {
		if !ole.owi.impliesGoal(goal) {
			continue
		}
		switch goal {
		case "load":
			leErr, err = ole.load()
		case "create":
			err = errors.ErrUnsupported
		case "update":
			err = errors.ErrUnsupported
		case "verify":
			err = errors.ErrUnsupported
		default:
			err = errors.ErrUnsupported
		}
		if leErr != nil || err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	return nil
}
