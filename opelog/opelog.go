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

type StoredEntry struct {
	Size          int64
	Mtime         int64
	Uid           uint32
	Gid           uint32
	Mode          uint32
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

func GrpcStoredEntry2StoredEntry(gse *opeloggrpc.StoredEntry) *StoredEntry {
	if gse == nil {
		return nil
	}
	children := make([]string, len(gse.Children))
	copy(children, gse.Children)
	return &StoredEntry{
		Size:          gse.Size,
		Mtime:         gse.Mtime,
		Uid:           gse.Uid,
		Gid:           gse.Gid,
		Mode:          gse.Mode,
		SymLinkTarget: gse.SymLinkTarget,
		Children:      children,
	}
}

func StoredEntry2GrpcStoredEntry(se *StoredEntry) *opeloggrpc.StoredEntry {
	if se == nil {
		return nil
	}
	children := make([]string, len(se.Children))
	copy(children, se.Children)
	return &opeloggrpc.StoredEntry{
		Size:          se.Size,
		Mtime:         se.Mtime,
		Uid:           se.Uid,
		Gid:           se.Gid,
		Mode:          se.Mode,
		SymLinkTarget: se.SymLinkTarget,
		Children:      children,
	}
}
