package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

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
	events   []*opelog.Event
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
	}
	if sTs != -1 {
		ele.source.State = ele.LogicalEntry.GetState(sTs, false)
		ele.source.events = ele.LogicalEntry.GetEvents(sTs, false)
		ele.target.State = ele.LogicalEntry.GetState(sTs, true)
		ele.target.events = ele.LogicalEntry.GetEvents(sTs, true)
	}
	ele.invent.setEmptySt()
	ele.source.setEmptySt()
	ele.target.setEmptySt()
	return ele
}

type RptConfig struct {
	ToolInfo  *opelog.ToolInformation
	IsSorted  bool
	PartsSize int
	IsReverse bool
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
			"SEvTime", "SEvKind", "SEvState", "SEvChildren", "SEvSymLink", "SEvSize", "SEvMtime", "SEvChks",
			"TState", "TChildren", "TSymLink", "TSize", "TMtime", "TChks",
			"TEvTime", "TEvKind", "TEvState", "TEvChildren", "TevSymLink", "TevSize", "TevMtime", "TEvChks",
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
	record[sBase] = ele.source.dispState(false)
	record[sBase+1] = ele.source.dispDir()
	record[sBase+2] = ele.source.dispSymlink()
	record[sBase+3] = ele.source.dispSize()
	record[sBase+4] = ele.source.dispMtime()
	record[sBase+5] = ele.source.dispChecksums()

	tBase := sBase + 6
	record[tBase] = ele.target.dispState(false)
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
	record[sBase] = ele.source.dispState(true)
	record[sBase+1] = ele.source.dispDir()
	record[sBase+2] = ele.source.dispSymlink()
	record[sBase+3] = ele.source.dispSize()
	record[sBase+4] = ele.source.dispMtime()
	record[sBase+5] = ele.source.dispChecksums()

	tBase := sBase + 6 + 8 // 8 == len(getEventsAsStrings())
	record[tBase] = ele.target.dispState(true)
	record[tBase+1] = ele.target.dispDir()
	record[tBase+2] = ele.target.dispSymlink()
	record[tBase+3] = ele.target.dispSize()
	record[tBase+4] = ele.target.dispMtime()
	record[tBase+5] = ele.target.dispChecksums()

	res := [][]string{record}

	for _, sev := range ele.source.getEventsAsStrings() {
		record = make([]string, len(cols))
		for i := range sev {
			record[sBase+6+i] = sev[i]
		}
		res = append(res, record)
	}
	for _, sev := range ele.target.getEventsAsStrings() {
		record = make([]string, len(cols))
		for i := range sev {
			record[tBase+6+i] = sev[i]
		}
		res = append(res, record)
	}
	return res
}

func (ele *expLe) dispSyncStatus() (sys string) {
	csts := ele.GetStats(ele.sTs)
	if csts == nil {
		return
	}
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
	if ele.IsIgnored {
		return "i"
	}
	stats := ele.GetStats(ele.sTs)
	if stats == nil {
		return ""
	}
	return stats.Display("summary")
}

func (est *expSt) dispState(full bool) string {
	if est.Stc == opelog.STC_DONE_PRESENT {
		return est.Se.ToDataEntry("").ModString()
	}
	if est.Stc == opelog.STC_DONE_ABSENT {
		return "-"
	}
	if est.Stc == opelog.STC_SE_ERROR {
		if full {
			return fmt.Sprintf("e: \"%s\"", est.Error)
		}
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
		var err error
		sEvs, sEvCn, sEvSL, sEvSz, sEvMt, sEvCss := "-", "", "", "", "", "?css"
		se := ev.Se
		if se != nil {
			sEvs = se.ToDataEntry("").ModString()
			if se.IsDir {
				sEvCn = fmt.Sprintf("%d", len(se.Children))
			}
			sEvSL = se.SymLinkTarget
			if !se.IsDir && !se.IsSymLink {
				sEvSz = fmt.Sprintf("%d", se.Size)
			}
			sEvMt = ts2s(se.Mtime)
		}
		if sEvCss, err = common.TypedChecksums2Checksums(ev.Tcss); err != nil {
			sEvCss = "e"
		}
		if err != nil {
			sEvCss = "e"
		}
		sev := []string{
			ts2s(ev.TimeStamp),
			ev.Kind.String(),
			sEvs,
			sEvCn,
			sEvSL,
			sEvSz,
			sEvMt,
			sEvCss,
		}
		res = append(res, sev)
	}
	return
}

func toolInfoHeader(ti *opelog.ToolInformation) string {
	return strings.Join([]string{
		fmt.Sprintf("exported: %s", ts2s(time.Now().Unix())),
		fmt.Sprintf("command: %s", ti.CmdLine),
		fmt.Sprintf("version: %s", ti.Version),
		fmt.Sprintf("host: %s", ti.HostName),
		fmt.Sprintf("started: %s", ts2s(ti.StartTime)),
		fmt.Sprintf("source: %s", ti.SourceDssInfo),
		fmt.Sprintf("target: %s", ti.TargetDssInfo),
	}, "\n")
}

func toolInfoFooter(ti *opelog.ToolInformation) string {
	if ti.StopTime == 0 {
		return "stopped: operation still on-going"
	}
	return strings.Join([]string{
		fmt.Sprintf("stopped: %s", ts2s(ti.StopTime)),
		fmt.Sprintf("summary: %s", ti.ExecSummary),
	}, "\n")
}

type csvPartsStruct struct {
	csvPath      string
	partsSize    int
	csvFileCount int
	csvFile      *os.File
	csvWriter    *csv.Writer
	writeHeader  func(*csv.Writer) error
}

func (cp *csvPartsStruct) write(record []string) error {
	return cp.csvWriter.Write(record)
}

func (cp *csvPartsStruct) closePart() error {
	if cp.csvFile == nil {
		return nil
	}
	cp.csvWriter.Flush()
	if cp.csvWriter.Error() != nil {
		cp.csvFile.Close()
		cp.csvFile = nil
		return cp.csvWriter.Error()
	}
	err := cp.csvFile.Close()
	cp.csvFile = nil
	return err
}

func (cp *csvPartsStruct) partPath() string {
	if cp.partsSize <= 0 {
		return cp.csvPath
	}
	dn := path.Dir(cp.csvPath)
	fn := path.Base(cp.csvPath)
	ext := path.Ext(fn)
	pfn := fn[:len(fn)-len(ext)]
	return path.Join(dn, fmt.Sprintf("%s-%03d%s", pfn, cp.csvFileCount, ext))
}

func (cp *csvPartsStruct) createPart() error {
	if cp.csvFile != nil {
		if err := cp.closePart(); err != nil {
			return err
		}
	}
	cf, err := os.Create(cp.partPath())
	if err != nil {
		return err
	}
	cp.csvFile = cf
	cp.csvWriter = csv.NewWriter(cf)
	cp.csvFileCount++
	return cp.writeHeader(cp.csvWriter)
}

func newCsvPartsStruct(csvPath string, partsSize int, writeHeader func(*csv.Writer) error) (*csvPartsStruct, error) {
	csvParts := &csvPartsStruct{
		csvPath:     csvPath,
		partsSize:   partsSize,
		writeHeader: writeHeader,
	}
	if err := csvParts.createPart(); err != nil {
		return nil, err
	}
	return csvParts, nil
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
		ise, ok = oplm.GetOlm().Sessions[invent]
		if !ok {
			return fmt.Errorf("inventory %s not found", invent)
		}
	}
	if session != "" {
		sse, ok = oplm.GetOlm().Sessions[session]
		if !ok {
			return fmt.Errorf("session %s not found", session)
		}
	}
	ce, ok := exporters[rt]
	if !ok {
		return fmt.Errorf("report type %d is unknown", rt)
	}
	cp, err := newCsvPartsStruct(csvPath, rptCfg.PartsSize, func(cw *csv.Writer) error {
		if err := cw.Write(ce.columns); err != nil {
			return err
		}
		if err := cw.Write([]string{toolInfoHeader(rptCfg.ToolInfo)}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	defer cp.closePart()

	entryNum := 0
	err = oplm.Walk(func(relPath string, le *opelog.LogicalEntry) error {
		entryWritten := true
		if rptCfg.PartsSize > 0 {
			entryNum++
			entryWritten = false
			if entryNum%rptCfg.PartsSize == 0 {
				if err := cp.createPart(); err != nil {
					return err
				}
			}
		}
		records := ce.rowExporter(relPath, le, ise.ReferenceTime, sse.ReferenceTime)
		for _, record := range records {
			if !entryWritten {
				entryWritten = true
				record[0] = fmt.Sprintf("%d", entryNum)
			}
			if err := cp.write(record); err != nil {
				return err
			}
		}
		return nil
	}, rptCfg.IsSorted, rptCfg.IsReverse)
	if err != nil {
		return err
	}
	if err := cp.write([]string{toolInfoFooter(rptCfg.ToolInfo)}); err != nil {
		return err
	}
	return cp.closePart()
}
