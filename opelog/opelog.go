package opelog

import (
	"bytes"
	"maps"
	"slices"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	Create(session, inventory, source, target string) (int64, int64, error)
	AddSession(label string) (int64, error)
	AddInventory(label string) (int64, error)
	Open(session, inventory string, readOnly bool) (int64, int64, error)
	Sync() error
	Close() error
	PutLogicalEntry(relPath string, ole *LogicalEntry) error
	GetLogicalEntry(relPath string) (*LogicalEntry, error)
	Walk(func(relPath string, ole *LogicalEntry) error) error
}

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

func (or *Rights) Clone() *Rights {
	return &Rights{or.Read, or.Write, or.Execute}
}

func (nr *Rights) Equal(or *Rights) (result bool) {
	if nr.Read != or.Read {
		return
	}
	if nr.Write != or.Write {
		return
	}
	if nr.Execute != or.Execute {
		return
	}
	result = true
	return
}

func cmpRights(nrp, orp **Rights) (result bool) {
	if *nrp == nil && *orp == nil {
		return true
	}
	if *nrp == nil || *orp == nil {
		return
	}
	return (*nrp).Equal(*orp)
}

type StoredEntry struct {
	IsDir         bool
	Size          int64
	Mtime         int64
	User          int32
	UserRights    *Rights
	Group         int32
	GroupRights   *Rights
	OtherRights   *Rights
	IsSymLink     bool
	SymLinkTarget string
	Children      []string
	AddMeta       []byte
}

func cloneRights(or *Rights) *Rights {
	if or == nil {
		return nil
	}
	return or.Clone()
}

func (tse *StoredEntry) Clone() *StoredEntry {
	return &StoredEntry{
		IsDir:         tse.IsDir,
		Size:          tse.Size,
		Mtime:         tse.Mtime,
		User:          int32(tse.User),
		UserRights:    cloneRights(tse.UserRights),
		Group:         int32(tse.Group),
		GroupRights:   cloneRights(tse.GroupRights),
		OtherRights:   cloneRights(tse.OtherRights),
		IsSymLink:     tse.IsSymLink,
		SymLinkTarget: tse.SymLinkTarget,
		Children:      slices.Clone(tse.Children),
		AddMeta:       bytes.Clone(tse.AddMeta),
	}
}

func (tse *StoredEntry) HasChild(child string) bool {
	return slices.Contains(tse.Children, child)
}

func (tse *StoredEntry) Equal(ose *StoredEntry, noEqMtime, noMtime, noMtLink, noRights bool) (result bool) {
	if ose == nil {
		return
	}
	if tse.IsDir != ose.IsDir {
		return
	}
	if tse.Size != ose.Size {
		return
	}
	if tse.Mtime != ose.Mtime {
		if !noEqMtime {
			return
		}
		if ose.Mtime > tse.Mtime {
			return
		}
		if ose.IsSymLink && !noMtime && !noMtLink {
			return
		}
		if !ose.IsSymLink && !noMtime {
			return
		}
	}
	if !noRights {
		if tse.User != ose.User {
			return
		}
		if !cmpRights(&tse.UserRights, &ose.UserRights) {
			return
		}
		if tse.Group != ose.Group {
			return
		}
		if !cmpRights(&tse.GroupRights, &ose.GroupRights) {
			return
		}
		if !cmpRights(&tse.OtherRights, &ose.OtherRights) {
			return
		}
	}
	if tse.IsSymLink != ose.IsSymLink {
		return
	}
	if tse.SymLinkTarget != ose.SymLinkTarget {
		return
	}
	if len(tse.Children) != len(ose.Children) {
		return
	}
	for _, nChild := range tse.Children {
		if !ose.HasChild(nChild) {
			return
		}
	}
	if !bytes.Equal(tse.AddMeta, ose.AddMeta) {
		return
	}
	result = true
	return
}

func (tse *StoredEntry) EqualType(ose *StoredEntry) bool {
	if ose == nil {
		return false
	}
	return tse.IsDir == ose.IsDir && tse.IsSymLink == ose.IsSymLink
}

func dr2r(dr *dssa.Rights) *Rights {
	return &Rights{Read: dr.Read, Write: dr.Write, Execute: dr.Execute}
}

func FromDataEntry(dse *dssa.DataEntry, children []string) *StoredEntry {
	return &StoredEntry{
		IsDir:         dse.IsDir,
		Size:          dse.Size,
		Mtime:         dse.Mtime,
		User:          int32(dse.User),
		UserRights:    dr2r(&dse.UserRights),
		Group:         int32(dse.Group),
		GroupRights:   dr2r(&dse.GroupRights),
		OtherRights:   dr2r(&dse.OtherRights),
		IsSymLink:     dse.IsSymLink,
		SymLinkTarget: dse.SymLinkTarget,
		Children:      slices.Clone(children),
		AddMeta:       bytes.Clone(dse.AddMeta),
	}
}

func r2dr(r *Rights) *dssa.Rights {
	return &dssa.Rights{Read: r.Read, Write: r.Write, Execute: r.Execute}
}

func (tse *StoredEntry) ToDataEntry(path_ string) *dssa.DataEntry {
	return &dssa.DataEntry{
		IsDir:         tse.IsDir,
		Path:          path_,
		Size:          tse.Size,
		Mtime:         tse.Mtime,
		User:          int(tse.User),
		UserRights:    *r2dr(tse.UserRights),
		Group:         int(tse.Group),
		GroupRights:   *r2dr(tse.GroupRights),
		OtherRights:   *r2dr(tse.OtherRights),
		IsSymLink:     tse.IsSymLink,
		SymLinkTarget: tse.SymLinkTarget,
		AddMeta:       bytes.Clone(tse.AddMeta),
	}
}

type TypedChecksum struct {
	// first byte is HalgoCode, checksum comes after
	Tcs []byte
}

func (tcs *TypedChecksum) Clone() *TypedChecksum {
	return &TypedChecksum{bytes.Clone(tcs.Tcs)}
}

type EventCode opeloggrpc.EventCode

const (
	EVC_UNSPECIFIED  = EventCode(opeloggrpc.EventCode_EVC_UNSPECIFIED)
	EVC_LOADED       = EventCode(opeloggrpc.EventCode_EVC_LOADED)
	EVC_REMOVED      = EventCode(opeloggrpc.EventCode_EVC_REMOVED)
	EVC_CREATED      = EventCode(opeloggrpc.EventCode_EVC_CREATED)
	EVC_UPDATED      = EventCode(opeloggrpc.EventCode_EVC_UPDATED)
	EVC_META_CHANGED = EventCode(opeloggrpc.EventCode_EVC_META_CHANGED)
)

func (ec EventCode) String() string {
	return opeloggrpc.EventCode(ec).String()
}

type Event struct {
	IsTarget  bool
	Kind      EventCode
	TimeStamp int64
	Se        *StoredEntry
	Tcss      [][]byte
	// values shared by index, negative index means no value
	seNum   int32
	tcsNums []int32
}

type AggInfo struct {
	Number int64
	Size   int64
}

type ComputedStats struct {
	SourceListOrStat *AggInfo
	TargetListOrStat *AggInfo
	Read             *AggInfo
	Create           *AggInfo
	Update           *AggInfo
	Remove           *AggInfo
	MetaChange       *AggInfo
	NoOp             *AggInfo
	Error            *AggInfo
}

func (cs *ComputedStats) Reset() {
	cs.SourceListOrStat = &AggInfo{}
	cs.TargetListOrStat = &AggInfo{}
	cs.Read = &AggInfo{}
	cs.Create = &AggInfo{}
	cs.Update = &AggInfo{}
	cs.Remove = &AggInfo{}
	cs.MetaChange = &AggInfo{}
	cs.NoOp = &AggInfo{}
	cs.Error = &AggInfo{}
}

func (cs *ComputedStats) IsSet() bool {
	return cs.SourceListOrStat.Number+cs.TargetListOrStat.Number+
		cs.Read.Number+cs.Create.Number+
		cs.Update.Number+cs.Remove.Number+cs.MetaChange.Number+cs.Error.Number == 0

}

type StateCode opeloggrpc.StateCode

const (
	STC_UNSPECIFIED  = StateCode(opeloggrpc.StateCode_STC_UNSPECIFIED)
	STC_DONE_ABSENT  = StateCode(opeloggrpc.StateCode_STC_DONE_ABSENT)
	STC_DONE_PRESENT = StateCode(opeloggrpc.StateCode_STC_DONE_PRESENT)
	STC_SE_ERROR     = StateCode(opeloggrpc.StateCode_STC_SE_ERROR)
	// error from descendants are propagated, loading is partial, modification is blocked
	// redo will retry required actions
	STC_DESC_ERROR = StateCode(opeloggrpc.StateCode_STC_DESC_ERROR)
)

func (sc StateCode) String() string {
	return opeloggrpc.StateCode(sc).String()
}

type State struct {
	Stc StateCode
	// after tool (re)start, entry with on-going STC_DIR state
	// needs children requeuing after proper evaluation
	ToolStartTime int64
	Error         string
	Se            *StoredEntry
	Tcss          [][]byte
	// value -1 by convention is a signal from last child sending notification to parent
	DepCount int32
	// values shared by index, negative index means no value
	seNum   int32
	tcsNums []int32
}

type LogicalEntry struct {
	// Checksums and stored entries are shared and referenced by index
	sharedSes    []*StoredEntry
	sharedTcss   []*TypedChecksum
	sourceStates map[int64]*State
	targetStates map[int64]*State
	eventsLists  map[int64][]*Event
	stats        map[int64]*ComputedStats
}

func NewLogicalEntry() *LogicalEntry {
	return new(LogicalEntry(LogicalEntry{
		sourceStates: map[int64]*State{},
		targetStates: map[int64]*State{},
		eventsLists:  map[int64][]*Event{},
		stats:        map[int64]*ComputedStats{},
	}))
}

func (le *LogicalEntry) CreateEvent(sessionTs, eventTs int64, isTarget bool, kind EventCode, se *StoredEntry, tcss [][]byte) *Event {
	evs, _ := le.eventsLists[sessionTs]
	if eventTs == 0 {
		eventTs = time.Now().Unix()
	}
	ev := &Event{
		IsTarget:  isTarget,
		Kind:      kind,
		TimeStamp: eventTs,
		Se:        se,
		Tcss:      tcss,
		seNum:     le.addOrShareSe(se),
		tcsNums:   le.addOrShareTcss(tcss),
	}
	le.eventsLists[sessionTs] = append(evs, ev)
	return ev
}

func (le *LogicalEntry) setupLoadedEvents() {
	for evs := range maps.Values(le.eventsLists) {
		for _, ev := range evs {
			if ev.seNum != -1 {
				ev.Se = le.sharedSes[ev.seNum].Clone()
			}
			for _, tcsNum := range ev.tcsNums {
				ev.Tcss = append(ev.Tcss, le.sharedTcss[tcsNum].Clone().Tcs)
			}
		}
	}
}

func (le *LogicalEntry) GetEvents(sessionTs int64, isTarget bool) (res []*Event) {
	evs, ok := le.eventsLists[sessionTs]
	if !ok {
		return nil
	}
	for _, ev := range evs {
		if ev.IsTarget == isTarget {
			res = append(res, ev)
		}
	}
	return
}

func (le *LogicalEntry) SetState(toolStartTime, sessInvTs int64, isTarget bool, stc StateCode, sErr string, se *StoredEntry, tcss [][]byte, depCount int) *State {
	st := &State{
		Stc:           stc,
		ToolStartTime: toolStartTime,
		Error:         sErr,
		Se:            se,
		Tcss:          tcss,
		DepCount:      int32(depCount),
		seNum:         le.addOrShareSe(se),
		tcsNums:       le.addOrShareTcss(tcss),
	}
	if isTarget {
		le.targetStates[sessInvTs] = st
	} else {
		le.sourceStates[sessInvTs] = st
	}
	return st
}

func (le *LogicalEntry) setupLoadedStates() {
	for st := range common.Concat(maps.Values(le.sourceStates), maps.Values(le.targetStates)) {
		if st.seNum != -1 {
			st.Se = le.sharedSes[st.seNum].Clone()
		}
		for _, tcsNum := range st.tcsNums {
			st.Tcss = append(st.Tcss, le.sharedTcss[tcsNum].Clone().Tcs)
		}
	}
}

func (le *LogicalEntry) GetState(sessInvTs int64, isTarget bool) *State {
	var (
		st *State
		ok bool
	)
	if isTarget {
		st, ok = le.targetStates[sessInvTs]
	} else {
		st, ok = le.sourceStates[sessInvTs]
	}
	if !ok {
		return nil
	}
	return st
}

func (le *LogicalEntry) GetStats(sessionTs int64) *ComputedStats {
	cst, _ := le.stats[sessionTs]
	return cst
}

func NewComputedStats() *ComputedStats {
	return &ComputedStats{SourceListOrStat: &AggInfo{}, TargetListOrStat: &AggInfo{},
		Read: &AggInfo{}, Create: &AggInfo{}, Update: &AggInfo{}, Remove: &AggInfo{}, MetaChange: &AggInfo{}, Error: &AggInfo{},
	}
}

func (le *LogicalEntry) addOrShareSe(se *StoredEntry) int32 {
	if se == nil {
		return -1
	}
	for i, exSe := range slices.Backward(le.sharedSes) {
		if se.Equal(exSe, false, false, false, false) {
			return int32(i)
		}
	}
	le.sharedSes = append(le.sharedSes, se)
	return int32(len(le.sharedSes) - 1)
}

func (le *LogicalEntry) addOrShareTcs(tcs []byte) int32 {
	if tcs == nil {
		return -1
	}
	for i, exTcs := range slices.Backward(le.sharedTcss) {
		if bytes.Equal(tcs, exTcs.Tcs) {
			return int32(i)
		}
	}
	le.sharedTcss = append(le.sharedTcss, &TypedChecksum{Tcs: bytes.Clone(tcs)})
	return int32(len(le.sharedTcss) - 1)
}

func (le *LogicalEntry) addOrShareTcss(tcss [][]byte) []int32 {
	if tcss == nil {
		return nil
	}
	tcssNums := make([]int32, len(tcss))
	for i := range len(tcss) {
		tcssNums[i] = le.addOrShareTcs(tcss[i])
	}
	return tcssNums
}

// For a memory to file simple implementation, limited to 2GiB, cf https://protobuf.dev/programming-guides/proto-limits/#total
type OpeLogAllInOne struct {
	SourceRoot string
	TargetRoot string
	// the key is the relative path
	LogicalEntries map[string]*LogicalEntry
	// the key is the session label the value is its unique reference time
	Sessions map[string]int64
	// the key is the inventory label the value is its unique reference time
	Inventories map[string]int64
}
