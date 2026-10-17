package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type RptConfig struct {
	ToolInfo *opelog.ToolInformation
}

type RptType int

const (
	RPT_SYNTHETIC RptType = iota
	RPT_FULL
)

type csvExporter struct {
	columns     []string
	rowExporter func(string, *opelog.LogicalEntry, int64, int64) [][]string
}

var exporters map[RptType]csvExporter = map[RptType]csvExporter{}

func init() {
	exporters[RPT_SYNTHETIC] = csvExporter{
		columns: []string{
			"Meta",
			"RelPath", "InvChks", "Sync", "Summary",
			"SState", "SChildren", "SSymLink", "SSize", "SMtime", "SChks",
			"TState", "TChildren", "TSymLink", "TSize", "TMtime", "TChks",
		},
		rowExporter: syntheticExporter,
	}
	exporters[RPT_FULL] = csvExporter{
		columns: []string{
			"Meta",
			"RelPath", "InvChks", "Sync", "Summary",
			"SState", "SChildren", "SSymLink", "SSize", "SMtime", "SChks",
			"SEvTime", "SEvKind", "SEvChks",
			"TState", "TChildren", "TSymLink", "TSize", "TMtime", "TChks",
			"TEvTime", "TEvKind", "TEvChks",
		},
		rowExporter: fullExporter,
	}
}

func syntheticExporter(relPath string, le *opelog.LogicalEntry, ti, ts int64) [][]string {
	ele := newExpLe(relPath, le, ti, ts)

	cols := exporters[RPT_SYNTHETIC].columns
	record := make([]string, len(cols))
	record[1] = relPath
	record[2] = ele.invent.dispChecksums()
	record[3] = ele.dispSyncStatus()
	record[4] = ele.dispSummary()

	sBase := 5
	record[sBase] = ele.source.dispState()
	record[sBase+1] = ele.source.dispDir()
	record[sBase+2] = ele.source.dispSymlink()
	record[sBase+3] = ele.source.dispSize()
	record[sBase+4] = ele.source.dispMtime()
	record[sBase+5] = ele.source.dispChecksums()

	tBase := sBase + 6
	record[tBase] = ele.target.dispState()
	record[tBase+1] = ele.target.dispDir()
	record[tBase+2] = ele.target.dispSymlink()
	record[tBase+3] = ele.target.dispSize()
	record[tBase+4] = ele.target.dispMtime()
	record[tBase+5] = ele.target.dispChecksums()

	return [][]string{record}
}

func fullExporter(relPath string, le *opelog.LogicalEntry, ti, ts int64) [][]string {
	ele := newExpLe(relPath, le, ti, ts)

	cols := exporters[RPT_FULL].columns
	record := make([]string, len(cols))
	record[1] = relPath
	record[2] = ele.invent.dispChecksums()
	record[3] = ele.dispSyncStatus()
	record[4] = ele.dispSummary()

	sBase := 5
	record[sBase] = ele.source.dispState()
	record[sBase+1] = ele.source.dispDir()
	record[sBase+2] = ele.source.dispSymlink()
	record[sBase+3] = ele.source.dispSize()
	record[sBase+4] = ele.source.dispMtime()
	record[sBase+5] = ele.source.dispChecksums()

	tBase := sBase + 6 + 3
	record[tBase] = ele.target.dispState()
	record[tBase+1] = ele.target.dispDir()
	record[tBase+2] = ele.target.dispSymlink()
	record[tBase+3] = ele.target.dispSize()
	record[tBase+4] = ele.target.dispMtime()
	record[tBase+5] = ele.target.dispChecksums()

	res := [][]string{record}

	for _, sev := range ele.source.getEventsAsStrings() {
		record = make([]string, len(cols))
		record[sBase+6] = sev[0]
		record[sBase+7] = sev[1]
		record[sBase+8] = sev[2]
		res = append(res, record)
	}
	for _, sev := range ele.target.getEventsAsStrings() {
		record = make([]string, len(cols))
		record[tBase+6] = sev[0]
		record[tBase+7] = sev[1]
		record[tBase+8] = sev[2]
		res = append(res, record)
	}
	return res
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
	isInvent bool
	isTarget bool
	ele      *expLe
	events []*opelog.Event
}

func (ele *expLe) dispSyncStatus() (sys string) {
	csts := ele.GetStats(ele.sTs)
	if csts.Remove.Number > 0 {
		sys += "r"
	}
	if csts.Create.Number > 0 {
		sys += "c"
	}
	if csts.Update.Number > 0 {
		sys += "u"
	}
	if csts.MetaChange.Number > 0 {
		sys += "m"
	}
	if csts.Error.Number > 0 {
		sys += "e"
	}
	if sys == "" && (csts.NoOp.Number > 0 || csts.Read.Number > 0) {
		sys = "-"
	}
	return sys
}

func (ele *expLe) dispSummary() string {
	return ele.GetStats(ele.sTs).Display("summary")
}

func (est *expSt) dispState() string {
	if est.Stc == opelog.STC_DONE_PRESENT {
		return est.Se.ToDataEntry("").ModString()
	}
	if est.Stc == opelog.STC_DONE_ABSENT {
		return "-"
	}
	if est.Stc == opelog.STC_SE_ERROR {
		return "e"
	}
	if est.Stc == opelog.STC_DESC_ERROR {
		return "^e"
	}
	return "?"
}

func (est *expSt) dispDir() string {
	if est.Se == nil {
		return ""
	}
	if !est.Se.IsDir {
		return "-"
	}
	sdc := ""
	if est.DepCount > 0 {
		sdc = fmt.Sprintf("(%d)", est.DepCount)
	}
	return fmt.Sprintf("%d%s", len(est.Se.Children), sdc)
}

func (est *expSt) dispSymlink() string {
	if est.Se == nil {
		return ""
	}
	if !est.Se.IsSymLink {
		return "-"
	}
	return est.Se.SymLinkTarget
}

func (est *expSt) dispSize() string {
	if est.Se == nil {
		return ""
	}
	if est.Se.IsDir || est.Se.IsSymLink || est.Stc != opelog.STC_DONE_PRESENT {
		return "-"
	}
	return fmt.Sprintf("%d", est.Se.Size)
}

func ts2s(ts int64) string {
	return time.Unix(ts, 0).UTC().Format(time.RFC3339)
}

func (est *expSt) dispMtime() string {
	if est.Se == nil {
		return ""
	}
	if est.Stc != opelog.STC_DONE_PRESENT {
		return "-"
	}
	return ts2s(est.Se.Mtime)
}

func (est *expSt) dispChecksums() string {
	css, err := common.TypedChecksums2Checksums(est.Tcss)
	if err != nil {
		return "e"
	}
	return css
}

func (est *expSt) getEventsAsStrings() (res [][]string) {
	for _, ev := range est.events {
		scs, err := common.TypedChecksums2Checksums(ev.Tcss)
		if err != nil {
			scs = "e"
		}
		sev := []string{
			ts2s(ev.TimeStamp),
			ev.Kind.String(),
			scs,
		}
		res = append(res, sev)
	}
	return
}

func (est *expSt) setEmptySt() {
	if est.State == nil {
		est.State = &opelog.State{}
	}
}

func newExpLe(relPath string, le *opelog.LogicalEntry, iTs, sTs int64) *expLe {
	ele := &expLe{
		relPath:      relPath,
		LogicalEntry: le,
		iTs:          iTs,
		sTs:          sTs,
		invent:       &expSt{isInvent: true},
		source:       &expSt{},
		target:       &expSt{isTarget: true},
	}
	ele.invent.ele = ele
	ele.source.ele = ele
	ele.target.ele = ele
	if iTs != -1 {
		ele.invent.State = ele.LogicalEntry.GetState(iTs, false)
		ele.invent.setEmptySt()
	}
	if sTs != -1 {
		ele.source.State = ele.LogicalEntry.GetState(sTs, false)
		ele.source.events = ele.LogicalEntry.GetEvents(sTs, false)
		ele.source.setEmptySt()
		ele.target.State = ele.LogicalEntry.GetState(sTs, true)
		ele.target.events = ele.LogicalEntry.GetEvents(sTs, true)
		ele.target.setEmptySt()
	}
	return ele
}

func OplCsvExport(lgr *slog.Logger, oplm opelog.OpeLogManager,
	csvPath string, rt RptType, invent, session string,
	rptCfg *RptConfig,
) error {
	if rptCfg == nil {
		rptCfg = &RptConfig{ToolInfo: &opelog.ToolInformation{CmdLine: "no tool information"}}
	}
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
	ce, ok := exporters[rt]
	if !ok {
		return fmt.Errorf("report type %d is unknown", rt)
	}
	wrw, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer wrw.Close()
	cw := csv.NewWriter(wrw)
	if err := cw.Write(ce.columns); err != nil {
		return err
	}
	meta := fmt.Sprintf("%v\nexport timestamp: %s", *rptCfg.ToolInfo, time.Now().Format(time.RFC3339))
	if err := cw.Write([]string{meta}); err != nil {
		return err
	}

	err = oplm.Walk(func(relPath string, le *opelog.LogicalEntry) error {
		records := ce.rowExporter(relPath, le, ise.ReferenceTime, sse.ReferenceTime)
		for _, record := range records {
			if err := cw.Write(record); err != nil {
				return err
			}
		}
		return nil
	})
	cw.Flush()
	return wrw.Close()
}
