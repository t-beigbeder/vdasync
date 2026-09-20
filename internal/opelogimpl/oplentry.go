package opelogimpl

import (
	"log/slog"
	"path"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about low-level structure for logical and stored entries
// Higer order services are either in opllentry (logical) or in oplsentry (stored)

type oplLogicalEntry struct {
	hasChanges bool
	relPath    string
	plgr       *slog.Logger
	owi        *oplWalkerImpl
	le         *opelog.LogicalEntry
	sHasParent bool
	sChildrenQ []string
	tHasParent bool
	tChildrenQ []string
}

func (ole *oplLogicalEntry) lgr() *slog.Logger {
	return ole.plgr.With("relPath", ole.relPath)
}

func (ole *oplLogicalEntry) detail(msg string, args ...any) {
	ole.lgr().Log(ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ole *oplLogicalEntry) source() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole}
}

func (ole *oplLogicalEntry) target() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole, isTarget: true}
}


type oplStoredEntry struct {
	*oplLogicalEntry
	isTarget bool
}

func (ose *oplStoredEntry) pfx() string {
	if ose.isTarget {
		return "{T}"
	} else {
		return "{S}"
	}
}

func (ose *oplStoredEntry) lgr() *slog.Logger {
	return ose.plgr.With("path", path.Join(ose.pfx(), ose.relPath))
}

func (ose *oplStoredEntry) detail(msg string, args ...any) {
	ose.lgr().Log(ose.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ose *oplStoredEntry) root() string {
	if ose.isTarget {
		return ose.owi.tRoot
	} else {
		return ose.owi.sRoot
	}
}

func (ose *oplStoredEntry) fullPath() string {
	return path.Join(ose.root(), ose.relPath)
}

func (ose *oplStoredEntry) events() (evs *[]*opelog.Event) {
	if ose.isTarget {
		if ose.le.TargetEvents == nil {
			ose.le.TargetEvents = []*opelog.Event{}
		}
		evs = &ose.le.TargetEvents
	} else {
		if ose.le.SourceEvents == nil {
			ose.le.SourceEvents = []*opelog.Event{}
		}
		evs = &ose.le.SourceEvents
	}
	return
}

func (ose *oplStoredEntry) currentEvent() *opelog.Event {
	evs := ose.events()
	if len(*evs) == 0 {
		return nil
	}
	return (*evs)[len(*evs)-1]
}

func (ose *oplStoredEntry) states() (sts *[]*opelog.StoredEntry) {
	if ose.isTarget {
		sts = &ose.le.TargetStates
	} else {
		sts = &ose.le.SourceStates
	}
	return
}

func (ose *oplStoredEntry) currentState() *opelog.StoredEntry {
	sts := ose.states()
	if len(*sts) == 0 {
		return nil
	}
	return (*sts)[len(*sts)-1]
}

func (ose *oplStoredEntry) dss() (ds dssa.Dssa) {
	if ose.isTarget {
		ds = ose.owi.tds
	} else {
		ds = ose.owi.sds
	}
	return
}
