package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/opelog"
)

type RptType int

const (
	RPT_SYNTHETIC RptType = iota
	RPT_FULL
)

type csvExporter struct {
	columns     []string
	rowExporter func(relPath string, ole *opelog.LogicalEntry) []string
}

var exporters map[RptType]csvExporter = map[RptType]csvExporter{}

type rptLe struct {
	*oplLogicalEntry
}

func (rle *rptLe) source() *rptSe {
	return &rptSe{oplStoredEntry: &oplStoredEntry{oplLogicalEntry: rle.oplLogicalEntry}}
}

func (rle *rptLe) target() *rptSe {
	return &rptSe{oplStoredEntry: &oplStoredEntry{oplLogicalEntry: rle.oplLogicalEntry, isTarget: true}}
}

type rptSe struct {
	*oplStoredEntry
}

func (rse *rptSe) curState() *opelog.StoredEntry {
	ole := rse.currentState()
	if ole == nil {
		ole = &opelog.StoredEntry{}
	}
	return ole
}

func (rse *rptSe) dispPres() string {
	eev := rse.currentEvent()
	if eev == nil {
		return ""
	}
	if eev.Error != "" {
		return "e"
	}
	if eev.Kind == opelog.EVT_ABS {
		return "-"
	}
	return "x"
}

func (rse *rptSe) dispCss() string {
	css, sErr := rse.currentChecksums()
	if sErr != "" {
		return "error"
	}
	return css
}

type dispBool bool

func (db dispBool) String() string {
	if db {
		return "x"
	}
	return ""
}

func dispDirChildren(se *opelog.StoredEntry) string {
	if !se.IsPresent {
		return ""
	}
	if !se.IsDir {
		return "-"
	}
	return fmt.Sprintf("%d", len(se.Children))
}

func dispSize(se *opelog.StoredEntry) string {
	if !se.IsPresent {
		return ""
	}
	if se.IsDir || se.IsSymLink {
		return "-"
	}
	return fmt.Sprintf("%d", se.Size)
}

func dispMtime(se *opelog.StoredEntry) string {
	if !se.IsPresent {
		return ""
	}
	return time.Unix(se.Mtime, 0).Format(time.RFC3339)
}

func syntheticExporter(relPath string, le *opelog.LogicalEntry) []string {
	rle := rptLe{oplLogicalEntry: &oplLogicalEntry{le: le}}
	scs := rle.source().curState()
	tcs := rle.target().curState()
	record := make([]string, len(exporters[RPT_SYNTHETIC].columns))
	record[0] = relPath
	record[1] = le.InvChecksums
	sBase := 2
	record[sBase] = rle.source().dispPres()
	record[sBase+1] = dispDirChildren(scs)
	record[sBase+2] = scs.SymLinkTarget
	record[sBase+3] = dispSize(scs)
	record[sBase+4] = dispMtime(scs)
	record[sBase+5] = rle.source().dispCss()
	tBase := 8
	record[tBase] = rle.target().dispPres()
	record[tBase+1] = dispDirChildren(tcs)
	record[tBase+2] = tcs.SymLinkTarget
	record[tBase+3] = dispSize(tcs)
	record[tBase+4] = dispMtime(tcs)
	record[tBase+5] = rle.target().dispCss()

	return record
}

func init() {
	exporters[RPT_SYNTHETIC] = csvExporter{
		columns: []string{
			"RelPath", "InvChecksums",
			"Source", "SChildren", "SSymLink", "SSize", "SMtime", "SCs",
			"Target", "TChildren", "TSymLink", "TSize", "TMtime", "TCs",
		},
		rowExporter: syntheticExporter,
	}
}

func OplCsvExport(oplm opelog.OpeLogManager, csvPath string, rt RptType) error {
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
	err = oplm.Walk(func(relPath string, ole *opelog.LogicalEntry) error {
		if err := cw.Write(ce.rowExporter(relPath, ole)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	cw.Flush()
	return wrw.Close()
}
