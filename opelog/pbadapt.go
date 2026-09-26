package opelog

import (
	"bytes"
	"maps"
	"slices"

	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

func gr2ser(gr *opeloggrpc.Rights) *Rights {
	if gr == nil {
		return nil
	}
	return &Rights{Read: gr.Read, Write: gr.Write, Execute: gr.Execute}
}

func protoBuf2StoredEntry(gse *opeloggrpc.StoredEntry) *StoredEntry {
	if gse == nil {
		return nil
	}
	return &StoredEntry{
		IsDir:         gse.IsDir,
		Size:          gse.Size,
		Mtime:         gse.Mtime,
		User:          gse.User,
		UserRights:    gr2ser(gse.UserRights),
		Group:         gse.Group,
		GroupRights:   gr2ser(gse.GroupRights),
		OtherRights:   gr2ser(gse.OtherRights),
		IsSymLink:     gse.IsSymLink,
		SymLinkTarget: gse.SymLinkTarget,
		Children:      slices.Clone(gse.Children),
		AddMeta:       bytes.Clone(gse.AddMeta),
	}
}

func protoBuf2StoredEntries(gses []*opeloggrpc.StoredEntry) []*StoredEntry {
	ses := make([]*StoredEntry, len(gses))
	for i, gse := range gses {
		ses[i] = protoBuf2StoredEntry(gse)
	}
	return ses
}

func protobuf2Event(gev *opeloggrpc.Event) *Event {
	if gev == nil {
		return nil
	}
	return &Event{
		IsTarget:  gev.IsTarget,
		Kind:      EventCode(gev.Kind),
		TimeStamp: gev.TimeStamp,
		seNum:     gev.SeNum,
		tcsNums:   slices.Clone(gev.TcsNums),
	}
}

func protoBuf2Events(gevs *opeloggrpc.Events) []*Event {
	evs := make([]*Event, len(gevs.EventsList))
	for i, gev := range gevs.EventsList {
		evs[i] = protobuf2Event(gev)
	}
	return evs
}

func protoBuf2EventsLists(gevsLs map[int64]*opeloggrpc.Events) map[int64][]*Event {
	evsls := make(map[int64][]*Event, len(gevsLs))
	for k, v := range maps.All(gevsLs) {
		evsls[k] = protoBuf2Events(v)
	}
	return evsls
}

func protoBuf2Tcss(gtcss []*opeloggrpc.TypedChecksum) []*TypedChecksum {
	tcss := make([]*TypedChecksum, len(gtcss))
	for i, tcs := range gtcss {
		tcss[i] = &TypedChecksum{Tcs: bytes.Clone(tcs.Tcs)}
	}
	return tcss
}

func protoBuf2AggInfo(gai *opeloggrpc.AggInfo) *AggInfo {
	if gai == nil {
		return nil
	}
	return &AggInfo{Number: gai.Number, Size: gai.Size}
}

func protoBuf2ComputedStats(gcs *opeloggrpc.ComputedStats) *ComputedStats {
	if gcs == nil {
		return nil
	}
	return &ComputedStats{
		SourceListOrStat: protoBuf2AggInfo(gcs.SourceListOrStat),
		TargetListOrStat: protoBuf2AggInfo(gcs.TargetListOrStat),
		Read:             protoBuf2AggInfo(gcs.Read),
		Create:           protoBuf2AggInfo(gcs.Create),
		Update:           protoBuf2AggInfo(gcs.Update),
		Remove:           protoBuf2AggInfo(gcs.Remove),
		MetaChange:       protoBuf2AggInfo(gcs.MetaChange),
		Error:            protoBuf2AggInfo(gcs.Error),
	}
}

func protoBuf2ComputedStatsMap(gcss map[int64]*opeloggrpc.ComputedStats) map[int64]*ComputedStats {
	css := make(map[int64]*ComputedStats, len(gcss))
	for k, v := range maps.All(gcss) {
		css[k] = protoBuf2ComputedStats(v)
	}
	return css
}

func protoBuf2State(gst *opeloggrpc.State) *State {
	if gst == nil {
		return nil
	}
	return &State{
		Stc:      StateCode(gst.Stc),
		seNum:    gst.SeNum,
		tcsNums:  slices.Clone(gst.TcsNums),
		DepCount: gst.DepCount,
	}
}

func protoBuf2States(gsts map[int64]*opeloggrpc.State) map[int64]*State {
	sts := make(map[int64]*State, len(gsts))
	for k, v := range maps.All(gsts) {
		sts[k] = protoBuf2State(v)
	}
	return sts
}

func ProtoBuf2LogicalEntry(gle *opeloggrpc.LogicalEntry) *LogicalEntry {
	if gle == nil {
		return nil
	}
	le := &LogicalEntry{
		sharedSes:    protoBuf2StoredEntries(gle.SharedSes),
		sharedTcss:   protoBuf2Tcss(gle.SharedTcss),
		sourceStates: protoBuf2States(gle.SourceStates),
		targetStates: protoBuf2States(gle.TargetStates),
		eventsLists:  protoBuf2EventsLists(gle.EventsLists),
		stats:        protoBuf2ComputedStatsMap(gle.Stats),
	}
	le.setupLoadedEvents()
	le.setupLoadedStates()
	return le
}

func ser2gr(ser *Rights) *opeloggrpc.Rights {
	if ser == nil {
		return nil
	}
	return &opeloggrpc.Rights{Read: ser.Read, Write: ser.Write, Execute: ser.Execute}
}

func storedEntry2ProtoBuf(se *StoredEntry) *opeloggrpc.StoredEntry {
	if se == nil {
		return nil
	}
	return &opeloggrpc.StoredEntry{
		IsDir:         se.IsDir,
		Size:          se.Size,
		Mtime:         se.Mtime,
		User:          se.User,
		UserRights:    ser2gr(se.UserRights),
		Group:         se.Group,
		GroupRights:   ser2gr(se.GroupRights),
		OtherRights:   ser2gr(se.OtherRights),
		IsSymLink:     se.IsSymLink,
		SymLinkTarget: se.SymLinkTarget,
		Children:      slices.Clone(se.Children),
		AddMeta:       bytes.Clone(se.AddMeta),
	}

}

func storedEntries2ProtoBuf(ses []*StoredEntry) []*opeloggrpc.StoredEntry {
	if ses == nil {
		return nil
	}
	gses := make([]*opeloggrpc.StoredEntry, len(ses))
	for i, se := range ses {
		gses[i] = storedEntry2ProtoBuf(se)
	}
	return gses
}

func event2ProtoBuf(ev *Event) *opeloggrpc.Event {
	if ev == nil {
		return nil
	}
	return &opeloggrpc.Event{
		IsTarget:  ev.IsTarget,
		Kind:      opeloggrpc.EventCode(ev.Kind),
		TimeStamp: ev.TimeStamp,
		SeNum:     ev.seNum,
		TcsNums:   slices.Clone(ev.tcsNums),
	}
}

func events2ProtoBuf(evs []*Event) *opeloggrpc.Events {
	if evs == nil {
		return nil
	}
	gevs := make([]*opeloggrpc.Event, len(evs))
	for i, ev := range evs {
		gevs[i] = event2ProtoBuf(ev)
	}
	return &opeloggrpc.Events{EventsList: gevs}
}

func eventsLists2ProtoBuf(evsLs map[int64][]*Event) map[int64]*opeloggrpc.Events {
	if len(evsLs) == 0 {
		return nil
	}
	gevsls := make(map[int64]*opeloggrpc.Events, len(evsLs))
	for k, v := range maps.All(evsLs) {
		gevsls[k] = events2ProtoBuf(v)
	}
	return gevsls
}

func tcss2ProtoBuf(tcss []*TypedChecksum) []*opeloggrpc.TypedChecksum {
	if tcss == nil {
		return nil
	}
	gtcss := make([]*opeloggrpc.TypedChecksum, len(tcss))
	for i, tcs := range tcss {
		gtcss[i] = &opeloggrpc.TypedChecksum{Tcs: bytes.Clone(tcs.Tcs)}
	}
	return gtcss
}

func aggInfo2ProtoBuf(ai *AggInfo) *opeloggrpc.AggInfo {
	if ai == nil {
		return nil
	}
	return &opeloggrpc.AggInfo{Number: ai.Number, Size: ai.Size}
}

func computedStats2ProtoBuf(cs *ComputedStats) *opeloggrpc.ComputedStats {
	if cs == nil {
		return nil
	}
	return &opeloggrpc.ComputedStats{
		SourceListOrStat: aggInfo2ProtoBuf(cs.SourceListOrStat),
		TargetListOrStat: aggInfo2ProtoBuf(cs.TargetListOrStat),
		Read:             aggInfo2ProtoBuf(cs.Read),
		Create:           aggInfo2ProtoBuf(cs.Create),
		Update:           aggInfo2ProtoBuf(cs.Update),
		Remove:           aggInfo2ProtoBuf(cs.Remove),
		MetaChange:       aggInfo2ProtoBuf(cs.MetaChange),
		Error:            aggInfo2ProtoBuf(cs.Error),
	}
}

func computedStatsMap2ProtoBuf(css map[int64]*ComputedStats) map[int64]*opeloggrpc.ComputedStats {
	if css == nil {
		return nil
	}
	gcss := make(map[int64]*opeloggrpc.ComputedStats, len(css))
	for k, v := range maps.All(css) {
		gcss[k] = computedStats2ProtoBuf(v)
	}
	return gcss
}

func state2ProtoBuf(st *State) *opeloggrpc.State {
	if st == nil {
		return nil
	}
	return &opeloggrpc.State{
		Stc:      opeloggrpc.StateCode(st.Stc),
		SeNum:    st.seNum,
		TcsNums:  slices.Clone(st.tcsNums),
		DepCount: st.DepCount,
	}
}

func states2ProtoBuf(sts map[int64]*State) map[int64]*opeloggrpc.State {
	if sts == nil {
		return nil
	}
	gsts := make(map[int64]*opeloggrpc.State, len(sts))
	for k, v := range maps.All(sts) {
		gsts[k] = state2ProtoBuf(v)
	}
	return gsts
}

func LogicalEntry2ProtoBuf(le *LogicalEntry) *opeloggrpc.LogicalEntry {
	if le == nil {
		return nil
	}
	return &opeloggrpc.LogicalEntry{
		SharedSes:    storedEntries2ProtoBuf(le.sharedSes),
		SharedTcss:   tcss2ProtoBuf(le.sharedTcss),
		SourceStates: states2ProtoBuf(le.sourceStates),
		TargetStates: states2ProtoBuf(le.targetStates),
		EventsLists:  eventsLists2ProtoBuf(le.eventsLists),
		Stats:        computedStatsMap2ProtoBuf(le.stats),
	}
}
