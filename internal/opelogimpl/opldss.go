package opelogimpl

import (
	"fmt"
	"io"
	"path"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ose *oplStoredEntry) dssCopyFile() (css string, err error) {
	var (
		rdr     io.ReadCloser
		cr      common.ChecksumsReader
		wrr     io.WriteCloser
		written int64
	)
	sOse := ose.ole.source()
	ose.detail("dss copyFile")
	ose.updateTime = time.Now().Unix()
	rdr, err = sOse.dss().GetReadCloser(sOse.fullPath())
	if err != nil {
		_ = ose.logErr("dss copyFile: GetReadCloser", err)
		return
	}
	defer rdr.Close()
	cr, err = common.NewChecksumsReader(rdr, ose.ole.getCsAlgos())
	if err != nil {
		_ = ose.logErr("dss copyFile: NewChecksumsReader", err)
		return
	}
	wrr, err = ose.dss().GetWriteCloser(ose.fullPath())
	if err != nil {
		_ = ose.logErr("dss copyFile: GetWriteCloser", err)
		return
	}
	defer wrr.Close()
	written, err = io.Copy(wrr, cr)
	if err != nil {
		_ = ose.logErr("dss copyFile: Copy", err)
		return
	}
	if written != sOse.getState().Se.Size {
		err = fmt.Errorf("copied %d from %d", written, sOse.getState().Se.Size)
		_ = ose.logErr("dss copyFile: Copy", err)
		return
	}
	if err = wrr.Close(); err != nil {
		_ = ose.logErr("dss copyFile: Close writer", err)
		return
	}
	css = cr.Checksums()
	return
}

func (ose *oplStoredEntry) dssSetStat(se *opelog.StoredEntry, noPerm bool, noMtime bool) error {
	ose.detail("dss setStat")
	ose.metaChangeTime = time.Now().Unix()
	if err := ose.dss().SetStat(se.ToDataEntry(ose.fullPath()), noPerm, noMtime); err != nil {
		return ose.logErr("dss stat", err)
	}
	return nil
}

func (ose *oplStoredEntry) dssStat() (*dssa.DataEntry, error) {
	ose.detail("dss stat")
	ose.loadTime = time.Now().Unix()
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil {
		return nil, ose.logErr("dss stat", err)
	}
	return de, nil
}

func (ose *oplStoredEntry) dssList() ([]*dssa.DataEntry, error) {
	ose.detail("dss list")
	ose.loadTime = time.Now().Unix()
	des, err := ose.dss().List(ose.fullPath())
	if err != nil {
		return nil, ose.logErr("dss list", err)
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

func (ose *oplStoredEntry) dssRm() error {
	ose.detail("dss rm")
	ose.removeTime = time.Now().Unix()
	if err := ose.dss().Rm(ose.fullPath()); err != nil {
		return ose.logErr("dss rm", err)
	}
	return nil
}

func (ose *oplStoredEntry) dssSymLink(target string) error {
	ose.detail("dss symlink")
	ose.metaChangeTime = time.Now().Unix()
	if err := ose.dss().Symlink(ose.fullPath(), target); err != nil {
		return ose.logErr("dss symlink", err)
	}
	return nil
}
