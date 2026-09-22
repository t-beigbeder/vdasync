package opelog

import (
	"bytes"
	"slices"

	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

func gr2ser(gr *opeloggrpc.Rights) *Rights {
	if gr == nil {
		return nil
	}
	return &Rights{Read: gr.Read, Write: gr.Write, Execute: gr.Execute}
}

func GrpcStoredEntry2StoredEntry(gse *opeloggrpc.StoredEntry) *StoredEntry {
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

func gses2ses(gses []*opeloggrpc.StoredEntry) []*StoredEntry {
	if gses == nil {
		return nil
	}
	ses := make([]*StoredEntry, len(gses))
	for i, gse := range gses {
		ses[i] = GrpcStoredEntry2StoredEntry(gse)
	}
	return ses
}

func GrpcEvent2Event(ge *opeloggrpc.Event) *Event {
	if ge == nil {
		return nil
	}
	return &Event{
		Kind:      EventCode(ge.Kind),
		TimeStamp: ge.TimeStamp,
		SeNum:     ge.SeNum,
		TcsNums:   slices.Clone(ge.TcsNums),
		Error:     ge.Error,
	}
}

func gevs2evs(gevs []*opeloggrpc.Event) []*Event {
	if gevs == nil {
		return nil
	}
	evs := make([]*Event, len(gevs))
	for i, gev := range gevs {
		evs[i] = GrpcEvent2Event(gev)
	}
	return evs
}

func gai2ai(gai *opeloggrpc.AggInfo) *AggInfo {
	if gai == nil {
		return nil
	}
	return &AggInfo{Number: gai.Number, Size: gai.Size}
}

func GrpcComputedStats2ComputedStats(gcs *opeloggrpc.ComputedStats) *ComputedStats {
	if gcs == nil {
		return nil
	}
	return &ComputedStats{
		TimeStamp:        gcs.TimeStamp,
		SourceListOrStat: gai2ai(gcs.SourceListOrStat),
		TargetListOrStat: gai2ai(gcs.TargetListOrStat),
		Read:             gai2ai(gcs.Read),
		Create:           gai2ai(gcs.Create),
		Update:           gai2ai(gcs.Update),
		Remove:           gai2ai(gcs.Remove),
		MetaChange:       gai2ai(gcs.MetaChange),
		Error:            gai2ai(gcs.Error),
	}
}

func gcss2css(gcss []*opeloggrpc.ComputedStats) []*ComputedStats {
	if gcss == nil {
		return nil
	}
	css := make([]*ComputedStats, len(gcss))
	for i, gcs := range gcss {
		css[i] = GrpcComputedStats2ComputedStats(gcs)
	}
	return css
}

func gtcss2tcss(gtcss []*opeloggrpc.TypedChecksum) []*TypedChecksum {
	if gtcss == nil {
		return nil
	}
	tcss := make([]*TypedChecksum, len(gtcss))
	for i, gtcs := range gtcss {
		tcss[i].Tcs = bytes.Clone(gtcs.Tcs)
	}
	return tcss
}

func GrpcLogicalEntry2LogicalEntry(gle *opeloggrpc.LogicalEntry) *LogicalEntry {
	if gle == nil {
		return nil
	}
	return &LogicalEntry{
		SharedSes:    gses2ses(gle.SharedSes),
		SharedTcss:   gtcss2tcss(gle.SharedTcss),
		SourcePrc:    ProcessingCode(gle.SourcePrc),
		SourceEvents: gevs2evs(gle.SourceEvents),
		TargetPrc:    ProcessingCode(gle.TargetPrc),
		TargetEvents: gevs2evs(gle.TargetEvents),
		DepCount:     gle.DepCount,
		StatsList:    gcss2css(gle.StatsList),
	}
}

func ser2gr(ser *Rights) *opeloggrpc.Rights {
	if ser == nil {
		return nil
	}
	return &opeloggrpc.Rights{Read: ser.Read, Write: ser.Write, Execute: ser.Execute}
}

func StoredEntry2GrpcStoredEntry(se *StoredEntry) *opeloggrpc.StoredEntry {
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

func ses2gses(ses []*StoredEntry) []*opeloggrpc.StoredEntry {
	if ses == nil {
		return nil
	}
	gses := make([]*opeloggrpc.StoredEntry, len(ses))
	for i, se := range ses {
		gses[i] = StoredEntry2GrpcStoredEntry(se)
	}
	return gses
}

func Event2GrpcEvent(ev *Event) *opeloggrpc.Event {
	if ev == nil {
		return nil
	}
	return &opeloggrpc.Event{
		Kind:      opeloggrpc.EventCode(ev.Kind),
		TimeStamp: ev.TimeStamp,
		SeNum:     ev.SeNum,
		TcsNums:   slices.Clone(ev.TcsNums),
		Error:     ev.Error,
	}
}

func evs2gevs(evs []*Event) []*opeloggrpc.Event {
	if evs == nil {
		return nil
	}
	gevs := make([]*opeloggrpc.Event, len(evs))
	for i, ev := range evs {
		gevs[i] = Event2GrpcEvent(ev)
	}
	return gevs
}

func ai2gai(ai *AggInfo) *opeloggrpc.AggInfo {
	if ai == nil {
		return nil
	}
	return &opeloggrpc.AggInfo{Number: ai.Number, Size: ai.Size}
}

func ComputedStats2GrpcComputedStats(cs *ComputedStats) *opeloggrpc.ComputedStats {
	if cs == nil {
		return nil
	}
	return &opeloggrpc.ComputedStats{
		TimeStamp:        cs.TimeStamp,
		SourceListOrStat: ai2gai(cs.SourceListOrStat),
		TargetListOrStat: ai2gai(cs.TargetListOrStat),
		Read:             ai2gai(cs.Read),
		Create:           ai2gai(cs.Create),
		Update:           ai2gai(cs.Update),
		Remove:           ai2gai(cs.Remove),
		MetaChange:       ai2gai(cs.MetaChange),
		Error:            ai2gai(cs.Error),
	}
}

func css2gcss(css []*ComputedStats) []*opeloggrpc.ComputedStats {
	if css == nil {
		return nil
	}
	gcss := make([]*opeloggrpc.ComputedStats, len(css))
	for i, cs := range css {
		gcss[i] = ComputedStats2GrpcComputedStats(cs)
	}
	return gcss
}

func tcss2gtcss(tcss []*TypedChecksum) []*opeloggrpc.TypedChecksum {
	if tcss == nil {
		return nil
	}
	gtcss := make([]*opeloggrpc.TypedChecksum, len(tcss))
	for i, tcs := range tcss {
		gtcss[i].Tcs = bytes.Clone(tcs.Tcs)
	}
	return gtcss
}

func LogicalEntry2GrpcLogicalEntry(le *LogicalEntry) *opeloggrpc.LogicalEntry {
	if le == nil {
		return nil
	}
	return &opeloggrpc.LogicalEntry{
		SharedSes:    ses2gses(le.SharedSes),
		SharedTcss:   tcss2gtcss(le.SharedTcss),
		SourcePrc:    opeloggrpc.ProcessingCode(le.SourcePrc),
		SourceEvents: evs2gevs(le.SourceEvents),
		TargetPrc:    opeloggrpc.ProcessingCode(le.TargetPrc),
		TargetEvents: evs2gevs(le.TargetEvents),
		DepCount:     le.DepCount,
		StatsList:    css2gcss(le.StatsList),
	}
}
