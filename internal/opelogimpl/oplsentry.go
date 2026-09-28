package opelogimpl

import "errors"

// This file is about high order services for stored entries
// common services on dss are in opldss

func (ose *oplStoredEntry) load() error {
	st := ose.getState()
	if st != nil && st.Se != nil {
		return nil
	}
	se, err := ose.dssStatAndList()
	if err != nil {
		return nil
	}
	_ = se
	return errors.ErrUnsupported
}

func (ose *oplStoredEntry) checkInventory() error {
	return errors.ErrUnsupported
}
