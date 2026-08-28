package opelog

import (
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	NewSession() error
	EndSession() error
	Init(source, target string) error
	PutLogicalEntry(relPath string, ole *LogicalEntry) error
	GetLogicalEntry(relPath string) (*LogicalEntry, error)
}

type EventCode opeloggrpc.EventCode

const (
	EVT_UNSPECIFIED = EventCode(opeloggrpc.EventCode_EVT_UNSPECIFIED)
	EVT_ABS         = EventCode(opeloggrpc.EventCode_EVT_ABS)
	EVT_EXIST       = EventCode(opeloggrpc.EventCode_EVT_EXIST)
	EVT_CR_MOD      = EventCode(opeloggrpc.EventCode_EVT_CR_MOD)
	EVT_ATTS_CHG    = EventCode(opeloggrpc.EventCode_EVT_ATTS_CHG)
	EVT_START_DIRUP = EventCode(opeloggrpc.EventCode_EVT_START_DIRUP)
	EVT_END_DIRUP   = EventCode(opeloggrpc.EventCode_EVT_END_DIRUP)
)

type OriginCode opeloggrpc.OriginCode

const (
	ORI_UNSPECIFIED = OriginCode(opeloggrpc.OriginCode_ORI_UNSPECIFIED)
	ORI_LIST        = OriginCode(opeloggrpc.OriginCode_ORI_LIST)
	ORI_STAT        = OriginCode(opeloggrpc.OriginCode_ORI_STAT)
	ORI_READ        = OriginCode(opeloggrpc.OriginCode_ORI_READ)
	ORI_MKDIR       = OriginCode(opeloggrpc.OriginCode_ORI_MKDIR)
	ORI_WRITE       = OriginCode(opeloggrpc.OriginCode_ORI_WRITE)
	ORI_SET_STAT    = OriginCode(opeloggrpc.OriginCode_ORI_SET_STAT)
	ORI_RM          = OriginCode(opeloggrpc.OriginCode_ORI_RM)
)

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
}

type StoredEntry struct {
	IsPresent     bool
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
}

type Event struct {
	Kind          EventCode
	Origin        OriginCode
	UpdateCount   int32
	ValidateCount int32
	TimeStamp     int64
	StateIndex    int32
	// comma-separated list algo:hexa-of-checksum
	Checksums string
	Error     string
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
	ModChange        *AggInfo
	Error            *AggInfo
}

type LogicalEntry struct {
	// keeping source and target states out of event saves storage when unchanged
	SourceStates  []*StoredEntry
	SourceEvents  []*Event
	TargetStates  []*StoredEntry
	DirupState    *StoredEntry
	DirupChildren []string
	TargetEvents  []*Event
	StatsList     []*ComputedStats
}

// For a memory to file simple implementation, limited to 2GiB, cf https://protobuf.dev/programming-guides/proto-limits/#total
type OpeLogAllInOne struct {
	SourceRoot string
	TargetRoot string
	// the key is the relative path
	LogicalEntries map[string]*LogicalEntry
}

func gr2ser(gr *opeloggrpc.Rights) *Rights {
	return &Rights{Read: gr.Read, Write: gr.Write, Execute: gr.Execute}
}

func GrpcStoredEntry2StoredEntry(gse *opeloggrpc.StoredEntry) *StoredEntry {
	if gse == nil {
		return nil
	}
	children := make([]string, len(gse.Children))
	copy(children, gse.Children)
	return &StoredEntry{
		IsPresent:     gse.IsPresent,
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
		Children:      children,
	}
}
func ser2gr(ser *Rights) *opeloggrpc.Rights {
	return &opeloggrpc.Rights{Read: ser.Read, Write: ser.Write, Execute: ser.Execute}
}

func StoredEntry2GrpcStoredEntry(se *StoredEntry) *opeloggrpc.StoredEntry {
	if se == nil {
		return nil
	}
	children := make([]string, len(se.Children))
	copy(children, se.Children)
	return &opeloggrpc.StoredEntry{
		IsPresent:     se.IsPresent,
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
		Children:      children,
	}
}
