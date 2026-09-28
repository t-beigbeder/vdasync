package opelogimpl

import (
	"errors"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ose *oplStoredEntry) dssStat() (*dssa.DataEntry, error) {
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil {
		return nil, ose.owErr("dssStat", err)
	}
	return de, nil
}

func (ose *oplStoredEntry) dssList() (*dssa.DataEntry, error) {
	des, err := ose.dss().List(ose.fullPath())
	if err != nil {
		return nil, ose.owErr("dssStat", err)
	}
	return de, nil
}

// noLstatOnList bool
func (ose *oplStoredEntry) dssStatAndList() (*opelog.StoredEntry, error) {
	de, err := ose.dssStat()
	if err != nil {
		return nil, err
	}
	if !de.IsDir {
		return opelog.FromDataEntry(de, nil), nil
	}
	opelog.FromDataEntry(de, nil)
}
