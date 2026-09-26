package opelogimpl

import (
	"errors"
	"fmt"
	"maps"
	"os"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
	"google.golang.org/protobuf/proto"
)

type m2fMng struct {
	path        string
	mx          sync.Mutex
	source      string
	target      string
	les         map[string]*opelog.LogicalEntry
	sessions    map[string]int64
	inventories map[string]int64
	readOnly    bool
	isOpen      bool
	hasUpdates  bool
}

func (m *m2fMng) save() error {
	aio := opeloggrpc.OpeLogAllInOne{
		SourceRoot:     m.source,
		TargetRoot:     m.target,
		LogicalEntries: make(map[string]*opeloggrpc.LogicalEntry, len(m.les)),
		Sessions:       maps.Clone(m.sessions),
		Inventories:    maps.Clone(m.inventories),
	}
	for rp, le := range m.les {
		aio.LogicalEntries[rp] = opelog.LogicalEntry2ProtoBuf(le)
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

// GetLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fMng) GetLogicalEntry(relPath string) (*opelog.LogicalEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return nil, errors.New("m2fMng.GetLogicalEntry: not opened")
	}
	le, _ := m.les[relPath]
	return le, nil
}

// Sync implements [opelog.OpeLogManager].
func (m *m2fMng) Sync() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Sync: not opened")
	}
	if !m.hasUpdates {
		return nil
	}
	if err := m.save(); err != nil {
		return err
	}
	m.hasUpdates = false
	return nil
}

// Close implements [opelog.OpeLogManager].
func (m *m2fMng) Close() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Close: not opened")
	}
	if !m.hasUpdates {
		m.isOpen = false
		return nil
	}
	if err := m.save(); err != nil {
		return err
	}
	m.hasUpdates = false
	m.isOpen = false
	if !m.readOnly {
		if err := os.Remove(fmt.Sprintf("%s.lock", m.path)); err != nil {
			return err
		}
	}
	return nil
}

// Create implements [opelog.OpeLogManager].
func (m *m2fMng) Create(session, inventory, source, target string) (int64, int64, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if common.FileExists(m.path) {
		return 0, 0, fmt.Errorf("m2fMng.Create: %s already exists", m.path)
	}
	if m.les != nil {
		return 0, 0, fmt.Errorf("m2fMng.Create: %s should be created without entries", m.path)
	}
	if len(m.sessions) != 0 {
		return 0, 0, fmt.Errorf("m2fMng.Create: %s should be created without sessions", m.path)
	}
	if len(m.inventories) != 0 {
		return 0, 0, fmt.Errorf("m2fMng.Create: %s should be created without inventories", m.path)
	}
	var (
		sTs, iTs int64
	)
	if session != "" {
		sTs = m.nextTimeStamp()
		m.sessions = map[string]int64{session: sTs}
	}
	if inventory != "" {
		iTs = m.nextTimeStamp()
		m.inventories = map[string]int64{inventory: iTs}
	}
	aio := opeloggrpc.OpeLogAllInOne{
		SourceRoot:  source,
		TargetRoot:  target,
		Sessions:    maps.Clone(m.sessions),
		Inventories: maps.Clone(m.inventories),
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return 0, 0, err
	}
	if err = common.WriteFile(m.path, bs); err != nil {
		return 0, 0, err
	}
	return sTs, iTs, nil
}

func (m *m2fMng) tsInUse(ts int64) (yes bool) {
	for sTs := range maps.Values(m.sessions) {
		if sTs == ts {
			return true
		}
	}
	for iTs := range maps.Values(m.inventories) {
		if iTs == ts {
			return true
		}
	}
	return
}

func (m *m2fMng) nextTimeStamp() int64 {
	for ts := time.Now().Unix(); ; ts += 1 {
		if !m.tsInUse(ts) {
			return ts
		}
	}
}

// AddInventory implements [opelog.OpeLogManager].
func (m *m2fMng) AddInventory(label string) (int64, error) {
	_, ok := m.inventories[label]
	if ok {
		return 0, fmt.Errorf("inventory %s already exists", label)
	}
	m.inventories[label] = m.nextTimeStamp()
	return m.inventories[label], nil
}

// AddSession implements [opelog.OpeLogManager].
func (m *m2fMng) AddSession(label string) (int64, error) {
	_, ok := m.inventories[label]
	if ok {
		return 0, fmt.Errorf("session %s already exists", label)
	}
	m.sessions[label] = m.nextTimeStamp()
	return m.sessions[label], nil
}

// NewSession implements [opelog.OpeLogManager].
func (m *m2fMng) Open(session, inventory string, readOnly bool) (int64, int64, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	lock := fmt.Sprintf("%s.lock", m.path)
	if m.isOpen {
		return 0, 0, errors.New("m2fMng.Open: already opened")
	}
	if !readOnly && common.FileExists(lock) {
		return 0, 0, fmt.Errorf("m2fMng.Open: locked (%s)", lock)
	}
	bs, err := common.UnsafeLoadFile(m.path)
	if err != nil {
		return 0, 0, err
	}
	var aio opeloggrpc.OpeLogAllInOne
	if err = proto.Unmarshal(bs, &aio); err != nil {
		return 0, 0, err
	}
	m.source = aio.SourceRoot
	m.target = aio.TargetRoot
	m.sessions = maps.Clone(aio.Sessions)
	if m.sessions == nil {
		m.sessions = map[string]int64{}
	}
	sTs, ok := m.sessions[session]
	if session != "" && !ok {
		return 0, 0, fmt.Errorf("unknown session: %s", session)
	}
	m.inventories = maps.Clone(aio.Inventories)
	if m.inventories == nil {
		m.inventories = map[string]int64{}
	}
	iTs, ok := m.inventories[inventory]
	if inventory != "" && !ok {
		return 0, 0, fmt.Errorf("unknown inventory: %s", inventory)
	}
	m.les = make(map[string]*opelog.LogicalEntry, len(aio.LogicalEntries))
	for rp, gle := range aio.LogicalEntries {
		m.les[rp] = opelog.ProtoBuf2LogicalEntry(gle)
	}
	if !readOnly {
		if err := common.WriteFile(lock, []byte{}); err != nil {
			return 0, 0, err
		}
	}
	m.isOpen = true
	m.readOnly = readOnly
	return sTs, iTs, nil
}

// PutLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fMng) PutLogicalEntry(relPath string, ole *opelog.LogicalEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.PutEntryLog: not opened")
	}
	if m.readOnly {
		return errors.New("m2fMng.PutEntryLog: opened in read-only")
	}
	m.les[relPath] = ole
	m.hasUpdates = true
	return nil
}

// Walk implements [opelog.OpeLogManager].
func (m *m2fMng) Walk(doIt func(relPath string, ole *opelog.LogicalEntry) error) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Walk: not opened")
	}
	for relPath, ole := range m.les {
		if err := doIt(relPath, ole); err != nil {
			return err
		}
	}
	return nil
}

var _ opelog.OpeLogManager = &m2fMng{}

func MakeM2fManager(path string) (opelog.OpeLogManager, error) {
	return &m2fMng{path: path}, nil
}
