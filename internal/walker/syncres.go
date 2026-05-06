package walker

import (
	"fmt"
	"io"
)

func computeSyncHeader(es *SyncEntryStatus) string {
	s := ""
	if es.IsDir {
		s += "d"
	} else {
		s += "f"
	}
	if es.Created {
		s += "c"
	} else {
		s += "-"
	}
	if es.Updated {
		s += "u"
	} else {
		s += "-"
	}
	if es.Removed {
		s += "r"
	} else {
		s += "-"
	}
	if es.ModChanged {
		s += "m"
	} else {
		s += "-"
	}
	if es.Error != nil {
		s += "e"
	} else {
		s += "-"
	}
	return s
}

func DisplaySyncResult(sr map[string]*SyncEntryStatus, wr io.Writer, agg bool) {
	for _, es := range sr {
		sAgg := ""
		if agg {
			sAgg = fmt.Sprintf(" %6d %9d c%6d u%6d", es.AggregatedChildrenNumber, es.AggregatedSize, es.AggregatedCreated, es.AggregatedUpdated)
		}
		wr.Write([]byte(fmt.Sprintf("%s %64s %s\n", computeSyncHeader(es), es.relPath, sAgg)))
	}
}
