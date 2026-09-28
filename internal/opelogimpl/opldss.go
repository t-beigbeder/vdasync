package opelogimpl

import (
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ose *oplStoredEntry) dssStat() (*dssa.DataEntry, error) {
	ose.detail("dss stat")
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil {
		return nil, ose.owErr("dss stat", err)
	}
	return de, nil
}

func (ose *oplStoredEntry) dssList() ([]*dssa.DataEntry, error) {
	ose.detail("dss list")
	des, err := ose.dss().List(ose.fullPath())
	if err != nil {
		return nil, ose.owErr("dss list", err)
	}
	return des, nil
}

func (ose *oplStoredEntry) dssStatAndList() (*opelog.StoredEntry, error) {
	de, err := ose.dssStat()
	if err != nil {
		return nil, err
	}
	if !de.IsDir {
		return opelog.FromDataEntry(de, nil), nil
	}
	des, err := ose.dssList()
	chs := make([]string, len(des))
	for i := range des {
		chs[i] = path.Base(des[i].Path)
	}
	return opelog.FromDataEntry(de, chs), nil
}
