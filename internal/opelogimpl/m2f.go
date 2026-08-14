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
	les        map[string]*opeloggrpc.LogEntry
	hasSession bool
	hasUpdates bool
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
		les := &opeloggrpc.LogEntry{
			RelPath:       le.RelPath,
			OpeLogEntries: make([]*opeloggrpc.OpeLogEntry, len(le.OpeLogEntries)),
		}
		oles := les.OpeLogEntries
		for _, ole := range le.OpeLogEntries {
			oles = append(oles, ole)
		}
		aio.LogEntries[i] = les
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
	bs, err := common.LoadFile(m.path)
	if err != nil {
		return err
	}
	var aio opeloggrpc.OpeLogAllInOne
	if err = proto.Unmarshal(bs, &aio); err != nil {
		return err
	}
	m.source = aio.Source
	m.target = aio.Target
	m.les = make(map[string]*opeloggrpc.LogEntry, len(aio.LogEntries))
	for _, le := range aio.LogEntries {
		m.les[le.RelPath] = le
	}
	return nil
}

// PutEntryLog implements [opelog.OpeLogManager].
func (m *m2fMng) PutEntryLog(relPath string, ole *opeloggrpc.OpeLogEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	le, ok := m.les[relPath]
	if !ok {
		le = &opeloggrpc.LogEntry{RelPath: relPath, OpeLogEntries: []*opeloggrpc.OpeLogEntry{}}
		m.les[relPath] = le
	}
	le.OpeLogEntries = append(le.OpeLogEntries, ole)
	return nil
}

var _ opelog.OpeLogManager = &m2fMng{}
