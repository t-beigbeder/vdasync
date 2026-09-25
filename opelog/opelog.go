package opelog

import (
	"bytes"
	"slices"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	Create(source, target string) error
	Open(readOnly bool) error
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
	EVT_UNSPECIFIED      = EventCode(opeloggrpc.EventCode_EVT_UNSPECIFIED)
	EVT_INV_LOADED       = EventCode(opeloggrpc.EventCode_EVT_INV_LOADED)
	EVT_LOADED           = EventCode(opeloggrpc.EventCode_EVT_LOADED)
	EVT_CREATED          = EventCode(opeloggrpc.EventCode_EVT_CREATED)
	EVT_REMOVED          = EventCode(opeloggrpc.EventCode_EVT_REMOVED)
	EVT_UPDATED          = EventCode(opeloggrpc.EventCode_EVT_UPDATED)
	EVT_META_CHANGED     = EventCode(opeloggrpc.EventCode_EVT_META_CHANGED)
	EVT_VERIF_PASSED     = EventCode(opeloggrpc.EventCode_EVT_VERIF_PASSED)
	EVT_VERIF_FAILED     = EventCode(opeloggrpc.EventCode_EVT_VERIF_FAILED)
	EVT_STE_ERR_RAISED   = EventCode(opeloggrpc.EventCode_EVT_STE_ERR_RAISED)
	EVT_OTHER_ERR_RAISED = EventCode(opeloggrpc.EventCode_EVT_OTHER_ERR_RAISED)
)

func (ec EventCode) String() string {
	return opeloggrpc.EventCode(ec).String()
}

type Event struct {
	Kind      EventCode
	TimeStamp int64
	// this should refer to a timestamped_action
	VerifiedOn int64
	SeNum      int32
	TcsNums    []int32
	Error      string
}

func (ev *Event) HasState() bool {
	switch ev.Kind {
	case EVT_LOADED:
		return true
	case EVT_CREATED:
		return true
	case EVT_REMOVED:
		return true
	case EVT_UPDATED:
		return true
	case EVT_META_CHANGED:
		return true
	default:
		return false
	}
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

type ProcessingCode opeloggrpc.ProcessingCode

const (
	PRC_UNSPECIFIED = ProcessingCode(opeloggrpc.ProcessingCode_PRC_UNSPECIFIED)
	PRC_LOADING     = ProcessingCode(opeloggrpc.ProcessingCode_PRC_LOADING)
	PRC_CREATING    = ProcessingCode(opeloggrpc.ProcessingCode_PRC_CREATING)
	PRC_REMOVING    = ProcessingCode(opeloggrpc.ProcessingCode_PRC_REMOVING)
	PRC_UPDATING    = ProcessingCode(opeloggrpc.ProcessingCode_PRC_UPDATING)
	PRC_VERIFYING   = ProcessingCode(opeloggrpc.ProcessingCode_PRC_VERIFYING)
	PRC_PRESENT     = ProcessingCode(opeloggrpc.ProcessingCode_PRC_PRESENT)
	PRC_ABSENT      = ProcessingCode(opeloggrpc.ProcessingCode_PRC_ABSENT)
)

func (prc ProcessingCode) String() string {
	return opeloggrpc.ProcessingCode(prc).String()
}

type LogicalEntry struct {
	SharedSes      []*StoredEntry
	SharedTcss     []*TypedChecksum
	SourcePrc      ProcessingCode
	SourceEvents   []*Event
	SourceDepCount int32
	TargetPrc      ProcessingCode
	TargetEvents   []*Event
	TargetDepCount int32
	StatsList      []*ComputedStats
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
	// the key is a label that may be used to reference existing actions
	TimestampedActions map[string]int64
}
