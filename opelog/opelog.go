package opelog

import (
	"bytes"
	"slices"

	"github.com/t-beigbeder/vdasync/dssa"
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
	Kind        EventCode
	SessionTime int64
	TimeStamp   int64
	// negative number means no entry
	SeNum   int32
	TcsNums []int32
}

type AggInfo struct {
	Number int64
	Size   int64
}

type ComputedStats struct {
	TimeStamp        int64
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
	Stc StateCode
	// values shared by index, negative index means no value
	SeNum    int32
	TcsNums  []int32
	DepCount int32
}

type LogicalEntry struct {
	// Checksums and stored entries are shared and referenced by index
	SharedSes    []*StoredEntry
	SharedTcss   []*TypedChecksum
	SourceStates map[int64]*State
	SourceEvents []*Event
	TargetStates map[int64]*State
	TargetEvents []*Event
	// when last_session changes, on-going dir states need to be recomputed
	LastSession int64
	StatsList   []*ComputedStats
}

func (le *LogicalEntry) AddOrShareSe(se *StoredEntry) int32 {
	if se == nil {
		return -1
	}
	for i, exSe := range slices.Backward(le.SharedSes) {
		if se.Equal(exSe) {
			return int32(i)
		}
	}
	le.SharedSes = append(le.SharedSes, se)
	return int32(len(le.SharedSes) - 1)
}

func (le *LogicalEntry) AddOrShareTcs(tcs []byte) int32 {
	if tcs == nil {
		return -1
	}
	for i, exTcs := range slices.Backward(le.SharedTcss) {
		if bytes.Equal(tcs, exTcs.Tcs) {
			return int32(i)
		}
	}
	le.SharedTcss = append(le.SharedTcss, &TypedChecksum{Tcs: bytes.Clone(tcs)})
	return int32(len(le.SharedTcss) - 1)
}

func (le *LogicalEntry) GetTcssFor(ev *Event) [][]byte {
	if len(ev.TcsNums) == 0 {
		return nil
	}
	tcss := make([][]byte, len(ev.TcsNums))
	for i := range ev.TcsNums {
		tcss[i] = bytes.Clone(le.SharedTcss[i].Tcs)
	}
	return tcss
}

func (le *LogicalEntry) AddOrShareTcss(tcss [][]byte) []int32 {
	if tcss == nil {
		return nil
	}
	tcssNums := make([]int32, len(tcss))
	for i := range len(tcss) {
		tcssNums[i] = le.AddOrShareTcs(tcss[i])
	}
	return tcssNums
}

// For a memory to file simple implementation, limited to 2GiB, cf https://protobuf.dev/programming-guides/proto-limits/#total
type OpeLogAllInOne struct {
	SourceRoot string
	TargetRoot string
	// the key is the relative path
	LogicalEntries map[string]*LogicalEntry
	// the key is the session label the value is its start time
	Sessions    map[string]int64
	Inventories map[string]int64
}
