package opelogimpl

import (
	"errors"
	"strings"
)

// This file is about high order services for logical entries

func (ole *oplLogicalEntry) load() error {
	ole.detail("load: start")
	if err := ole.source().load(); err != nil {
		return err
	}
	if err := ole.target().load(); err != nil {
		return err
	}
	if err := ole.source().checkInventory(); err != nil {
		return err
	}

	return nil
}

func (ole *oplLogicalEntry) process() error {
	var err error
	ole.lgr().Debug("process: start")

	for goal := range strings.SplitSeq("verify,update,create,load", ",") {
		if !ole.owi.impliesGoal(goal) {
			continue
		}
		switch goal {
		case "load":
			err = ole.load()
			break
		case "create":
			err = errors.ErrUnsupported
		case "update":
			err = errors.ErrUnsupported
		case "verify":
			err = errors.ErrUnsupported
		default:
			err = errors.ErrUnsupported
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	return nil
}
