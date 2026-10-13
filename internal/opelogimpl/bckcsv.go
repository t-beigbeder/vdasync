package opelogimpl

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type rptLe struct {
	*oplLogicalEntry
	rseSrc *rptSe
	rseTgt *rptSe
}

func newRptLe(owi *oplWalkerImpl, relPath string, le *opelog.LogicalEntry) *rptLe {
	ole := &oplLogicalEntry{
		relPath: relPath,
		owi:     owi,
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
	if se == nil || !se.IsDir {
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

func BckOplCsvExport(lgr *slog.Logger, ow OplWalker, oplm opelog.OpeLogManager,
	csvPath string, rt RptType, session string,
) error {
	var ok bool
	if ow == nil {
		ow = NewOplWalker(
			lgr, 1, nil, oplm, &config.OpeLogOptionsType{},
			nil, nil, "", "", session, "",
		)
	}
	owi, ok := ow.(*oplWalkerImpl)
	if !ok {
		return fmt.Errorf("OplCsvExport need an oplWalkerImpl")
	}
	// FIXME: loose coupling and session ts
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
		if err := cw.Write(ce.rowExporter(owi, relPath, ole)); err != nil {
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
