package opelogimpl

import (
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/t-beigbeder/vdasync/opelog"
)

func clioLe(le *opelog.LogicalEntry, sTs int64) (sys string, hasErr bool) {
	csts := le.GetStats(sTs)
	if csts == nil {
		sys = "?"
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
		hasErr = true
	}
	if sys == "" && (csts.NoOp.Number > 0 || csts.Read.Number > 0) {
		sys = "-"
	}
	return
}

func CliOutput(lgr *slog.Logger, oplm opelog.OpeLogManager, out io.Writer, session string, isSorted, isReverse bool) error {
	sse, ok := oplm.GetOlm().Sessions[session]
	if !ok {
		return fmt.Errorf("session %s not found", session)
	}
	sTs := sse.ReferenceTime
	rootErr := false
	err := oplm.Walk(func(relPath string, le *opelog.LogicalEntry) error {
		outS, hasErr := clioLe(le, sTs)
		if relPath == "" {
			relPath = "."
			if hasErr {
				rootErr = true
			}
		}
		_, err := fmt.Fprintf(out, "%3s %s\n", outS, relPath)
		return err
	}, isSorted, isReverse)
	if err != nil {
		return err
	}
	if rootErr {
		return errors.New("error(s) found, use csv export")
	}
	return nil
}
