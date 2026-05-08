package walker

import (
	"fmt"
	"io"
	"slices"
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

func DisplaySyncResult(sr map[string]*SyncEntryStatus, wr io.Writer, agg, all bool) {
	keys := make([]string, len(sr))

	i := 0
	for k := range sr {
		keys[i] = k
		i++
	}
	slices.Sort(keys)

	for _, key := range keys {
		es := sr[key]
		hdr := computeSyncHeader(es)
		if all || hdr[1:] != "-----" {
			wr.Write([]byte(fmt.Sprintf("%s %s\n", hdr, es.relPath)))
		}
	}
	res := sr[""]
	if agg {
		wr.Write([]byte(fmt.Sprintf(
			"total: %d errors: %d size: %d created: %d updated: %d removed: %d mod changed: %d\n",
			res.AggregatedChildrenNumber, res.AggregatedError, res.AggregatedSize,
			res.AggregatedCreated, res.AggregatedUpdated,
			res.AggregatedRemoved, res.AggregatedModChanged,
		)))
	}
}
