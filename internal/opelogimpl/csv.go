package opelogimpl

import (
	"fmt"
	"log/slog"

	"github.com/t-beigbeder/vdasync/opelog"
)

type RptType int

const (
	RPT_SYNTHETIC RptType = iota
	RPT_FULL
)

type csvExporter struct {
	columns     []string
	rowExporter func(owi *oplWalkerImpl, relPath string, ole *opelog.LogicalEntry) []string
}

var exporters map[RptType]csvExporter = map[RptType]csvExporter{}

func init() {
	exporters[RPT_SYNTHETIC] = csvExporter{
		columns: []string{
			"RelPath", "InvChecksums",
			"rseSrc", "SChildren", "SSymLink", "SSize", "SMtime", "SCs",
			"rseTgt", "TChildren", "TSymLink", "TSize", "TMtime", "TCs",
		},
		rowExporter: syntheticExporter,
	}
}

func syntheticExporter(owi *oplWalkerImpl, relPath string, le *opelog.LogicalEntry) []string {
	rle := newRptLe(owi, relPath, le)
	scs := rle.rseSrc.se()
	tcs := rle.rseTgt.se()
	record := make([]string, len(exporters[RPT_SYNTHETIC].columns))
	record[0] = relPath
	record[1] = "" // FIXME: le.InvChecksums
	sBase := 2
	record[sBase] = rle.rseSrc.dispPres()
	record[sBase+1] = dispDirChildren(scs)
	record[sBase+2] = "" // FIXME: scs.SymLinkrseTgt
	record[sBase+3] = dispSize(scs)
	record[sBase+4] = dispMtime(scs)
	record[sBase+5] = rle.rseSrc.dispCss()
	tBase := 8
	record[tBase] = rle.rseTgt.dispPres()
	record[tBase+1] = dispDirChildren(tcs)
	record[tBase+2] = "" // FIXME: tcs.SymLinkrseTgt
	record[tBase+3] = dispSize(tcs)
	record[tBase+4] = dispMtime(tcs)
	record[tBase+5] = rle.rseTgt.dispCss()

	return record
}

type expLe struct {
	relPath string
	*opelog.LogicalEntry
	iTs    int64
	sTs    int64
	invent *expSt
	source *expSt
	target *expSt
}

type expSt struct {
	*opelog.State
	isTarget bool
	ele      *expLe
}

func newExpLe(relPath string, le *opelog.LogicalEntry, iTs, sTs int64) *expLe {
	ele := &expLe{
		relPath:      relPath,
		LogicalEntry: le,
		iTs:          iTs,
		sTs:          sTs,
		invent: &expSt{},
		source:       &expSt{},
		target:       &expSt{isTarget: true},
	}
	ele.invent.ele = ele
	ele.invent.State = ele.LogicalEntry.GetState(iTs, false)
	ele.source.ele = ele
	ele.source.State = ele.LogicalEntry.GetState(sTs, false)
	ele.target.ele = ele
	ele.target.State = ele.LogicalEntry.GetState(sTs, true)
	return ele
}

func OplCsvExport(lgr *slog.Logger, oplm opelog.OpeLogManager,
	csvPath string, rt RptType, invent, session string,
) error {
	ise := &opelog.SessionInfo{ReferenceTime: -1}
	sse := &opelog.SessionInfo{ReferenceTime: -1}
	var ok bool
	if invent != "" {
		ise, ok = oplm.Olm().Sessions[invent]
		if !ok {
			return fmt.Errorf("inventory %s not found", invent)
		}
	}
	if session != "" {
		sse, ok = oplm.Olm().Sessions[session]
		if !ok {
			return fmt.Errorf("session %s not found", session)
		}
	}

	err := oplm.Walk(func(relPath string, le *opelog.LogicalEntry) error {
		ele := newExpLe(relPath, le, ise.ReferenceTime, sse.ReferenceTime)
		_ = ele
		return nil
	})
	return err
}
