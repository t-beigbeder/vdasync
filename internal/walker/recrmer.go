package walker

import (
	"fmt"
	"log/slog"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
)

func NewRecursiveRemover(
	lgr *slog.Logger, concurrency int,
	sourceDs dssa.Dssa,
) Walker {
	return MakeWalker(
		lgr,
		concurrency,
		sourceDs,
		onStartDirEntryRRm,
		nil,
		nil,
		nil,
		onDoneEntryRRm,
	)
}

func onStartDirEntryRRm(pe *ProcessedEntry) []*dssa.DataEntry {
	des, err := pe.Dssa_().List(pe.DataEntry.Path)
	if err != nil {
		pe.Error = fmt.Errorf("onStartDirEntryRRm: List(%v): %v", pe.DataEntry.Path, err)
		return nil
	}
	return des
}

func onDoneEntryRRm(pe *ProcessedEntry) {
	err := pe.Dssa_().Rm(pe.DataEntry.Path)
	if err != nil {
		pe.Error = fmt.Errorf("onStartNdirEntryRRm: Rm(%v): %v", pe.DataEntry.Path, err)
		return
	}
}

func RemoveAll(lgr *slog.Logger,  concurrency int,ds dssa.Dssa, path_ dssa.Path) error {
	walker := NewRecursiveRemover(lgr, concurrency, ds)
	de, err := ds.Stat(path_)
	if err != nil {
		return err
	}
	err = walker.Run(de)
	if err != nil {
		return err
	}
	if de.Error != nil {
		return de.Error
	}
	return nil
}