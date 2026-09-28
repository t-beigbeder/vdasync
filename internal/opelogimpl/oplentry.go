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

func (ole *oplLogicalEntry) owo() *config.OpeLogOptionsType { return ole.owi.owo }

func (ole *oplLogicalEntry) oplq() opelog.Queue { return ole.owi.oplq }

func (ole *oplLogicalEntry) oplm() opelog.OpeLogManager { return ole.owi.oplm }

func (ole *oplLogicalEntry) source() *oplStoredEntry {
	return &oplStoredEntry{ole: ole}
}

func (ole *oplLogicalEntry) target() *oplStoredEntry {
	return &oplStoredEntry{ole: ole, isTarget: true}
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
