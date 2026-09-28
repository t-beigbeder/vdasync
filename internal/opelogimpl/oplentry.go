package opelogimpl

import (
	"log/slog"
	"path"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about low-level structure for logical and stored entries
// Higher order services are either in opllentry (logical) or in oplsentry (stored)
// Common dss related services are in opldss

type oplLogicalEntry struct {
	hasChanges bool
	relPath    string
	plgr       *slog.Logger
	owi        *oplWalkerImpl
	le         *opelog.LogicalEntry
}

func (ole *oplLogicalEntry) lgr() *slog.Logger { return ole.plgr.With("relPath", ole.relPath) }

func (ole *oplLogicalEntry) detail(msg string, args ...any) {
	ole.lgr().Log(ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ole *oplLogicalEntry) owo() *config.OpeLogOptionsType { return ole.owi.owo }

func (ole *oplLogicalEntry) oplq() opelog.Queue { return ole.owi.oplq }

func (ole *oplLogicalEntry) oplm() opelog.OpeLogManager { return ole.owi.oplm }

func (ole *oplLogicalEntry) source() *oplStoredEntry {
	return &oplStoredEntry{ole: ole}
}

func (ole *oplLogicalEntry) target() *oplStoredEntry {
	return &oplStoredEntry{ole: ole, isTarget: true}
}

func (ole *oplLogicalEntry) owErr(msg string, err error) error {
	return ole.owi.owErr(ole.lgr(), msg, err)
}

type oplStoredEntry struct {
	ole      *oplLogicalEntry
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
	return ose.ole.plgr.With("path", path.Join(ose.pfx(), ose.ole.relPath))
}

func (ose *oplStoredEntry) detail(msg string, args ...any) {
	ose.lgr().Log(ose.ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ose *oplStoredEntry) owErr(msg string, err error) error {
	return ose.ole.owi.owErr(ose.lgr(), msg, err)
}

func (ose *oplStoredEntry) owo() *config.OpeLogOptionsType { return ose.ole.owi.owo }

func (ose *oplStoredEntry) oplq() opelog.Queue { return ose.ole.owi.oplq }

func (ose *oplStoredEntry) oplm() opelog.OpeLogManager { return ose.ole.owi.oplm }

func (ose *oplStoredEntry) root() string {
	if ose.isTarget {
		return ose.ole.owi.tRoot
	} else {
		return ose.ole.owi.sRoot
	}
}

func (ose *oplStoredEntry) fullPath() string {
	return path.Join(ose.root(), ose.ole.relPath)
}

func (ose *oplStoredEntry) dss() (ds dssa.Dssa) {
	if ose.isTarget {
		ds = ose.ole.owi.tds
	} else {
		ds = ose.ole.owi.sds
	}
	return
}

// may be nil or Se may be nil meaning simply listed by parent
func (ose *oplStoredEntry) getState() *opelog.State {
	return ose.ole.le.GetState(ose.ole.owi.sessionTime, ose.isTarget)
}
