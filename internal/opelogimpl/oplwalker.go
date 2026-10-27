package opelogimpl

import (
	"context"
	"fmt"
	"log/slog"
	"path"
	"regexp"
	"slices"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opelog"
)

type OplWalker interface {
	Run() error
	Shutdown(*slog.Logger) error
	ToolInfo() *opelog.ToolInformation
}

type workerNotif struct {
	wkn     int
	isStart bool
}

type oplWalkerImpl struct {
	mx             sync.Mutex
	lgr            *slog.Logger
	conc           int
	oplq           opelog.Queue
	oplm           opelog.OpeLogManager
	oplTi          *opelog.ToolInformation
	owo            *config.OpeLogOptionsType
	sds            dssa.Dssa
	tds            dssa.Dssa
	sRoot          string
	tRoot          string
	inclRegs       []*regexp.Regexp
	exclRegs       []*regexp.Regexp
	toolStartTime  int64
	restartCount   int
	session        string
	sessionTime    int64
	inventory      string
	invTime        int64
	eRate          float64
	errRtl         *common.CancelEvLter
	errRtlCFunc    func()
	errCanceling   bool
	gErrs          []error
	closedQKnown   bool
	wkNtfChan      chan workerNotif
	wkCtlStopChan  chan bool
	wkCtlDoneChan  chan bool
	shdownDoneChan chan bool
	bg             context.Context
	testCtx        string
	restartCb      func(testCtx string, count int) bool
}

func (ow *oplWalkerImpl) doOplmSync(lgr *slog.Logger) {
	ow.detail(lgr, "doOplmSync")
	if err := ow.oplm.Sync(); err != nil {
		ow.owErr(lgr, "failed to synchronize logs", err)
	}
}

func (ow *oplWalkerImpl) doOplmExport(lgr *slog.Logger) {
	ow.detail(lgr, "doOplmExport")
	if ow.owo.ExpBck > 0 {
		if err := common.LimitedBackup(localfiles.MakeLocalFilesDssa(), ow.owo.ExpFile, ow.owo.ExpBck); err != nil {
			ow.owErr(lgr, "failed to backup export logs", err)
		}
	}
	rpt := RPT_SYNTHETIC
	if ow.owo.ExpRpt == "full" {
		rpt = RPT_FULL
	} else if ow.owo.ExpRpt != "synth" {
		lgr.Warn("doOplmExport: unknown oplrpt", "rpt", ow.owo.ExpRpt)
	}
	if err := OplCsvExport(ow.lgr, ow.oplm, ow.owo.ExpFile, rpt, ow.inventory, ow.session, &RptConfig{ToolInfo: ow.oplTi}); err != nil {
		ow.owErr(lgr, "failed to export logs", err)
	}
}

func (ow *oplWalkerImpl) startOrRestart(restart bool) bool {
	if restart {
		_ = ow.dssEndSessions("startOrRestart")
		_ = ow.dssNewSessions("startOrRestart")
		ow.restartCount++
		if ow.restartCb == nil || ow.restartCb(ow.testCtx, ow.restartCount) {
			ow.detail(ow.lgr, "oplWalkerImpl: startOrRestart", "count", ow.restartCount, "restart", true)
		} else {
			ow.detail(ow.lgr, "oplWalkerImpl: startOrRestart", "count", ow.restartCount, "restart", false)
			return false
		}
	}
	time.Sleep(time.Millisecond)
	ow.toolStartTime = time.Now().UnixMilli()
	ow.oplq.Put("")
	return true
}

func (ow *oplWalkerImpl) workersController() {
	lgr := ow.lgr.With("worker", "workerController")
	stoppedWkNum := ow.conc
	lgr.Debug("oplWalkerImpl", "start", true)
	rsTo := ow.owo.ResetMilTimeout * int64(time.Millisecond)
	hasTimeOut := true
	if rsTo == 0 {
		rsTo = int64(60 * time.Second)
		hasTimeOut = false
	}
	ticker := time.NewTicker(time.Duration(rsTo))
	for stopped := false; !stopped; {
		select {
		case <-ow.wkCtlStopChan:
			lgr.Debug("oplWalkerImpl", "wkCtlStopChan", "received")
			stopped = true
			ticker.Stop()
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
				if !ow.owo.EnableRestart || !ow.startOrRestart(true) {
					// will interrupt all workers
					ow.oplq.Close()
				}
			}
		}
	}
	lgr.Debug("oplWalkerImpl", "wkCtlDoneChan", "sending")
	ow.wkCtlDoneChan <- true
}

// notifyParent merges source and target state after processing and notifies parent if last child
// from both branches
func (ow *oplWalkerImpl) notifyParent(ole *oplLogicalEntry) error {
	var ratedErr error
	// reload parent and locks to modify
	ow.mx.Lock()
	defer func() {
		ow.mx.Unlock()
		if ratedErr != nil && ow.errRtl != nil {
			if err := ow.errRtl.RtlThis(); err != nil {
				ow.lgr.Debug("notifyParent", "err", err)
			}
		}
	}()
	pRelPath := common.ParentPath(ole.relPath)
	ple, err := ow.oplm.GetLogicalEntry(pRelPath)
	if err != nil {
		ratedErr = ow.lockedOwErr(ole.lgr, "oplWalkerImpl: GetLogicalEntry on parent final", err)
		return err
	}
	cName := path.Base(ole.relPath)
	parentSSt := ple.GetState(ow.sessionTime, false)
	parentTSt := ple.GetState(ow.sessionTime, true)

	hasChanges := false
	if isChildInState(cName, parentSSt) {
		ole.lgr.Debug("notifyParent", "parentSSt.DepCount", parentSSt.DepCount)
		if parentSSt.DepCount <= 0 {
			err := fmt.Errorf("source child %s notifies twice parent", ole.relPath)
			ratedErr = ow.lockedOwErr(ole.lgr, "oplWalkerImpl: internal", err)
			return err
		}
		hasChanges = true
		parentSSt.DepCount--
		if parentSSt.DepCount == 0 {
			parentSSt.DepCount = -1
		}
	}
	if isChildInState(cName, parentTSt) {
		if parentTSt.DepCount <= 0 {
			err := fmt.Errorf("target child %s notifies twice parent", ole.relPath)
			ratedErr = ow.lockedOwErr(ole.lgr, "oplWalkerImpl: internal", err)
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
			ratedErr = ow.lockedOwErr(ole.lgr, "oplWalkerImpl: put parent le", err)
			return err
		}
	}

	// if one branche just terminated and other one too or inactive
	if (parentSSt.DepCount == -1 && parentTSt.DepCount <= 0) ||
		(parentTSt.DepCount == -1 && parentSSt.DepCount <= 0) {

		// notifies parent
		if err := ow.oplq.Put(pRelPath); err != nil {
			ratedErr = ow.lockedOwErr(ole.lgr, "oplWalkerImpl: put parent in queue", err)
			return err
		}
	}
	return nil
}

// processEntry is the entry point for processing each DSS entry, source and/or target
func (ow *oplWalkerImpl) processEntry(lgr *slog.Logger, wkn int, relPath string) {
	ow.wkNtfChan <- workerNotif{wkn: wkn, isStart: true}
	defer func() {
		ow.wkNtfChan <- workerNotif{wkn: wkn}
	}()
	ow.detail(lgr, "oplWalkerImpl", "readPathFromQueue", relPath)
	if slices.Contains(ow.owo.SpecialInterest, relPath) {
		if ow.owo.SpeIntPhase == ow.testCtx || ow.owo.SpeIntPhase == "" {
			lgr.Debug("processEntry: SpecialInterest", "sRoot", ow.sRoot, "tRoot", ow.tRoot)
			lgr.Info("processEntry: SpecialInterest")
		}
	}
	ole, err := ow.getLogicalEntry(lgr, relPath)
	if err != nil {
		return
	}

	if err := ole.process(); err != nil {
		_ = ow.oplm.PutLogicalEntry(relPath, ole.le)
		ow.owErr(ole.lgr, "oplWalkerImpl: process entry", err)
		return
	}

	children := ole.setChildrenQueue() // side effect hasChanges
	if ole.hasChanges {
		if err = ole.propagateChildrenStats(); err != nil {
			ow.owErr(ole.lgr, "oplWalkerImpl: children stats propagation:", err)
		}
		if sst := ole.source.getState(); sst != nil {
			sst.ToolStartTime = ow.toolStartTime
		}
		if tst := ole.target.getState(); tst != nil {
			tst.ToolStartTime = ow.toolStartTime
		}
		if err := ow.oplm.PutLogicalEntry(relPath, ole.le); err != nil {
			ow.owErr(ole.lgr, "oplWalkerImpl: put le", err)
			return
		}
	}
	for _, child := range children {
		if err := ow.oplq.Put(path.Join(relPath, child)); err != nil {
			ow.owErr(ole.lgr, "oplWalkerImpl: put child in queue", err)
			return
		}
	}
	if len(children) != 0 || relPath == "" {
		return
	}
	if err = ow.notifyParent(ole); err != nil {
		ow.owErr(ole.lgr, "oplWalkerImpl: parent notification:", err)
	}
}

// work is the lifetime activity of each worker
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
		ow.processEntry(lgr, wkn, relPath)
	}
	lgr.Debug("oplWalkerImpl: stop")
}

func (ow *oplWalkerImpl) Run() error {
	var (
		err       error
		syncJob   *common.PeriodicJob
		exportJob *common.PeriodicJob
	)
	defer close(ow.shdownDoneChan)

	ow.lgr.Info("oplWalkerImpl: Run")
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
	if err = ow.oplm.Open(false); err != nil {
		return err
	}
	defer ow.oplm.Close()
	if err = ow.checkConsistency(); err != nil {
		return err
	}
	ow.wkNtfChan = make(chan workerNotif)
	go ow.workersController()

	var wg sync.WaitGroup
	for wkn := range ow.conc {
		wg.Add(1)
		go ow.work(wkn, &wg)
	}
	if ow.owo.SyncPeriod != 0 {
		lgr := ow.lgr.With("syncJob", ow.owo.SyncPeriod)
		syncJob = common.NewPeriodicJob(lgr, ow.owo.SyncPeriod, func() { ow.doOplmSync(lgr) })
		go syncJob.Start()
	}
	if ow.owo.ExpFile != "" && ow.owo.ExpPeriod != 0 {
		lgr := ow.lgr.With("exportJob", ow.owo.ExpPeriod)
		exportJob = common.NewPeriodicJob(lgr, ow.owo.ExpPeriod, func() { ow.doOplmExport(lgr) })
		go exportJob.Start()
	}

	// start walker
	if err := ow.logInitToolInfo(); err != nil {
		ow.lgr.Error("oplWalkerImpl: Run: log initial tool info", "err", err)
		return err
	}
	ow.startOrRestart(false)

	// wait workers done
	wg.Wait()
	// close dss
	_ = ow.dssEndSessions("Run")
	// stopping controller
	ow.lgr.Debug("oplWalkerImpl", "wkCtlStopChan", "sending")
	ow.wkCtlStopChan <- true
	ow.lgr.Debug("oplWalkerImpl", "wkCtlDoneChan", "receiving")
	<-ow.wkCtlDoneChan

	// log final tool info
	if err := ow.logFinalToolInfo(); err != nil {
		ow.lgr.Error("oplWalkerImpl: Run: log final tool info", "err", err)
	}

	// stop periodic jobs
	if syncJob != nil {
		syncJob.Stop()
		<-syncJob.Done()
	}
	if exportJob != nil {
		exportJob.Stop()
		<-exportJob.Done()
	}

	// all done
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

func (ow *oplWalkerImpl) Shutdown(lgr *slog.Logger) error {
	ow.lgr.Info("oplWalkerImpl: Shutdown requested")
	if ow.errRtl != nil {
		ow.detail(ow.lgr, "oplWalkerImpl: canceling error limits")
		ow.errCanceling = true
		ow.errRtlCFunc()
	}
	ow.mx.Lock()
	lgr.Info("oplWalkerImpl Shutdown in progress...")
	if err := ow.oplq.Close(); err != nil {
		ow.mx.Unlock()
		return err
	}
	ow.mx.Unlock()
	ow.lgr.Debug("oplWalkerImpl: waiting for Run termination")
	<-ow.shdownDoneChan
	ow.lgr.Debug("oplWalkerImpl: Run terminated")
	lgr.Info("oplWalkerImpl Shutdown done")
	return nil
}

// ToolInfo implements [OplWalker].
func (ow *oplWalkerImpl) ToolInfo() *opelog.ToolInformation {
	return ow.oplTi
}

func NewOplWalker(lgr *slog.Logger, conc int,
	oplq opelog.Queue, oplm opelog.OpeLogManager,
	oplTi *opelog.ToolInformation, owo *config.OpeLogOptionsType,
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
		conc: conc, oplq: oplq, oplm: oplm,
		owo: owo, oplTi: oplTi,
		sds: sds, tds: tds, sRoot: sRoot, tRoot: tRoot,
		session: session, inventory: inventory,
		wkCtlStopChan:  make(chan bool),
		wkCtlDoneChan:  make(chan bool),
		shdownDoneChan: make(chan bool),
		bg:             context.Background(),
	}
}
