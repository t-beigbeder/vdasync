package opelogimpl

import (
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
	*opelog.LogicalEntry
	source *expSe
	target *expSe
}

type expSe struct {
	*opelog.StoredEntry
	isTarget bool
}

func OplCsvExport(lgr *slog.Logger, oplm opelog.OpeLogManager,
	csvPath string, rt RptType, session string,
) error {
	return nil
}
