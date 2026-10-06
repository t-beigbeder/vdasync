package opelogimpl

import (
	"fmt"
	"log/slog"
	"path"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

// This file is about low-level structure for logical and stored entries
// Higher order services are either in opllentry (logical) or in oplsentry (stored)
// Common dss related services are in opldss

// LeError may group one or two errors on source and/or target stored entry
type LeError struct {
	source error
	target error
}

func (e *LeError) Error() string {
	s := ""
	if e.source != nil {
		s += fmt.Sprintf("source %s", e.source.Error())
	}
	if e.target != nil {
		if s != "" {
			s += " - "
		}
		s += fmt.Sprintf("target %s", e.target.Error())
	}
	return s
}

// oplLogicalEntry groups operations on both source and target (oplStoredEntry)
//
// Its services notifies errors that are meaningful from walker point of view,
// for instance regarding parent-child communication.
// oplStoredEntry-related errors on the other hand are only meaningful
// at the logical entry level and may be kept silent from its services.
type oplLogicalEntry struct {
	lgr        *slog.Logger
	hasChanges bool
	relPath    string
	owi        *oplWalkerImpl
	le         *opelog.LogicalEntry
	sOse       *oplStoredEntry
	tOse       *oplStoredEntry
	source     *oplStoredEntry
	target     *oplStoredEntry
	// needed to understand what is requested from parent's stored entries
	parentLe  *opelog.LogicalEntry
	parentSSt *opelog.State
	parentTSt *opelog.State
	// share source/target
	childrenLeCache map[string]*opelog.LogicalEntry
}

func (ole *oplLogicalEntry) detail(msg string, args ...any) {
	ole.lgr.Log(ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ole *oplLogicalEntry) owo() *config.OpeLogOptionsType { return ole.owi.owo }

func (ole *oplLogicalEntry) logErr(msg string, err error) error {
	ole.lgr.Error(msg, "err", err)
	return err
}

func (ole *oplLogicalEntry) seEqualType() bool {
	if ole.target.se() == nil {
		return false
	}
	return ole.target.se().EqualType(ole.source.se())
}

// getCsAlgos retrieves requested checksums algorithms
func (ole *oplLogicalEntry) getCsAlgos() (csAlgos string) {
	csAlgos = ole.owo().CsAlgos
	if csAlgos == "" {
		csAlgos = "sha256"
	}
	return
}

// isIncluded checks relPath is included (empty list means all) or not excluded
func (ole *oplLogicalEntry) isIncluded() bool {
	return common.IsIncluded(ole.relPath, ole.owi.inclRegs, ole.owi.exclRegs)
}

// oplStoredEntry groups operations and state relevant either for source or for target
//
// errors returned by its services are logged but are only relevant to oplLogicalEntry
// that may keep them silent
type oplStoredEntry struct {
	lgr      *slog.Logger
	ole      *oplLogicalEntry
	isTarget bool
	// processing state
	toolRestarted bool
	// events information, to be created once full processing done
	loadTime       int64
	removeTime     int64
	createTime     int64
	updateTime     int64
	metaChangeTime int64
	// current's dir children must be queued
	childrenQueued bool
}

func (ose *oplStoredEntry) detail(msg string, args ...any) {
	ose.lgr.Log(ose.ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ose *oplStoredEntry) logErr(msg string, err error) error {
	ose.lgr.Error(msg, "err", err)
	return err
}

func (ose *oplStoredEntry) owo() *config.OpeLogOptionsType { return ose.ole.owi.owo }

func (ose *oplStoredEntry) root() string {
	if ose.isTarget {
		return ose.ole.owi.tRoot
	} else {
		return ose.ole.owi.sRoot
	}
}

func (ose *oplStoredEntry) hasChild(child string) bool {
	st := ose.getState()
	if st == nil || st.Se == nil {
		return false
	}
	return st.Se.HasChild(child)
}

func (ose *oplStoredEntry) hasError() bool {
	st := ose.getState()
	if st == nil {
		return false
	}
	return st.Stc == opelog.STC_SE_ERROR || st.Stc == opelog.STC_DESC_ERROR
}

// isPresent detects entry not absent, state may be in progress or have error
func (ose *oplStoredEntry) isPresent() bool {
	st := ose.getState()
	if st == nil {
		return false
	}
	return st.Stc != opelog.STC_DONE_ABSENT
}

// isAbsent detects entry absent
func (ose *oplStoredEntry) isAbsent() bool {
	st := ose.getState()
	if st == nil {
		return false
	}
	return st.Stc == opelog.STC_DONE_ABSENT
}

func (ose *oplStoredEntry) se() *opelog.StoredEntry {
	st := ose.getState()
	if st == nil {
		return nil
	}
	return st.Se
}

func (ose *oplStoredEntry) isDir() bool {
	se := ose.se()
	if se == nil {
		return false
	}
	return se.IsDir
}

func (ose *oplStoredEntry) isSymLink() bool {
	se := ose.se()
	if se == nil {
		return false
	}
	return se.IsSymLink
}

func (ose *oplStoredEntry) isRegularFile() bool {
	se := ose.se()
	if se == nil {
		return false
	}
	return !se.IsDir && !se.IsSymLink
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

func (ose *oplStoredEntry) createEvent(ts int64, kind opelog.EventCode, se *opelog.StoredEntry, tcss [][]byte) {
	ose.ole.le.CreateEvent(ose.ole.owi.sessionTime, ts, ose.isTarget, kind, se, tcss)
	ose.ole.hasChanges = true
}

func (ose *oplStoredEntry) getEvents() []*opelog.Event {
	return ose.ole.le.GetEvents(ose.ole.owi.sessionTime, ose.isTarget)
}

func (ose *oplStoredEntry) setState(isInv bool, stc opelog.StateCode, sErr string, se *opelog.StoredEntry, tcss [][]byte, depCount int) {
	sessInvTs := ose.ole.owi.sessionTime
	if isInv {
		sessInvTs = ose.ole.owi.invTime
	}
	ose.ole.le.SetState(ose.ole.owi.toolStartTime, sessInvTs, ose.isTarget, stc, sErr, se, tcss, depCount)
	ose.ole.hasChanges = true
	if sErr != "" {
		ose.setStatsFor("er", 0)
	}
}

// getState may be nil or Se may be nil meaning simply listed by parent
//
// toolRestarted if loaded state differs
func (ose *oplStoredEntry) getState() *opelog.State {
	st := ose.ole.le.GetState(ose.ole.owi.sessionTime, ose.isTarget)
	owi := ose.ole.owi
	if st != nil && st.ToolStartTime != owi.toolStartTime {
		if !ose.toolRestarted {
			ose.toolRestarted = true
			stats := ose.ole.le.GetStats(owi.sessionTime)
			stats.Reset()
			ose.ole.hasChanges = true
		}
	}
	return st
}

func (ose *oplStoredEntry) getStats() *opelog.ComputedStats {
	stats := ose.ole.le.GetStats(ose.ole.owi.sessionTime)
	if stats == nil {
		stats = opelog.NewComputedStats()
	}
	return stats
}

// setStatsFor updates the entry's stats according to given keyword
//
// size -1 asks to read cached state for entry's size
func (ose *oplStoredEntry) setStatsFor(kw string, size int64) {
	stats := ose.getStats()
	if size == -1 {
		se := ose.se()
		if se != nil {
			size = se.Size
		} else {
			size = 0
		}
	}
	switch kw {
	case "sls":
		stats.SourceListOrStat.Number = 1
	case "tls":
		stats.TargetListOrStat.Number = 1
	case "rd":
		stats.Read.Number = 1
		stats.Read.Size = size
	case "cr":
		stats.Create.Number = 1
		stats.Create.Size = size
	case "up":
		stats.Update.Number = 1
		stats.Update.Size = size
	case "rm":
		stats.Remove.Number = 1
		stats.Remove.Size = size
	case "mc":
		stats.MetaChange.Number = 1
	case "no":
		stats.NoOp.Number = 1
		stats.NoOp.Size = size
	case "er":
		stats.Error.Number = 1
	default:
		return
	}
	ose.ole.hasChanges = true
}
