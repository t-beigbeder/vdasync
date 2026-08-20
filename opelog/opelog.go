package opelog

import (
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	NewSession() error
	EndSession() error
	Init(source, target string) error
	PutEntryLog(relPath string, ole *OpeLogEntry) error
	GetEntryLog(relPath string) (*LogEntry, error)
}

type OpeCode opeloggrpc.OpeCode

const (
	OPE_CODE_UNSPECIFIED     = OpeCode(opeloggrpc.OpeCode_OPE_CODE_UNSPECIFIED)
	OPE_CODE_SOURCE_STAT     = OpeCode(opeloggrpc.OpeCode_OPE_CODE_SOURCE_STAT)
	OPE_CODE_TARGET_STAT     = OpeCode(opeloggrpc.OpeCode_OPE_CODE_TARGET_STAT)
	OPE_CODE_SOURCE_CHECKSUM = OpeCode(opeloggrpc.OpeCode_OPE_CODE_SOURCE_CHECKSUM)
	OPE_CODE_TARGET_CHECKSUM = OpeCode(opeloggrpc.OpeCode_OPE_CODE_TARGET_CHECKSUM)
	OPE_CODE_MKDIR           = OpeCode(opeloggrpc.OpeCode_OPE_CODE_MKDIR)
	OPE_CODE_COPY            = OpeCode(opeloggrpc.OpeCode_OPE_CODE_COPY)
	OPE_CODE_SYMLINK         = OpeCode(opeloggrpc.OpeCode_OPE_CODE_SYMLINK)
	OPE_CODE_DELETE          = OpeCode(opeloggrpc.OpeCode_OPE_CODE_DELETE)
	OPE_CODE_SET_STAT        = OpeCode(opeloggrpc.OpeCode_OPE_CODE_SET_STAT)
)

type LogEntry struct {
	RelPath       string
	OpeLogEntries []*OpeLogEntry
}

type Rights struct {
	Read    bool
	Write   bool
	Execute bool
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
}

type OpeLogEntry struct {
	Code            OpeCode
	Check           bool
	TimeStamp       int64
	ErrorId         uint64
	Source          *StoredEntry
	Target          *StoredEntry
	SourceChecksums string
	TargetChecksums string
}

type OpeLogAllInOne struct {
	Source     string
	Target     string
	LogEntries []*LogEntry
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
