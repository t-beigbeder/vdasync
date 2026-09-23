package opelogimpl

import (
	"errors"
	"fmt"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ose *oplStoredEntry) isPresent() bool {
	return *ose.prc() == opelog.PRC_PRESENT
}

func (ose *oplStoredEntry) isAbsent() bool {
	return *ose.prc() == opelog.PRC_ABSENT
}

func (ose *oplStoredEntry) newEvent(
	kind opelog.EventCode,
	se *opelog.StoredEntry,
	tcss [][]byte,
	sErr string,
) {
	ev := &opelog.Event{
		Kind:      kind,
		TimeStamp: time.Now().Unix(),
		SeNum:     ose.le.AddOrShareSe(se),
		TcsNums:   ose.le.AddOrShareTcss(tcss),
		Error:     sErr,
	}
	evs := ose.events()
	*evs = append(*evs, ev)
	ose.hasChanges = true
}

func (ose *oplStoredEntry) seLoad() error {
	return errors.ErrUnsupported
}

func (ose *oplStoredEntry) checkInventory() error {
	// TODO: check entry existence
	iev, ise := ose.lastInvEvent()
	if iev == nil || ose.owi.owo.NoInvCheck || !ose.isPresent() {
		return nil
	}
	if ise.IsDir {
		return nil
	}
	if len(iev.TcsNums) == 0 {
		return nil
	}
	if ose.owi.impliesGoal("create") && ose.requiresCreate() {
		return nil
	}
	if ose.owi.impliesGoal("update") && ose.requiresUpdate() {
		return nil
	}

	*ose.prc() = opelog.PRC_LOADING
	ose.detail("dss.GetReadCloser", "path", ose.fullPath(), "algos", ose.owi.owo.InvCsAlgos)
	rr, err := ose.dss().GetReadCloser(ose.fullPath())
	if err != nil {
		ose.newEvent(opelog.EVT_ERROR_RAISED, nil, nil, err.Error())
		return nil
	}
	defer rr.Close()
	rCss, err := common.ReaderChecksum(rr, ose.owi.owo.InvCsAlgos)
	if err != nil {
		ose.newEvent(opelog.EVT_ERROR_RAISED, nil, nil, err.Error())
		return nil
	}
	iCss, err := common.TypedChecksums2Checksums(ose.le.GetTcssFor(iev))
	if err != nil {
		ose.newEvent(opelog.EVT_ERROR_RAISED, nil, nil, err.Error())
		return nil
	}
	if rCss != iCss {
		err := fmt.Errorf("inventory checksum failed: inv %s actual %s", iCss, rCss)
		ose.newEvent(opelog.EVT_ERROR_RAISED, nil, nil, err.Error())
		return nil
	}
	return nil
}

func (ose *oplStoredEntry) seCreate() error {
	return errors.ErrUnsupported
}
