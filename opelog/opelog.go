package opelog

import (
	"github.com/t-beigbeder/vdasync/opeloggrpc"
)

type OpeLogManager interface {
	NewSession() error
	EndSession() error
	Init(source, target string) error
	PutEntryLog(relPath string, ole *OpeLogEntry) error
}

type OpeCode opeloggrpc.OpeCode

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
	return &StoredEntry{
		Size:          gse.Size,
		Mtime:         gse.Mtime,
		Uid:           gse.Uid,
		Gid:           gse.Gid,
		Mode:          gse.Mode,
		SymLinkTarget: gse.SymLinkTarget,
	}
}

func StoredEntry2GrpcStoredEntry(gse *StoredEntry) *opeloggrpc.StoredEntry {
	return &opeloggrpc.StoredEntry{
		Size:          gse.Size,
		Mtime:         gse.Mtime,
		Uid:           gse.Uid,
		Gid:           gse.Gid,
		Mode:          gse.Mode,
		SymLinkTarget: gse.SymLinkTarget,
	}
}
