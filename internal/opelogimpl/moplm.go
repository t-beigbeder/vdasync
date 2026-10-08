package opelogimpl

import (
	"errors"
	"fmt"
	"maps"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/opelog"
)

type memOplm struct {
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

func (m *memOplm) tsInUse(ts int64) (yes bool) {
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

func (m *memOplm) nextTimeStamp() int64 {
	for ts := time.Now().Unix(); ; ts += 1 {
		if !m.tsInUse(ts) {
			return ts
		}
	}
}

// AddInventory implements [opelog.OpeLogManager].
func (m *memOplm) AddInventory(label string) (int64, error) {
	_, ok := m.inventories[label]
	if ok {
		return 0, fmt.Errorf("inventory %s already exists", label)
	}
	m.inventories[label] = m.nextTimeStamp()
	return m.inventories[label], nil
}

// AddSession implements [opelog.OpeLogManager].
func (m *memOplm) AddSession(label string) (int64, error) {
	_, ok := m.inventories[label]
	if ok {
		return 0, fmt.Errorf("session %s already exists", label)
	}
	m.sessions[label] = m.nextTimeStamp()
	return m.sessions[label], nil
}

// Create implements [opelog.OpeLogManager].
func (m *memOplm) Create(session, inventory, source, target string) (int64, int64, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.les != nil {
		return 0, 0, fmt.Errorf("memOplm.Create: should be created without entries")
	}
	if len(m.sessions) != 0 {
		return 0, 0, fmt.Errorf("memOplm.Create: should be created without sessions")
	}
	if len(m.inventories) != 0 {
		return 0, 0, fmt.Errorf("memOplm.Create: should be created without inventories")
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
	return sTs, iTs, nil
}

// NewSession implements [opelog.OpeLogManager].
func (m *memOplm) Open(session, inventory string, readOnly bool) (int64, int64, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.isOpen {
		return 0, 0, errors.New("memOplm.Open: already opened")
	}
	if m.sessions == nil {
		m.sessions = map[string]int64{}
	}
	sTs, ok := m.sessions[session]
	if session != "" && !ok {
		return 0, 0, fmt.Errorf("unknown session: %s", session)
	}
	if m.inventories == nil {
		m.inventories = map[string]int64{}
	}
	iTs, ok := m.inventories[inventory]
	if inventory != "" && !ok {
		return 0, 0, fmt.Errorf("unknown inventory: %s", inventory)
	}
	m.les = make(map[string]*opelog.LogicalEntry, 0)
	m.isOpen = true
	m.readOnly = readOnly
	return sTs, iTs, nil
}

// Sync implements [opelog.OpeLogManager].
func (m *memOplm) Sync() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("memOplm.Sync: not opened")
	}
	if !m.hasUpdates {
		return nil
	}
	m.hasUpdates = false
	return nil
}

// Close implements [opelog.OpeLogManager].
func (m *memOplm) Close() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("memOplm.Close: not opened")
	}
	if !m.hasUpdates {
		m.isOpen = false
		return nil
	}
	m.hasUpdates = false
	m.isOpen = false
	return nil
}

// GetLogicalEntry implements [opelog.OpeLogManager].
func (m *memOplm) GetLogicalEntry(relPath string) (*opelog.LogicalEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return nil, errors.New("memOplm.GetLogicalEntry: not opened")
	}
	le, _ := m.les[relPath]
	return le, nil
}

// PutLogicalEntry implements [opelog.OpeLogManager].
func (m *memOplm) PutLogicalEntry(relPath string, ole *opelog.LogicalEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("memOplm.PutEntryLog: not opened")
	}
	if m.readOnly {
		return errors.New("memOplm.PutEntryLog: opened in read-only")
	}
	m.les[relPath] = ole
	m.hasUpdates = true
	return nil
}

// Walk implements [opelog.OpeLogManager].
func (m *memOplm) Walk(doIt func(relPath string, ole *opelog.LogicalEntry) error) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("memOplm.Walk: not opened")
	}
	for relPath, ole := range m.les {
		if err := doIt(relPath, ole); err != nil {
			return err
		}
	}
	return nil
}

var _ opelog.OpeLogManager = &memOplm{}

func MakeMemOplm() (opelog.OpeLogManager, error) {
	return &memOplm{}, nil
}
