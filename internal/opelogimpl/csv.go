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
	columns []string
	rowExporter func(relPath string, ole *opelog.LogicalEntry) []string
}

var exporters map[RptType]csvExporter = map[RptType]csvExporter{
	RPT_SYNTHETIC: csvExporter{
		columns: []string{"a"},
		rowExporter: func(relPath string, ole *opelog.LogicalEntry) []string {return []string{relPath, ole.InvChecksums}},
	},
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
