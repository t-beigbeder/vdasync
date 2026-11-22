package opelogimpl

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opelog"
	"github.com/t-beigbeder/vdasync/opeloggrpc"
	"google.golang.org/protobuf/proto"
)

type pbOplMng struct {
	lgr      *slog.Logger
	path_    string
	bckLimit int
	mx       sync.Mutex
	om       *opelog.OpeLogMaster
	readOnly bool
	isOpen   bool
}

func (m *pbOplMng) saveMaster() error {
	if m.bckLimit > 0 {
		if err := common.LimitedBackup(localfiles.MakeLocalFilesDssa(), m.path_, m.bckLimit); err != nil {
			return err
		}
	}
	gom := opelog.OpeLogMaster2ProtoBuf(m.om)
	bs, err := proto.Marshal(gom)
	if err != nil {
		return err
	}
	if err = common.WriteFile(m.path_, bs); err != nil {
		return err
	}
	return nil
}

// Create implements [opelog.OpeLogManager].
func (m *pbOplMng) Create(source string, target string) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.path_ != "" && common.FileExists(m.path_) {
		return fmt.Errorf("pbOplMng.Create: %s already exists", m.path_)
	}
	if m.om != nil && len(m.om.Sessions) != 0 {
		return fmt.Errorf("m2fMng.Create: %s should be created without sessions", m.path_)
	}
	m.om = &opelog.OpeLogMaster{
		SourceRoot: source,
		TargetRoot: target,
		Sessions:   map[string]*opelog.SessionInfo{},
	}
	if err := m.saveMaster(); err != nil {
		return err
	}
	return nil
}

// AddSession implements [opelog.OpeLogManager].
func (m *pbOplMng) AddSession(label string, si *opelog.SessionInfo) (int64, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return 0, errors.New("pbOplMng.AddSession: not opened")
	}
	if m.readOnly {
		return 0, errors.New("pbOplMng.AddSession: opened in read-only")
	}
	_, ok := m.om.Sessions[label]
	if ok {
		return 0, fmt.Errorf("session %s already exists", label)
	}
	rt := m.om.NextTimeStamp()
	m.om.Sessions[label] = &opelog.SessionInfo{
		Comment:       si.Comment,
		ReferenceTime: rt,
		IsInventory:   si.IsInventory,
	}
	if err := m.saveMaster(); err != nil {
		return 0, err
	}
	return rt, nil
}

// Open implements [opelog.OpeLogManager].
func (m *pbOplMng) Open(readOnly bool) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if m.isOpen {
		return errors.New("pbOplMng.Open: already opened")
	}
	lock := fmt.Sprintf("%s.lock", m.path_)
	if !readOnly && common.FileExists(lock) {
		return fmt.Errorf("pbOplMng.Open: locked (%s)", lock)
	}
	bs, err := common.UnsafeLoadFile(m.path_)
	if err != nil {
		return err
	}
	var gom opeloggrpc.OpeLogMaster
	if err = proto.Unmarshal(bs, &gom); err != nil {
		return err
	}
	m.om = opelog.ProtoBuf2OpeLogMaster(&gom)

	if !readOnly {
		if err := common.WriteFile(lock, []byte{}); err != nil {
			return err
		}
	}
	m.isOpen = true
	m.readOnly = readOnly
	m.om.StoreType = "pbf"
	return nil
}

// GetOlm implements [opelog.OpeLogManager].
func (m *pbOplMng) GetOlm() *opelog.OpeLogMaster {
	return m.om.Clone()
}

// SetOlm implements [opelog.OpeLogManager].
func (m *pbOplMng) SetOlm(om *opelog.OpeLogMaster) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("pbOplMng.SetOlm: not opened")
	}
	if m.readOnly {
		return errors.New("pbOplMng.SetOlm: opened in read-only")
	}
	*m.om = *om.Clone()
	m.saveMaster()
	return nil
}

// Sync implements [opelog.OpeLogManager].
func (m *pbOplMng) Sync() error {
	return nil
}

// Close implements [opelog.OpeLogManager].
func (m *pbOplMng) Close() error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("pbOplMng.Close: not opened")
	}
	m.isOpen = false
	if !m.readOnly {
		if err := os.Remove(fmt.Sprintf("%s.lock", m.path_)); err != nil {
			return err
		}
	}
	return nil
}

func rp2Id(relPath string) string {
	return fmt.Sprintf("%064x", sha256.Sum256([]byte(relPath)))
}

// PutLogicalEntry implements [opelog.OpeLogManager].
func (m *pbOplMng) PutLogicalEntry(relPath string, ole *opelog.LogicalEntry) error {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return errors.New("pbOplMng.PutEntryLog: not opened")
	}
	if m.readOnly {
		return errors.New("pbOplMng.PutEntryLog: opened in read-only")
	}
	ole.EmbedPath = relPath
	gole := opelog.LogicalEntry2ProtoBuf(ole)
	bole, err := proto.Marshal(gole)
	if err != nil {
		return err
	}
	pbPath := path.Join(path.Dir(m.path_), common.Id2Path(rp2Id(relPath)))
	if err := os.MkdirAll(path.Dir(pbPath), 0750); err != nil {
		return err
	}
	hasBck := false
	if common.FileExists(pbPath) {
		hasBck = true
		if err := os.Rename(pbPath, pbPath+".bck"); err != nil {
			return err
		}
	}
	if err = common.WriteFile(pbPath, bole); err != nil {
		return err
	}
	if hasBck {
		_ = os.Remove(pbPath + ".bck")
	}
	return nil
}

func (m *pbOplMng) doGetLogicalEntry(path_ string) (*opelog.LogicalEntry, error) {
	bole, err := common.UnsafeLoadFile(path_)
	if err != nil {
		return nil, err
	}
	gole := opeloggrpc.LogicalEntry{}
	if err := proto.Unmarshal(bole, &gole); err != nil {
		return nil, err
	}
	return opelog.ProtoBuf2LogicalEntry(&gole), nil
}

// GetLogicalEntry implements [opelog.OpeLogManager].
func (m *pbOplMng) GetLogicalEntry(relPath string) (*opelog.LogicalEntry, error) {
	m.mx.Lock()
	defer m.mx.Unlock()
	if !m.isOpen {
		return nil, errors.New("pbOplMng.GetLogicalEntry: not opened")
	}
	ole, err := m.doGetLogicalEntry(path.Join(path.Dir(m.path_), common.Id2Path(rp2Id(relPath))))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	if ole.EmbedPath != relPath {
		return nil, fmt.Errorf("pbOplMng.GetLogicalEntry: inconsistent embedPath %s vs %s", ole.EmbedPath, relPath)
	}
	return ole, nil
}

// Walk implements [opelog.OpeLogManager].
func (m *pbOplMng) doWalk(doIt func(path_ string) error) error {
	filepath.Walk(path.Dir(m.path_), func(path_ string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(path.Base(path_), path.Base(m.path_)) {
			return nil
		}
		if strings.HasSuffix(path_, ".bck") {
			return nil
		}
		return doIt(path_)
	})
	return nil
}

// Walk implements [opelog.OpeLogManager].
func (m *pbOplMng) Walk(doIt func(relPath string, ole *opelog.LogicalEntry) error, sorted bool, reverse bool) error {
	if !m.isOpen {
		return errors.New("pbOplMng.Walk: not opened")
	}
	var hasErrors bool
	if !sorted {
		err := m.doWalk(func(path_ string) error {
			ole, err := m.doGetLogicalEntry(path_)
			if err != nil {
				hasErrors = true
				m.lgr.Error("pbOplMng.Walk: direct", "relPath", path_, "err", err)
				return nil
			}
			if err := doIt(ole.EmbedPath, ole); err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return err
		}
		if hasErrors {
			return errors.New("pbOplMng.Walk: errors encountered, please consult the logs")
		}
		return nil
	}
	rps := []string{}
	if err := m.doWalk(func(path_ string) error {
		ole, err := m.doGetLogicalEntry(path_)
		if err != nil {
			hasErrors = true
			m.lgr.Error("pbOplMng.Walk: phase1", "relPath", path_, "err", err)
			return nil
		}
		rps = append(rps, ole.EmbedPath)
		return nil
	}); err != nil {
		return err
	}
	slices.SortFunc(rps, func(a, b string) int {
		if reverse {
			return strings.Compare(b, a)
		}
		return strings.Compare(a, b)
	})
	for _, relPath := range rps {
		ole, err := m.doGetLogicalEntry(path.Join(path.Dir(m.path_), common.Id2Path(rp2Id(relPath))))
		if err != nil {
			hasErrors = true
			m.lgr.Error("pbOplMng.Walk: phase2", "relPath", relPath, "err", err)
			continue
		}
		if err := doIt(ole.EmbedPath, ole); err != nil {
			return err
		}
	}
	if hasErrors {
		return errors.New("pbOplMng.Walk: errors encountered, please consult the logs")
	}
	return nil
}

var _ opelog.OpeLogManager = &pbOplMng{}

func MakePbfManager(lgr *slog.Logger, path_ string, bckLimit int) (opelog.OpeLogManager, error) {
	return &pbOplMng{lgr: lgr, path_: path_, bckLimit: bckLimit}, nil
}
