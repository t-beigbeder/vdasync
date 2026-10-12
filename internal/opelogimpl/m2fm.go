package opelogimpl

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
	"google.golang.org/protobuf/proto"
)

type m2fOplMng struct {
	pathOpt    string
	bckLimit   int
	mx         sync.Mutex
	om         *opelog.OpeLogMaster
	les        map[string]*opelog.LogicalEntry
	readOnly   bool
	isOpen     bool
	hasUpdates bool
	testMarsh  bool
	marshLes   map[string][]byte
}

func (m *m2fOplMng) tsInUse(ts int64) (yes bool) {
	for _, sTs := range m.om.Sessions {
		if sTs.ReferenceTime == ts {
			return true
		}
	}
	return
}

func (m *m2fOplMng) nextTimeStamp() int64 {
	for ts := time.Now().Unix(); ; ts += 1 {
		if !m.tsInUse(ts) {
			return ts
		}
	}
}

func (m *m2fOplMng) save() error {
	if m.pathOpt == "" {
		return nil
	}
	aio := opeloggrpc.OpeLogAllInOne{
		Master:         opelog.OpeLogMaster2ProtoBuf(m.om),
		LogicalEntries: make(map[string]*opeloggrpc.LogicalEntry, len(m.les)),
	}
	for rp, le := range m.les {
		if m.testMarsh {
			ble := m.marshLes[rp]
			le2 := opeloggrpc.LogicalEntry{}
			if err := proto.Unmarshal(ble, &le2); err != nil {
				return err
			}
			aio.LogicalEntries[rp] = &le2
		} else {
			aio.LogicalEntries[rp] = opelog.LogicalEntry2ProtoBuf(le)
		}
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.pathOpt, bs); err != nil {
		return err
	}
	return nil
}

// AddSession implements [opelog.OpeLogManager].
func (m *m2fOplMng) AddSession(label string, si *opelog.SessionInfo) (int64, error) {
	_, ok := m.om.Sessions[label]
	if ok {
		return 0, fmt.Errorf("session %s already exists", label)
	}
	rt := m.nextTimeStamp()
	m.om.Sessions[label] = &opelog.SessionInfo{
		Comment:       si.Comment,
		ReferenceTime: rt,
		IsInventory:   si.IsInventory,
	}
	m.hasUpdates = true
	return rt, nil
}

// Create implements [opelog.OpeLogManager].
func (m *m2fOplMng) Create(source, target string) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.pathOpt != "" && common.FileExists(m.pathOpt) {
		return fmt.Errorf("m2fMng.Create: %s already exists", m.pathOpt)
	}
	if m.les != nil {
		return fmt.Errorf("m2fMng.Create: %s should be created without entries", m.pathOpt)
	}
	if m.om != nil && len(m.om.Sessions) != 0 {
		return fmt.Errorf("m2fMng.Create: %s should be created without sessions", m.pathOpt)
	}
	om := &opelog.OpeLogMaster{
		SourceRoot: source,
		TargetRoot: target,
		Sessions:   map[string]*opelog.SessionInfo{},
	}
	if m.pathOpt == "" {
		m.om = om
		m.les = map[string]*opelog.LogicalEntry{}
		m.marshLes = map[string][]byte{}
		return nil
	}
	aio := opeloggrpc.OpeLogAllInOne{
		Master:         opelog.OpeLogMaster2ProtoBuf(om),
		LogicalEntries: nil,
	}
	bs, err := proto.Marshal(&aio)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.pathOpt, bs); err != nil {
		return err
	}
	return nil
}

// Open implements [opelog.OpeLogManager].
func (m *m2fOplMng) Open(readOnly bool) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.isOpen {
		return errors.New("m2fMng.Open: already opened")
	}
	if m.pathOpt == "" {
		if m.om == nil {
			return errors.New("m2fMng.Open: memory-only m2f master has not been created")
		}
		m.isOpen = true
		m.readOnly = readOnly
		return nil
	}
	lock := fmt.Sprintf("%s.lock", m.pathOpt)
	if !readOnly && common.FileExists(lock) {
		return fmt.Errorf("m2fMng.Open: locked (%s)", lock)
	}
	bs, err := common.UnsafeLoadFile(m.pathOpt)
	if err != nil {
		return err
	}
	var aio opeloggrpc.OpeLogAllInOne
	if err = proto.Unmarshal(bs, &aio); err != nil {
		return err
	}
	m.om = opelog.ProtoBuf2OpeLogMaster(aio.Master)
	m.les = make(map[string]*opelog.LogicalEntry, len(aio.LogicalEntries))
	if m.testMarsh {
		m.marshLes = make(map[string][]byte, len(aio.LogicalEntries))
	}
	for rp, gle := range aio.LogicalEntries {
		ole := opelog.ProtoBuf2LogicalEntry(gle)
		if m.testMarsh {
			gle = opelog.LogicalEntry2ProtoBuf(ole)
			bs, err := proto.Marshal(gle)
			if err != nil {
				return err
			}
			m.marshLes[rp] = bs
			m.les[rp] = nil
		} else {
			m.les[rp] = ole
		}
	}
	if !readOnly {
		if err := common.WriteFile(lock, []byte{}); err != nil {
			return err
		}
	}
	m.isOpen = true
	m.readOnly = readOnly
	return nil
}

// Olm implements [opelog.OpeLogManager].
func (m *m2fOplMng) Olm() *opelog.OpeLogMaster {
	return m.om.Clone()
}

// Sync implements [opelog.OpeLogManager].
func (m *m2fOplMng) Sync() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Sync: not opened")
	}
	if !m.hasUpdates {
		return nil
	}
	if m.pathOpt == "" {
		return nil
	}
	if m.bckLimit > 0 {
		if err := common.LimitedBackup(localfiles.MakeLocalFilesDssa(), m.pathOpt, m.bckLimit); err != nil {
			return err
		}
	}
	if err := m.save(); err != nil {
		return err
	}
	m.hasUpdates = false
	return nil
}

// Close implements [opelog.OpeLogManager].
func (m *m2fOplMng) Close() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Close: not opened")
	}
	if m.hasUpdates {
		if err := m.save(); err != nil {
			return err
		}
	}
	m.hasUpdates = false
	m.isOpen = false
	if !m.readOnly && m.pathOpt != "" {
		if err := os.Remove(fmt.Sprintf("%s.lock", m.pathOpt)); err != nil {
			return err
		}
	}
	return nil
}

// GetLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fOplMng) GetLogicalEntry(relPath string) (*opelog.LogicalEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return nil, errors.New("m2fMng.GetLogicalEntry: not opened")
	}
	var ole *opelog.LogicalEntry
	if m.testMarsh {
		bole, _ := m.marshLes[relPath]
		gole := opeloggrpc.LogicalEntry{}
		if err := proto.Unmarshal(bole, &gole); err != nil {
			return nil, err
		}
		ole = opelog.ProtoBuf2LogicalEntry(&gole)
	} else {
		ole, _ = m.les[relPath]
	}
	return ole, nil
}

// PutLogicalEntry implements [opelog.OpeLogManager].
func (m *m2fOplMng) PutLogicalEntry(relPath string, ole *opelog.LogicalEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.PutEntryLog: not opened")
	}
	if m.readOnly {
		return errors.New("m2fMng.PutEntryLog: opened in read-only")
	}
	if m.testMarsh {
		gole := opelog.LogicalEntry2ProtoBuf(ole)
		bole, err := proto.Marshal(gole)
		if err != nil {
			return err
		}
		m.marshLes[relPath] = bole
		m.les[relPath] = nil
	} else {
		m.les[relPath] = ole
	}
	m.hasUpdates = true
	return nil
}

// Walk implements [opelog.OpeLogManager].
func (m *m2fOplMng) Walk(doIt func(relPath string, ole *opelog.LogicalEntry) error) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("m2fMng.Walk: not opened")
	}
	for relPath, ole := range m.les {
		if m.testMarsh {
			bole, _ := m.marshLes[relPath]
			gole := opeloggrpc.LogicalEntry{}
			if err := proto.Unmarshal(bole, &gole); err != nil {
				return err
			}
			ole = opelog.ProtoBuf2LogicalEntry(&gole)
		}
		if err := doIt(relPath, ole); err != nil {
			return err
		}
	}
	return nil
}

var _ opelog.OpeLogManager = &m2fOplMng{}

func MakeM2fManager(pathOpt string, bckLimit int) (opelog.OpeLogManager, error) {
	return &m2fOplMng{pathOpt: pathOpt, bckLimit: bckLimit}, nil
}
