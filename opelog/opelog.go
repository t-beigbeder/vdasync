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

func (ose *StoredEntry) Clone() *StoredEntry {
	return &StoredEntry{
		IsDir:         ose.IsDir,
		Size:          ose.Size,
		Mtime:         ose.Mtime,
		User:          int32(ose.User),
		UserRights:    cloneRights(ose.UserRights),
		Group:         int32(ose.Group),
		GroupRights:   cloneRights(ose.GroupRights),
		OtherRights:   cloneRights(ose.OtherRights),
		IsSymLink:     ose.IsSymLink,
		SymLinkTarget: ose.SymLinkTarget,
		Children:      slices.Clone(ose.Children),
		AddMeta:       bytes.Clone(ose.AddMeta),
	}
}

func (se *StoredEntry) HasChild(cChild string) bool {
	return slices.Contains(se.Children, cChild)
}

func (nse *StoredEntry) Equal(ose *StoredEntry) (result bool) {
	if nse == nil && ose == nil {
		return true
	}
	if nse == nil || ose == nil {
		return
	}
	if nse.IsDir != ose.IsDir {
		return
	}
	if nse.Size != ose.Size {
		return
	}
	if nse.Mtime != ose.Mtime {
		return
	}
	if nse.User != ose.User {
		return
	}
	if !cmpRights(&nse.UserRights, &ose.UserRights) {
		return
	}
	if nse.Group != ose.Group {
		return
	}
	if !cmpRights(&nse.GroupRights, &ose.GroupRights) {
		return
	}
	if !cmpRights(&nse.OtherRights, &ose.OtherRights) {
		return
	}
	if nse.IsSymLink != ose.IsSymLink {
		return
	}
	if nse.SymLinkTarget != ose.SymLinkTarget {
		return
	}
	if len(nse.Children) != len(ose.Children) {
		return
	}
	for _, nChild := range nse.Children {
		if !ose.HasChild(nChild) {
			return
		}
	}
	if !bytes.Equal(nse.AddMeta, ose.AddMeta) {
		return
	}
	result = true
	return
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

func (ose *StoredEntry) ToDataEntry(path_ string) *dssa.DataEntry {
	return &dssa.DataEntry{
		IsDir:         ose.IsDir,
		Path:          path_,
		Size:          ose.Size,
		Mtime:         ose.Mtime,
		User:          int(ose.User),
		UserRights:    *r2dr(ose.UserRights),
		Group:         int(ose.Group),
		GroupRights:   *r2dr(ose.GroupRights),
		OtherRights:   *r2dr(ose.OtherRights),
		IsSymLink:     ose.IsSymLink,
		SymLinkTarget: ose.SymLinkTarget,
		AddMeta:       bytes.Clone(ose.AddMeta),
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
	Error            *AggInfo
}

type StateCode opeloggrpc.StateCode

const (
	STC_UNSPECIFIED  = StateCode(opeloggrpc.StateCode_STC_UNSPECIFIED)
	STC_DONE_ABSENT  = StateCode(opeloggrpc.StateCode_STC_DONE_ABSENT)
	STC_DONE_PRESENT = StateCode(opeloggrpc.StateCode_STC_DONE_PRESENT)
	STC_DIR_LOAD     = StateCode(opeloggrpc.StateCode_STC_DIR_LOAD)
	STC_DIR_RM       = StateCode(opeloggrpc.StateCode_STC_DIR_RM)
	STC_DIR_CHANGE   = StateCode(opeloggrpc.StateCode_STC_DIR_CHANGE)
	STC_SE_ERROR     = StateCode(opeloggrpc.StateCode_STC_SE_ERROR)
	// error from descendants are propagated, loading is partial, modification is blocked
	// redo will retry required actions
	STC_DESC_ERROR = StateCode(opeloggrpc.StateCode_STC_DESC_ERROR)
)

func (sc StateCode) String() string {
	return opeloggrpc.StateCode(sc).String()
}

type State struct {
	Stc      StateCode
	Se       *StoredEntry
	Tcss     [][]byte
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

func (le *LogicalEntry) CreateEvent(sessionTs int64, isTarget bool, kind EventCode, se *StoredEntry, tcss [][]byte) *Event {
	evs, _ := le.eventsLists[sessionTs]
	ev := &Event{
		IsTarget:  isTarget,
		Kind:      kind,
		TimeStamp: time.Now().Unix(),
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

func (le *LogicalEntry) SetState(sessInvTs int64, isTarget bool, stc StateCode, se *StoredEntry, tcss [][]byte, depCount int) *State {
	st := &State{
		Stc:      stc,
		Se:       se,
		Tcss:     tcss,
		DepCount: int32(depCount),
		seNum:    le.addOrShareSe(se),
		tcsNums:  le.addOrShareTcss(tcss),
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

func (le *LogicalEntry) GetState(sessionTs int64, isTarget bool) *State {
	var (
		st *State
		ok bool
	)
	if isTarget {
		st, ok = le.targetStates[sessionTs]
	} else {
		st, ok = le.sourceStates[sessionTs]
	}
	if !ok {
		return nil
	}
	return st
}

func (le *LogicalEntry) addOrShareSe(se *StoredEntry) int32 {
	if se == nil {
		return -1
	}
	for i, exSe := range slices.Backward(le.sharedSes) {
		if se.Equal(exSe) {
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
