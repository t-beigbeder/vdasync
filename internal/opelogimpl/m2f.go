package opelogimpl

import (
	"errors"
	"fmt"
	"sync"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
	"google.golang.org/protobuf/proto"
)

type m2fMng struct {
	path       string
	mx         sync.Mutex
	source     string
	target     string
	les        map[string]*opelog.LogEntry
	hasSession bool
	hasUpdates bool
}

// GetEntryLog implements [opelog.OpeLogManager].
func (m *m2fMng) GetEntryLog(relPath string) (*opelog.LogEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.hasSession {
		return nil, errors.New("m2fMng.GetEntryLog: no session to get")
	}
	le, ok := m.les[relPath]
	if !ok {
		return nil, fmt.Errorf("m2fMng.GetEntryLog: %s: no such entry", relPath)
	}
	return le, nil
}

func MakeM2fManager(path, source, target string) (opelog.OpeLogManager, error) {
	return &m2fMng{path: path, source: source, target: target}, nil
}

// EndSession implements [opelog.OpeLogManager].
func (m *m2fMng) EndSession() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.hasSession {
		return errors.New("m2fMng.EndSession: no session to end")
	}
	if !m.hasUpdates {
		return nil
	}
	aio := opeloggrpc.OpeLogAllInOne{
		Source:     m.source,
		Target:     m.target,
		LogEntries: make([]*opeloggrpc.LogEntry, len(m.les)),
	}
	i := 0
	for _, le := range m.les {
		gles := &opeloggrpc.LogEntry{
			RelPath:       le.RelPath,
			OpeLogEntries: make([]*opeloggrpc.OpeLogEntry, len(le.OpeLogEntries)),
		}
		for j, ole := range le.OpeLogEntries {
			gles.OpeLogEntries[j] = &opeloggrpc.OpeLogEntry{
				Code:            opeloggrpc.OpeCode(ole.Code),
				Check:           ole.Check,
				TimeStamp:       ole.TimeStamp,
				ErrorId:         ole.ErrorId,
				Source:          opelog.StoredEntry2GrpcStoredEntry(ole.Source),
				Target:          opelog.StoredEntry2GrpcStoredEntry(ole.Target),
				SourceChecksums: ole.SourceChecksums,
				TargetChecksums: ole.TargetChecksums,
			}
		}
		aio.LogEntries[i] = gles
		i++
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.path, bs); err != nil {
		return err
	}
	m.hasUpdates = false
	return nil
}

// Init implements [opelog.OpeLogManager].
func (m *m2fMng) Init(source string, target string) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if common.FileExists(m.path) {
		return fmt.Errorf("m2fMng.Init: %s already exists", m.path)
	}
	if m.les != nil {
		return fmt.Errorf("m2fMng.Init: %s should be created without entries", m.path)
	}
	aio := opeloggrpc.OpeLogAllInOne{
		Source: m.source,
		Target: m.target,
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.path, bs); err != nil {
		return err
	}
	return nil
}

// NewSession implements [opelog.OpeLogManager].
func (m *m2fMng) NewSession() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.hasSession {
		return nil
	}
	bs, err := common.UnsafeLoadFile(m.path)
	if err != nil {
		return err
	}
	var aio opeloggrpc.OpeLogAllInOne
	if err = proto.Unmarshal(bs, &aio); err != nil {
		return err
	}
	m.source = aio.Source
	m.target = aio.Target
	m.les = make(map[string]*opelog.LogEntry, len(aio.LogEntries))
	for _, gle := range aio.LogEntries {
		m.les[gle.RelPath] = &opelog.LogEntry{
			RelPath:       gle.RelPath,
			OpeLogEntries: make([]*opelog.OpeLogEntry, len(gle.OpeLogEntries)),
		}
		le := m.les[gle.RelPath]
		for jx, gole := range gle.OpeLogEntries {
			le.OpeLogEntries[jx] = &opelog.OpeLogEntry{
				Code:            opelog.OpeCode(gole.Code),
				Check:           gole.Check,
				TimeStamp:       gole.TimeStamp,
				ErrorId:         gole.ErrorId,
				Source:          opelog.GrpcStoredEntry2StoredEntry(gole.Source),
				Target:          opelog.GrpcStoredEntry2StoredEntry(gole.Target),
				SourceChecksums: gole.SourceChecksums,
				TargetChecksums: gole.TargetChecksums,
			}
		}
	}
	m.hasSession = true
	return nil
}

// PutEntryLog implements [opelog.OpeLogManager].
func (m *m2fMng) PutEntryLog(relPath string, ole *opelog.OpeLogEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.hasSession {
		return errors.New("m2fMng.PutEntryLog: no session to put")
	}
	le, ok := m.les[relPath]
	if !ok {
		le = &opelog.LogEntry{RelPath: relPath, OpeLogEntries: []*opelog.OpeLogEntry{}}
		m.les[relPath] = le
	}
	le.OpeLogEntries = append(le.OpeLogEntries, ole)
	m.hasUpdates = true
	return nil
}

var _ opelog.OpeLogManager = &m2fMng{}
