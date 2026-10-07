package opelogimpl

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type OplWalker interface {
	Run() error
}

type workerNotif struct {
	wkn     int
	isStart bool
}

type oplWalkerImpl struct {
	mx            sync.Mutex
	lgr           *slog.Logger
	conc          int
	oplq          opelog.Queue
	oplm          opelog.OpeLogManager
	owo           *config.OpeLogOptionsType
	sds           dssa.Dssa
	tds           dssa.Dssa
	sRoot         string
	tRoot         string
	inclRegs      []*regexp.Regexp
	exclRegs      []*regexp.Regexp
	toolStartTime int64
	session       string
	sessionTime   int64
	inventory     string
	invTime       int64
	gErrs         []error
	syncTicker    *time.Ticker
	wkNtfChan     chan workerNotif
	bg            context.Context
}

func (ow *oplWalkerImpl) owErr(lgr *slog.Logger, msg string, err error) error {
	lgr.Error(msg, "err", err)
	ow.mx.Lock()
	defer ow.mx.Unlock()
	ow.gErrs = append(ow.gErrs, fmt.Errorf("%s: %v", msg, err))
	return err
}

func (ow *oplWalkerImpl) detail(lgr *slog.Logger, msg string, args ...any) {
	lgr.Log(ow.bg, slog.LevelDebug+2, msg, args...)
}

func (ow *oplWalkerImpl) impliesGoal(goal string) bool {
	reqGoals := strings.Split(ow.owo.Goals, ",")
	hasAny := func(goals string) bool {
		sgs := strings.Split(goals, ",")
		for rg := range slices.Values(reqGoals) {
			if slices.Contains(sgs, rg) {
				return true
			}
		}
		return false
	}
	switch goal {
	case "load":
		return hasAny("load,create,update")
	case "create":
		return hasAny("create,update")
	case "update":
		return hasAny("update")
	case "verify":
		return hasAny("verify")
	default:
		return false
	}
}

func (ow *oplWalkerImpl) oplmSync() {
	lgr := ow.lgr.With("worker", "oplmSync")
	lgr.Debug("oplWalkerImpl", "start", true)

	for tick := range ow.syncTicker.C {
		lgr.Info("oplWalkerImpl", "tick", tick)
		if err := ow.oplm.Sync(); err != nil {
			ow.owErr(lgr, "failed to synchronize logs", err)
		}
	}
}

func (ow *oplWalkerImpl) startOrRestart() {
	ow.toolStartTime = time.Now().Unix()
	ow.oplq.Put("")
}

func (ow *oplWalkerImpl) workersController() {
	lgr := ow.lgr.With("worker", "workerController")
	stoppedWkNum := ow.conc
	lgr.Debug("oplWalkerImpl", "start", true)
	rsTo := ow.owo.ResetTimeout * int64(time.Second)
	hasTimeOut := true
	if rsTo == 0 {
		rsTo = int64(60 * time.Second)
		hasTimeOut = false
	}
	ticker := time.NewTicker(time.Duration(rsTo))

	for {
		select {
		case notif := <-ow.wkNtfChan:
			if notif.isStart {
				ticker.Reset(time.Duration(rsTo))
				stoppedWkNum--
			} else {
				stoppedWkNum++
			}
			lgr.Debug("oplWalkerImpl", "wkn", notif.wkn, "isStart", notif.isStart, "stoppedWkNum", stoppedWkNum)
		case <-ticker.C:
			ticker.Reset(time.Duration(rsTo))
			lgr.Debug("oplWalkerImpl", "hasTo", hasTimeOut, "stoppedWkNum", stoppedWkNum)
			if hasTimeOut && stoppedWkNum == ow.conc {
				if ow.owo.EnableRestart {
					ow.startOrRestart()
				} else {
					// will interrupt all workers
					ow.oplq.Close()
				}
			}
		}
	}
}

// newOplLogicalEntry initializes oplLogicalEntry state from loaded state in le
func (ow *oplWalkerImpl) newOplLogicalEntry(plgr *slog.Logger, relPath string, le *opelog.LogicalEntry) *oplLogicalEntry {
	hc := false
	if le == nil {
		le = opelog.NewLogicalEntry()
		hc = true
	}
	ole := &oplLogicalEntry{
		lgr:             plgr.With("relPath", relPath),
		hasChanges:      hc,
		relPath:         relPath,
		owi:             ow,
		le:              le,
		childrenLeCache: map[string]*opelog.LogicalEntry{},
	}
	ole.source = &oplStoredEntry{
		lgr: plgr.With("path", path.Join("{S}", relPath)),
		ole: ole}
	ole.target = &oplStoredEntry{
		lgr: plgr.With("path", path.Join("{T}", relPath)),
		ole: ole, isTarget: true}
	return ole
}

// getLogicalEntry ensures entry is loaded or newly created and retrieves related parent state if needed
//
// as an optimization, absent parent's state is propagated to child
func (ow *oplWalkerImpl) getLogicalEntry(lgr *slog.Logger, relPath string) (*oplLogicalEntry, error) {
	le, err := ow.oplm.GetLogicalEntry(relPath)
	if err != nil {
		ow.owErr(lgr, "oplWalkerImpl: GetLogicalEntry", err)
		return nil, err
	}
	ole := ow.newOplLogicalEntry(lgr, relPath, le)
	if relPath == "" {
		return ole, nil
	}
	pRelPath := common.ParentPath(relPath)
	ple, err := ow.oplm.GetLogicalEntry(pRelPath)
	if err != nil {
		ow.owErr(lgr, "oplWalkerImpl: GetLogicalEntry on parent initial", err)
		return nil, err
	}
	ole.parentLe = ple
	ole.parentSSt = ple.GetState(ow.sessionTime, false)
	ole.parentTSt = ple.GetState(ow.sessionTime, true)

	// absent parent's state is propagated directly to child
	pOle := ow.newOplLogicalEntry(lgr, pRelPath, ple)
	if pOle.source.isAbsent() && ole.source.getState() == nil {
		ole.source.setState(false, opelog.STC_DONE_ABSENT, "", nil, nil, 0)
		ole.hasChanges = true
	}
	if pOle.target.isAbsent() && ole.target.getState() == nil {
		ole.target.setState(false, opelog.STC_DONE_ABSENT, "", nil, nil, 0)
		ole.hasChanges = true
	}
	return ole, nil
}

func isChildInState(cName string, st *opelog.State) bool {
	if st == nil || st.Se == nil {
		return false
	}
	for _, child := range st.Se.Children {
		if child == cName {
			return true
		}
	}
	return false
}

// notifyParent merges source and target state after processing and notifies parent if last child
// from both branches
func (ow *oplWalkerImpl) notifyParent(lgr *slog.Logger, ole *oplLogicalEntry) error {
	if ole.parentLe == nil {
		return nil
	}
	// reload parent and locks to modify
	ow.mx.Lock()
	defer ow.mx.Unlock()
	pRelPath := common.ParentPath(ole.relPath)
	ple, err := ow.oplm.GetLogicalEntry(pRelPath)
	if err != nil {
		ow.owErr(lgr, "oplWalkerImpl: GetLogicalEntry on parent final", err)
		return err
	}
	cName := path.Base(ole.relPath)
	parentSSt := ple.GetState(ow.sessionTime, false)
	hasChanges := false
	if isChildInState(cName, parentSSt) {
		if parentSSt.DepCount <= 0 {
			err := fmt.Errorf("source child %s notifies twice parent", ole.relPath)
			ow.owErr(lgr, "oplWalkerImpl: internal", err)
			return err
		}
		lgr.Debug("parentSSt", "relPath", ole.relPath, "DepCount", parentSSt.DepCount)
		hasChanges = true
		parentSSt.DepCount--
		if parentSSt.DepCount == 0 {
			parentSSt.DepCount = -1
		}
	}
	parentTSt := ple.GetState(ow.sessionTime, true)
	if isChildInState(cName, parentTSt) {
		if parentTSt.DepCount <= 0 {
			err := fmt.Errorf("target child %s notifies twice parent", ole.relPath)
			ow.owErr(lgr, "oplWalkerImpl: internal", err)
			return err
		}
		hasChanges = true
		parentTSt.DepCount--
		if parentTSt.DepCount == 0 {
			parentTSt.DepCount = -1
		}
	}
	if hasChanges {
		if err := ow.oplm.PutLogicalEntry(pRelPath, ple); err != nil {
			ow.owErr(lgr, "oplWalkerImpl: notify parent: put ple", err)
			return err
		}
	}
	// both branches either just terminated or inactive
	if (parentSSt.DepCount == -1 && parentTSt.DepCount <= 0) ||
		(parentTSt.DepCount == -1 && parentSSt.DepCount <= 0) {
		// notifies parent
		if err := ow.oplq.Put(pRelPath); err != nil {
			ow.owErr(lgr, "oplWalkerImpl: process entry: put parent in queue", err)
			return err
		}
	}
	return nil
}

func (ow *oplWalkerImpl) processEntry(lgr *slog.Logger, wkn int, relPath string) {
	ow.wkNtfChan <- workerNotif{wkn: wkn, isStart: true}
	defer func() {
		ow.wkNtfChan <- workerNotif{wkn: wkn}
	}()
	ow.detail(lgr, "oplWalkerImpl", "readPathFromQueue", relPath)
	ole, err := ow.getLogicalEntry(lgr, relPath)
	if err != nil {
		return
	}

	if err := ole.process(); err != nil {
		_ = ow.oplm.PutLogicalEntry(relPath, ole.le)
		ow.owErr(lgr, "oplWalkerImpl: process entry: put err le", err)
		return
	}

	if ole.hasChanges {
		if err := ow.oplm.PutLogicalEntry(relPath, ole.le); err != nil {
			ow.owErr(lgr, "oplWalkerImpl: process entry: put le", err)
			return
		}
	}
	for _, child := range ole.childrenQueue() {
		if err := ow.oplq.Put(child); err != nil {
			ow.owErr(lgr, "oplWalkerImpl: process entry: put child in queue", err)
			return
		}
	}
	_ = ow.notifyParent(lgr, ole)
}

func (ow *oplWalkerImpl) work(wkn int, wg *sync.WaitGroup) {
	defer wg.Done()
	lgr := ow.lgr.With("worker", wkn)
	lgr.Debug("oplWalkerImpl: start")
	for {
		relPath, err := ow.oplq.Get()
		if err != nil {
			if err != common.ErrReadClosedQueue {
				ow.owErr(lgr, "oplWalkerImpl", err)
			}
			break
		}
		ow.processEntry(lgr.With("relPath", relPath), wkn, relPath)
	}
	lgr.Debug("oplWalkerImpl: stop")
}

func (ow *oplWalkerImpl) Run() error {
	ow.lgr.Info("oplWalkerImpl: Run")
	var err error
	ow.inclRegs, err = common.ReFromFile(ow.owo.InclListPath, "inclusion list")
	if err != nil {
		ow.lgr.Error("oplWalkerImpl", "err", err)
		return err
	}
	ow.exclRegs, err = common.ReFromFile(ow.owo.InclListPath, "inclusion list")
	if err != nil {
		ow.lgr.Error("oplWalkerImpl", "err", err)
		return err
	}
	sTs, iTs, err := ow.oplm.Open(ow.session, ow.inventory, false)
	if err != nil {
		ow.lgr.Error("oplWalkerImpl: Run: open logs", "err", err)
		return err
	}
	ow.sessionTime, ow.invTime = sTs, iTs
	ow.wkNtfChan = make(chan workerNotif)
	go ow.workersController()

	var wg sync.WaitGroup
	for wkn := range ow.conc {
		wg.Add(1)
		go ow.work(wkn, &wg)
	}
	if ow.owo.SyncPeriod != 0 {
		ow.syncTicker = time.NewTicker(time.Duration(ow.owo.SyncPeriod))
		go ow.oplmSync()
	}

	// start walker
	ow.startOrRestart()
	wg.Wait()
	if ow.owo.SyncPeriod != 0 {
		ow.syncTicker.Stop()
	}
	if err := ow.oplm.Close(); err != nil {
		ow.lgr.Error("oplWalkerImpl: Run: close logs", "err", err)
		return err
	}
	ow.lgr.Info("oplWalkerImpl: Run: end")
	if len(ow.gErrs) > 0 {
		err := fmt.Errorf("walker %d errors occured", len(ow.gErrs))
		ow.lgr.Error("oplWalkerImpl: Run:", "err", err)
		ow.detail(ow.lgr, "oplWalkerImpl: Run:", "err", err, "details", ow.gErrs)
		return err
	}
	return nil
}

func NewOplWalker(lgr *slog.Logger, conc int,
	oplq opelog.Queue, oplm opelog.OpeLogManager,
	owo *config.OpeLogOptionsType,
	sds, tds dssa.Dssa, sRoot, tRoot string,
	session, inventory string,
) OplWalker {
	if conc == 0 {
		conc = 1
	}
	if oplq == nil {
		oplq = NewMemQueue()
	}

	return &oplWalkerImpl{
		lgr:  lgr,
		conc: conc, oplq: oplq, oplm: oplm, owo: owo,
		sds: sds, tds: tds, sRoot: sRoot, tRoot: tRoot,
		session: session, inventory: inventory,
		bg: context.Background(),
	}
}
