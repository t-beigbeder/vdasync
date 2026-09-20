package opelogimpl

import (
	"fmt"
	"io"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ole *oplLogicalEntry) getCsAlgos() (string, string) {
	algos := ""
	if ole.owi.owo.Check {
		algos = ole.owi.owo.CsAlgos
		if algos == "" {
			algos = "sha256"
		}
	}
	readAlgos := common.AddAlgos(algos, common.AlgosFrom(ole.le.InvChecksums))
	return algos, readAlgos
}

func (ole *oplLogicalEntry) validateChecksums(readCss string) (string, error) {
	oAlgos, _ := ole.getCsAlgos()
	iAlgos := common.AlgosFrom(ole.le.InvChecksums)
	sose, _ := ole.source(), ole.target()
	iCss := common.FilterCss(readCss, iAlgos)
	if iCss != ole.le.InvChecksums {
		err := fmt.Errorf(
			"inventory checksum failed: inv %s actual %s", sose.le.InvChecksums, iCss)
		sose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return "", err
	}
	oCss := common.FilterCss(readCss, oAlgos)
	pCss := sose.currentEvent().Checksums
	if pCss != "" && oCss != pCss {
		err := fmt.Errorf(
			"validate checksum failed: previous %s actual %s", pCss, oCss)
		sose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return "", err
	}
	return oCss, nil
}

func (ole *oplLogicalEntry) copy() error {
	sose, tose := ole.source(), ole.target()
	sose.detail("dss.GetReadCloser", "path", sose.fullPath())
	rdr, err := sose.dss().GetReadCloser(sose.fullPath())
	if err != nil {
		sose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	defer rdr.Close()

	_, readAlgos := ole.getCsAlgos()
	cr, err := common.NewChecksumsReader(rdr, readAlgos)
	if err != nil {
		sose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}

	tose.detail("dss.GetWriteCloser", "path", tose.fullPath())
	wrr, err := tose.dss().GetWriteCloser(tose.fullPath())
	if err != nil {
		tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
		return nil
	}
	defer wrr.Close()

	written, err := io.Copy(wrr, cr)
	if err != nil {
		tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
		return nil
	}
	if written != sose.currentState().Size {
		err := fmt.Errorf("copied %d from %d", written, sose.currentState().Size)
		tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
		return nil
	}
	if err := wrr.Close(); err != nil {
		tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
		return nil
	}
	oCss, err := ole.validateChecksums(cr.Checksums())
	if err != nil {
		return nil
	}
	sose.currentEvent().Checksums = oCss
	sose.hasChanges = true
	return nil
}

func (ole *oplLogicalEntry) copyStat() error {
	sose, tose := ole.source(), ole.target()
	sse := sose.currentState()
	tose.detail("dss.SetStat", "path", tose.fullPath())
	if err := tose.dss().SetStat(
		sse.ToDataEntry(tose.fullPath()),
		tose.owi.owo.NoPerm, tose.owi.owo.NoMtime); err != nil {
		tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_SET_STAT, err.Error())
		return nil
	}
	return nil
}
