package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"os"

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
	oplLogicalEntry
}

func (rle *rptLe)curSState() *opelog.StoredEntry {
	ole := rle.source().currentState()
	if ole == nil {
		ole = &opelog.StoredEntry{}
	}
	return ole
}

type dispBool bool

func (db dispBool) String() string {
	if db {
		return "x"
	}
	return ""
}

func syntheticExporter(relPath string, le *opelog.LogicalEntry) []string {
	rle := rptLe{oplLogicalEntry: oplLogicalEntry{le: le}}
	record := make([]string, len(exporters[RPT_SYNTHETIC].columns))
	record[0] = relPath
	record[1] = le.InvChecksums
	record[2] = dispBool(rle.curSState().IsPresent).String()
	return record
}

func init() {
	exporters[RPT_SYNTHETIC] = csvExporter{
		columns:     []string{"RelPath", "InvChecksums", "Source"},
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
