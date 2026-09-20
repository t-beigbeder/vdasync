package opelogimpl

import (
	"fmt"
	"path"
	"slices"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about high order services for stored entries

func (ose *oplStoredEntry) existOrAbsEv() (ev *opelog.Event) {
	evs := ose.events()
	for i := range slices.Backward(*evs) {
		if (*evs)[i].Kind == opelog.EVT_ABS || (*evs)[i].Kind == opelog.EVT_EXIST {
			return (*evs)[i]
		}
	}
	return nil
}

func (ose *oplStoredEntry) currentChecksums() (string, string) {
	ev := ose.currentEvent()
	if ev == nil {
		return "", ""
	}
	return ev.Checksums, ev.Error
}

func (ose *oplStoredEntry) newState(se *opelog.StoredEntry) {
	sts := ose.states()
	if len(*sts) == 0 {
		*sts = append(*sts, se)
		return
	}
	if se.Equal((*sts)[len(*sts)-1]) {
		return
	}
	*sts = append(*sts, se)
}

func (ose *oplStoredEntry) newEvent(kind opelog.EventCode, origin opelog.OriginCode, sErr string) {
	evs := ose.events()
	*evs = append(*evs,
		&opelog.Event{
			Kind: kind, Origin: origin, TimeStamp: time.Now().Unix(),
			StateIndex: int32(len(*ose.states()) - 1), Error: sErr})
	ose.hasChanges = true
	if sErr != "" {
		ose.lgr().Error("oplStoredEntry.newEvent", "kind", kind, "origin", origin, "err", sErr)
	}
}

func (ose *oplStoredEntry) isKnown() bool {
	eev := ose.existOrAbsEv()
	if eev == nil {
		return false
	}
	// may be on error anyway
	return true
}

func (ose *oplStoredEntry) isPresent() bool {
	eev := ose.existOrAbsEv()
	if eev == nil || eev.Error != "" || eev.Kind != opelog.EVT_EXIST {
		return false
	}
	return true
}

func (ose *oplStoredEntry) isAbsent() bool {
	eev := ose.existOrAbsEv()
	if eev == nil || eev.Error != "" || eev.Kind != opelog.EVT_ABS {
		return false
	}
	return true
}

func (ose *oplStoredEntry) hasParent() bool {
	if ose.isTarget {
		return ose.tHasParent
	} else {
		return ose.sHasParent
	}
}

func (ose *oplStoredEntry) setChildrenQ(children []string) {
	if ose.isTarget {
		ose.tChildrenQ = slices.Clone(children)
	} else {
		ose.sChildrenQ = slices.Clone(children)
	}
}

func (ose *oplStoredEntry) load() error {
	eev := ose.existOrAbsEv()
	if eev != nil && eev.Error == "" {
		return nil
	}
	ose.lgr().Debug("load: start", "hasParent", ose.hasParent())
	if !ose.hasParent() {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_UNSPECIFIED, "")
		return nil
	}
	ose.detail("dss.Stat", "path", ose.fullPath())
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil && !de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, err.Error())
		return nil
	}
	if de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, "")
		return nil
	}
	var children, fCn []string
	var cdes []*dssa.DataEntry
	if de.IsDir {
		ose.detail("dss.List", "path", ose.fullPath())
		cdes, err = ose.dss().List(ose.fullPath())
		if err != nil {
			se := opelog.FromDataEntry(de, nil)
			se.IsPresent = true
			ose.newState(se)
			ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, err.Error())
			return nil
		}

		for _, cde := range cdes {
			if cde.IsDir {
				children = append(children, path.Base(cde.Path))
			} else {
				fCn = append(fCn, path.Base(cde.Path))
			}
		}
		children = slices.Concat(children, fCn)
		ose.setChildrenQ(children)
	}

	se := opelog.FromDataEntry(de, children)
	se.IsPresent = true
	ose.newState(se)
	ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, "")

	return nil
}

func (ose *oplStoredEntry) checkInventory() error {
	if ose.le.InvChecksums == "" || ose.owi.owo.NoInvCheck || !ose.isPresent() {
		return nil
	}
	if ose.currentState().IsDir {
		return nil
	}

	eev := ose.existOrAbsEv()
	if eev.Checksums != "" {
		return nil
	}
	if ose.owi.impliesGoal("create") && ose.requiresCreate() {
		return nil
	}
	if ose.owi.impliesGoal("update") && ose.requiresUpdate() {
		return nil
	}

	ose.detail("dss.GetReadCloser", "path", ose.fullPath(), "algos", ose.owi.owo.InvCsAlgos)
	rr, err := ose.dss().GetReadCloser(ose.fullPath())
	if err != nil {
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	defer rr.Close()
	if eev.Checksums, err = common.ReaderChecksum(rr, ose.owi.owo.InvCsAlgos); err != nil {
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	if eev.Checksums != ose.le.InvChecksums {
		err := fmt.Errorf(
			"inventory checksum failed: inv %s actual %s",
			ose.le.InvChecksums, eev.Checksums)
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	return nil
}

func (ose *oplStoredEntry) create() error {
	sose := ose.source()
	sse := sose.currentState()
	tde := sse.ToDataEntry(ose.fullPath())
	// algorithm implies target has parent existing
	if sse.IsDir {
		ose.detail("dss.Mkdir", "path", ose.fullPath())
		if err := ose.dss().Mkdir(tde); err != nil {
			ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_MKDIR, err.Error())
			return nil
		}
		ose.setChildrenQ(sse.Children)
		ose.le.DirupChildren = nil // FIXME: needed?
		se := sse.CreatedFrom()
		ose.newState(se)
		ose.newEvent(opelog.EVT_EXIST, opelog.ORI_MKDIR, "")
		if len(sse.Children) != 0 {
			ose.le.DirUpdating = true
			ose.newEvent(opelog.EVT_START_DIRUP, opelog.ORI_UNSPECIFIED, "")
			return nil
		}
	}
	if sse.IsSymLink {
		ose.detail("dss.Symlink", "SymLinkTarget", sse.SymLinkTarget, "path", ose.fullPath())
		if err := ose.dss().Symlink(sse.SymLinkTarget, ose.fullPath()); err != nil {
			ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
			return nil
		}
	}
	if !sse.IsDir && !sse.IsSymLink {
		if err := ose.copy(); err != nil {
			return err
		}
	}
	if err := ose.copyStat(); err != nil {
		return nil
	}
	return nil
}
