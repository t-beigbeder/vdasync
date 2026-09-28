package opelogimpl

import "errors"

// This file is about high order services for stored entries
// common services on dss are in opldss

func (ose *oplStoredEntry) load() error {
	st := ose.getState()
	if st == nil {
		ose.dssStat()
	}
	return errors.ErrUnsupported
}
