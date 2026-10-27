package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func (ow *oplWalkerImpl) needInvCheck() bool {
	return ow.owo.Check && !ow.owo.NoInvCheck && ow.inventory != ""
}

func (ow *oplWalkerImpl) lockedOwErr(lgr *slog.Logger, msg string, err error) error {
	sErr := err.Error()
	if strings.Contains(sErr, "closed queue") {
		if ow.closedQKnown {
			return err
		}
		ow.closedQKnown = true
	}
	lgr.Error(msg, "err", err)
	ow.gErrs = append(ow.gErrs, fmt.Errorf("%s: %v", msg, err))
	return err
}

func (ow *oplWalkerImpl) owErr(lgr *slog.Logger, msg string, err error) error {
	ow.mx.Lock()
	defer func() {
		ow.mx.Unlock()
		if ow.errRtl != nil {
			if err := ow.errRtl.RtlThis(); err != nil {
				ow.lgr.Debug("owErr", "err", err)
			}
		}
	}()
	return ow.lockedOwErr(lgr, msg, err)
}

func (ow *oplWalkerImpl) detail(lgr *slog.Logger, msg string, args ...any) {
	lgr.Log(ow.bg, slog.LevelDebug+2, msg, args...)
}

func (ow *oplWalkerImpl) impliesGoal(goal string) bool {
	reqGoals := strings.Split(ow.owo.Goal, ",")
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
	default:
		return false
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

// getLogicalEntry ensures entry is loaded or newly created and retrieves related parent state if applies
//
// as an optimization, absent parent's state is propagated to (current) child
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

	// absent parent's state is propagated directly to new child
	pOle := ow.newOplLogicalEntry(lgr, pRelPath, ple)
	if pOle.source.isAbsent() && ole.source.getState() == nil {
		ole.source.setState(opelog.STC_DONE_ABSENT, "", nil, nil, 0)
		ole.hasChanges = true
	}
	if pOle.target.isAbsent() && ole.target.getState() == nil {
		ole.target.setState(opelog.STC_DONE_ABSENT, "", nil, nil, 0)
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

func (ow *oplWalkerImpl) setupErtl() (err error) {
	if ow.owo.ErrorRate == "" {
		return
	}
	if ow.eRate, err = strconv.ParseFloat(ow.owo.ErrorRate, 64); err != nil {
		err = fmt.Errorf("error rate: %s", err.Error())
		return
	}
	ow.errRtl, ow.errRtlCFunc = common.NewCancelEvLter(ow.eRate, func(doing chan any, d time.Duration) {
		ow.lgr.With("ertl", true).Info("error rate limit raised", "delay", d)
	})
	return nil
}

func (ow *oplWalkerImpl) checkConsistency() (err error) {
	defer func() {
		if err != nil {
			ow.lgr.Error("oplWalkerImpl: Run: open logs", "err", err)
		}
	}()

	if err = ow.setupErtl(); err != nil {
		return
	}
	if ow.sRoot != ow.oplm.GetOlm().SourceRoot {
		err = fmt.Errorf("source root in opelog path differs from %s", ow.sRoot)
		return
	}
	if ow.tRoot != ow.oplm.GetOlm().TargetRoot {
		err = fmt.Errorf("target root in opelog path differs from %s", ow.tRoot)
		return
	}
	if ow.session != "" {
		sSi := ow.oplm.GetOlm().Session(ow.session)
		if sSi == nil {
			err = fmt.Errorf("no such session %s", ow.session)
			return
		}
		if sSi.IsInventory {
			err = fmt.Errorf("cannot use inventory %s as a synchronization session", ow.session)
			return
		}
		ow.sessionTime = sSi.ReferenceTime
	} else if ow.oplm.GetOlm().StoreType == "mem" {
		ow.session = "auto"
		if ow.sessionTime, err = ow.oplm.AddSession("auto", &opelog.SessionInfo{}); err != nil {
			return
		}
	} else {
		err = errors.New("no session provided")
		return
	}
	if ow.inventory != "" {
		iSi := ow.oplm.GetOlm().Session(ow.inventory)
		if iSi == nil {
			err = fmt.Errorf("no such inventory %s", ow.inventory)
			return
		}
		ow.invTime = iSi.ReferenceTime
	}
	if ow.owo.Check && !ow.owo.NoInvCheck && ow.inventory == "" {
		err = errors.New("an inventory must be provided as check enabled")
		return
	}
	return
}

func (ow *oplWalkerImpl) logInitToolInfo() error {
	ow.oplTi.StartTime = time.Now().Unix()
	hn, _ := os.Hostname()
	ow.oplTi.HostName = hn
	ow.oplTi.SourceDssInfo = fmt.Sprintf("%s root %s", ow.sds.DssInfo(), ow.sRoot)
	ow.oplTi.TargetDssInfo = fmt.Sprintf("%s root %s", ow.tds.DssInfo(), ow.tRoot)
	om := ow.oplm.GetOlm()
	*om.Sessions[ow.session].ToolInfo = *ow.oplTi
	if err := ow.oplm.SetOlm(om); err != nil {
		return err
	}
	if err := ow.oplm.Sync(); err != nil {
		return err
	}
	return nil
}

func (ow *oplWalkerImpl) logFinalToolInfo() error {
	ow.oplTi.StopTime = time.Now().Unix()
	if le, _ := ow.oplm.GetLogicalEntry(""); le != nil {
		if stats := le.GetStats(ow.sessionTime); stats != nil {
			ow.oplTi.ExecSummary = stats.Display("summary")
		} else {
			ow.oplTi.ExecSummary = fmt.Sprintf("no stats on session %s", ow.session)
		}
	} else {
		ow.oplTi.ExecSummary = "no root entry in operation log"
	}
	om := ow.oplm.GetOlm()
	*om.Sessions[ow.session].ToolInfo = *ow.oplTi
	if err := ow.oplm.SetOlm(om); err != nil {
		return err
	}
	if err := ow.oplm.Sync(); err != nil {
		return err
	}
	return nil
}

func (ow *oplWalkerImpl) dssNewSessions(origin string) error {
	ow.lgr.Debug("dssNewSessions: source", "dss", ow.sds.DssInfo())
	if err := ow.sds.NewSession(); err != nil {
		msg := fmt.Sprintf("oplWalkerImpl: %s: NewSession source DSS", origin)
		ow.lgr.Error(msg, "err", err)
	}
	ow.lgr.Debug("dssNewSessions: target", "dss", ow.tds.DssInfo())
	if err := ow.tds.NewSession(); err != nil {
		msg := fmt.Sprintf("oplWalkerImpl: %s: NewSession target DSS", origin)
		ow.lgr.Error(msg, "err", err)
	}
	return nil
}

func (ow *oplWalkerImpl) dssEndSessions(origin string) error {
	ow.lgr.Debug("dssEndSessions: source", "dss", ow.sds.DssInfo())
	if err := ow.sds.EndSession(); err != nil {
		msg := fmt.Sprintf("oplWalkerImpl: %s: EndSession source DSS", origin)
		ow.lgr.Error(msg, "err", err)
	}
	ow.lgr.Debug("dssEndSessions: target", "dss", ow.tds.DssInfo())
	if err := ow.tds.EndSession(); err != nil {
		msg := fmt.Sprintf("oplWalkerImpl: %s: EndSession target DSS", origin)
		ow.lgr.Error(msg, "err", err)
	}
	return nil
}
