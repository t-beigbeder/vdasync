package opelog

import "github.com/t-beigbeder/vdasync/opeloggrpc"

type OpeLogManager interface {
	NewSession() error
	EndSession() error
	Init(source, target string) error
	PutEntryLog(relPath string, ole *opeloggrpc.OpeLogEntry) error
}
