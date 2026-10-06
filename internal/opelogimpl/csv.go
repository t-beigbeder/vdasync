package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
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
	rseSrc *rptSe
	rseTgt *rptSe
}

func newRptLe(relPath string, le *opelog.LogicalEntry) *rptLe {
	ole := &oplLogicalEntry{
		relPath: relPath,
		le:      le,
	}
	ole.source = &oplStoredEntry{ole: ole}
	ole.target = &oplStoredEntry{ole: ole, isTarget: true}

	rptLe := &rptLe{oplLogicalEntry: ole}
	rptLe.rseSrc = &rptSe{oplStoredEntry: ole.source}
	rptLe.rseTgt = &rptSe{oplStoredEntry: ole.target}
	return rptLe
}

type rptSe struct {
	*oplStoredEntry
}

func (rse *rptSe) dispPres() string {
	st := rse.getState()
	if st == nil {
		return ""
	}
	if st.Error != "" {
		return "e"
	}
	if st.Stc == opelog.STC_DONE_ABSENT {
		return "-"
	}
	if rse.ole.le.IsIgnored {
		return "i"
	}
	return "x"
}

func (rse *rptSe) dispCss() string {
	st := rse.getState()
	if st == nil {
		return "error"
	}
	tcss := st.Tcss
	css, err := common.TypedChecksums2Checksums(tcss)
	if err != nil {
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
	if !se.IsDir {
		return "-"
	}
	return fmt.Sprintf("%d", len(se.Children))
}

func dispSize(se *opelog.StoredEntry) string {
	if se == nil {
		return ""
	}
	if se.IsDir || se.IsSymLink {
		return "-"
	}
	return fmt.Sprintf("%d", se.Size)
}

func dispMtime(se *opelog.StoredEntry) string {
	if se == nil {
		return ""
	}
	return time.Unix(se.Mtime, 0).Format(time.RFC3339)
}

func syntheticExporter(relPath string, le *opelog.LogicalEntry) []string {
	rle := rptLe{oplLogicalEntry: &oplLogicalEntry{le: le}}
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
