package opelogimpl

import (
	"errors"
	"fmt"
	"log/slog"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

type oplLogicalEntry struct {
	hasChanges bool
	relPath    string
	plgr       *slog.Logger
	owi        *oplWalkerImpl
	le         *opelog.LogicalEntry
	sHasParent bool
	sChildrenQ []string
	tHasParent bool
	tChildrenQ []string
}

func (ole *oplLogicalEntry) lgr() *slog.Logger {
	return ole.plgr.With("relPath", ole.relPath)
}

func (ole *oplLogicalEntry) detail(msg string, args ...any) {
	ole.lgr().Log(ole.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ole *oplLogicalEntry) source() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole}
}

func (ole *oplLogicalEntry) target() *oplStoredEntry {
	return &oplStoredEntry{oplLogicalEntry: ole, isTarget: true}
}

func (ole *oplLogicalEntry) requiresCreate() (yes bool) {
	if !ole.source().isPresent() {
		return
	}
	if !ole.target().isAbsent() {
		return
	}
	return true
}

func (ole *oplLogicalEntry) requiresUpdate() (yes bool) {
	if !ole.source().isPresent() {
		return
	}
	if !ole.target().isPresent() {
		return
	}
	// may need remove anyway, that's part of this test for update in the broad sense
	return true
}

func (ole *oplLogicalEntry) qPfx() string {
	s := "1"
	if ole.source().isAbsent() {
		s = "0"
	}
	if ole.target().isAbsent() {
		s += "0"
	} else {
		s += "1"
	}
	return s
}

func (ole *oplLogicalEntry) queueChildren() error {
	merged := slices.Clone(ole.sChildrenQ)
	for _, childRp := range ole.tChildrenQ {
		if !slices.Contains(ole.sChildrenQ, childRp) {
			merged = append(merged, childRp)
		}
	}
	ole.sChildrenQ, ole.tChildrenQ = nil, nil
	for _, childRp := range merged {
		if err := ole.owi.oplq.Put(ole.qPfx() + path.Join(ole.relPath, childRp)); err != nil {
			return ole.owi.owErr(ole.lgr(), "oplq.Put error", err)
		}
	}
	ole.le.DepCount = int32(len(merged))
	return nil
}

func (ole *oplLogicalEntry) computeNext() error {
	ole.owi.mx.Lock()
	defer ole.owi.mx.Unlock()
	if !ole.source().isKnown() || !ole.target().isKnown() || ole.le.DepCount != 0 {
		return nil
	}
	if ole.relPath == "" {
		return ole.owi.oplq.Close()
	}
	prp := common.ParentPath(ole.relPath)
	ple, err := ole.owi.oplm.GetLogicalEntry(prp)
	if err != nil {
		return err
	}
	if ple.DepCount == 0 {
		return fmt.Errorf("parent %s for entry %s dependency count is nul", prp, ole.relPath)
	}
	ple.DepCount--
	if err = ole.owi.oplm.PutLogicalEntry(prp, ple); err != nil {
		return err
	}
	if ple.DepCount == 0 {
		oPle := &oplLogicalEntry{le: ple}
		if err = ole.owi.oplq.Put(oPle.qPfx() + prp); err != nil {
			return err
		}
	}
	return nil
}

func (ole *oplLogicalEntry) load() error {
	ole.lgr().Debug("load: start")
	if err := ole.source().load(); err != nil {
		return err
	}
	if err := ole.target().load(); err != nil {
		return err
	}
	if err := ole.source().checkInventory(); err != nil {
		return err
	}

	return nil
}

func (ole *oplLogicalEntry) create() error {
	ole.lgr().Debug("create: start")
	sose, tose := ole.source(), ole.target()
	if !ole.owi.hasGoal("load") {
		// FIXME: perhaps is there a better way to factorize
		if err := sose.load(); err != nil {
			return err
		}
		if err := tose.load(); err != nil {
			return err
		}
	}

	if !sose.isPresent() || !tose.isAbsent() {
		return nil
	}
	tde := sose.currentState().ToDataEntry(tose.fullPath())
	if sose.currentState().IsDir {
		tose.detail("dss.Mkdir", "path", tose.fullPath())
		if err := tose.dss().Mkdir(tde); err != nil {
			tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_MKDIR, err.Error())
			return nil
		}
		// TODO: perform DIRUP or CHMOD if done
		return nil
	}
	if sose.currentState().IsSymLink {
		tose.detail("dss.Symlink", "SymLinkTarget", sose.currentState().SymLinkTarget, "path", tose.fullPath())
		if err := tose.dss().Symlink(sose.currentState().SymLinkTarget, tose.fullPath()); err != nil {
			tose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_WRITE, err.Error())
			return nil
		}
		// TODO: perform CHMOD
		
	}
	if err := ole.copy(); err != nil {
		return err
	}
	// TODO: perform CHMOD
	return nil
}

func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	var (
		err error
	)
	for goal := range strings.SplitSeq("load,create,update,verify", ",") {
		if !ole.owi.hasGoal(goal) {
			continue
		}
		switch goal {
		case "load":
			err = ole.load()
		case "create":
			err = ole.create()
		default:
			err = errors.ErrUnsupported
		}
		if err != nil {
			break
		}
	}
	if err != nil {
		return err
	}
	if err = ole.queueChildren(); err != nil {
		return err
	}
	if err = ole.computeNext(); err != nil {
		return err
	}

	return err
}

type oplStoredEntry struct {
	*oplLogicalEntry
	isTarget bool
}

func (ose *oplStoredEntry) pfx() string {
	if ose.isTarget {
		return "{T}"
	} else {
		return "{S}"
	}
}

func (ose *oplStoredEntry) lgr() *slog.Logger {
	return ose.plgr.With("path", path.Join(ose.pfx(), ose.relPath))
}

func (ose *oplStoredEntry) detail(msg string, args ...any) {
	ose.lgr().Log(ose.owi.bg, slog.LevelDebug+2, msg, args...)
}

func (ose *oplStoredEntry) root() string {
	if ose.isTarget {
		return ose.owi.tRoot
	} else {
		return ose.owi.sRoot
	}
}

func (ose *oplStoredEntry) fullPath() string {
	return path.Join(ose.root(), ose.relPath)
}

func (ose *oplStoredEntry) events() (evs *[]*opelog.Event) {
	if ose.isTarget {
		if ose.le.TargetEvents == nil {
			ose.le.TargetEvents = []*opelog.Event{}
		}
		evs = &ose.le.TargetEvents
	} else {
		if ose.le.SourceEvents == nil {
			ose.le.SourceEvents = []*opelog.Event{}
		}
		evs = &ose.le.SourceEvents
	}
	return
}

func (ose *oplStoredEntry) existOrAbsEv() (ev *opelog.Event) {
	evs := ose.events()
	for i := range slices.Backward(*evs) {
		if (*evs)[i].Kind == opelog.EVT_ABS || (*evs)[i].Kind == opelog.EVT_EXIST {
			return (*evs)[i]
		}
	}
	return nil
}

func (ose *oplStoredEntry) currentEvent() *opelog.Event {
	evs := ose.events()
	if len(*evs) == 0 {
		return nil
	}
	return (*evs)[len(*evs)-1]
}

func (ose *oplStoredEntry) currentChecksums() (string, string) {
	ev := ose.currentEvent()
	if ev == nil {
		return "", ""
	}
	return ev.Checksums, ev.Error
}

func (ose *oplStoredEntry) states() (sts *[]*opelog.StoredEntry) {
	if ose.isTarget {
		sts = &ose.le.TargetStates
	} else {
		sts = &ose.le.SourceStates
	}
	return
}

func (ose *oplStoredEntry) currentState() *opelog.StoredEntry {
	sts := ose.states()
	if len(*sts) == 0 {
		return nil
	}
	return (*sts)[len(*sts)-1]
}

func (ose *oplStoredEntry) dss() (ds dssa.Dssa) {
	if ose.isTarget {
		ds = ose.owi.tds
	} else {
		ds = ose.owi.sds
	}
	return
}

func (ose *oplStoredEntry) newState(se *opelog.StoredEntry) {
	sts := ose.states()
	if len(*sts) == 0 {
		*sts = append(*sts, se)
		return
	}
	if se.Equal((*sts)[len(*sts)-1]) {
		return
	}
	*sts = append(*sts, se)
}

func (ose *oplStoredEntry) newEvent(kind opelog.EventCode, origin opelog.OriginCode, sErr string) {
	evs := ose.events()
	*evs = append(*evs,
		&opelog.Event{
			Kind: kind, Origin: origin, TimeStamp: time.Now().Unix(),
			StateIndex: int32(len(*ose.states()) - 1), Error: sErr})
	ose.hasChanges = true
	if sErr != "" {
		ose.lgr().Error("oplStoredEntry.newEvent", "kind", kind, "origin", origin, "err", sErr)
	}
}

func (ose *oplStoredEntry) isKnown() bool {
	eev := ose.existOrAbsEv()
	if eev == nil {
		return false
	}
	// may be on error anyway
	return true
}

func (ose *oplStoredEntry) isPresent() bool {
	eev := ose.existOrAbsEv()
	if eev == nil || eev.Error != "" || eev.Kind != opelog.EVT_EXIST {
		return false
	}
	return true
}

func (ose *oplStoredEntry) isAbsent() bool {
	eev := ose.existOrAbsEv()
	if eev == nil || eev.Error != "" || eev.Kind != opelog.EVT_ABS {
		return false
	}
	return true
}

func (ose *oplStoredEntry) hasParent() bool {
	if ose.isTarget {
		return ose.tHasParent
	} else {
		return ose.sHasParent
	}
}

func (ose *oplStoredEntry) setChildrenQ(children []string) {
	if ose.isTarget {
		ose.tChildrenQ = slices.Clone(children)
	} else {
		ose.sChildrenQ = slices.Clone(children)
	}
}

func (ose *oplStoredEntry) load() error {
	eev := ose.existOrAbsEv()
	if eev != nil && eev.Error == "" {
		return nil
	}
	ose.lgr().Debug("load: start", "hasParent", ose.hasParent())
	if !ose.hasParent() {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_UNSPECIFIED, "")
		return nil
	}
	ose.detail("dss.Stat", "path", ose.fullPath())
	de, err := ose.dss().Stat(ose.fullPath())
	if err != nil && !de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, err.Error())
		return nil
	}
	if de.ErrNotExist {
		ose.newState(&opelog.StoredEntry{})
		ose.newEvent(opelog.EVT_ABS, opelog.ORI_STAT, "")
		return nil
	}
	var children, fCn []string
	var cdes []*dssa.DataEntry
	if de.IsDir {
		ose.detail("dss.List", "path", ose.fullPath())
		cdes, err = ose.dss().List(ose.fullPath())
		if err != nil {
			se := opelog.FromDataEntry(de, nil)
			se.IsPresent = true
			ose.newState(se)
			ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, err.Error())
			return nil
		}

		for _, cde := range cdes {
			if cde.IsDir {
				children = append(children, path.Base(cde.Path))
			} else {
				fCn = append(fCn, path.Base(cde.Path))
			}
			children = slices.Concat(children, fCn)
			ose.setChildrenQ(children)
		}
	}

	se := opelog.FromDataEntry(de, children)
	se.IsPresent = true
	ose.newState(se)
	ose.newEvent(opelog.EVT_EXIST, opelog.ORI_STAT, "")

	return nil
}

func (ose *oplStoredEntry) checkInventory() error {
	if ose.le.InvChecksums == "" || ose.owi.owo.NoInvCheck || !ose.isPresent() {
		return nil
	}
	if ose.currentState().IsDir {
		return nil
	}

	eev := ose.existOrAbsEv()
	if eev.Checksums != "" {
		return nil
	}
	if ose.owi.hasGoal("create") && ose.requiresCreate() {
		return nil
	}
	if ose.owi.hasGoal("update") && ose.requiresUpdate() {
		return nil
	}

	ose.detail("dss.GetReadCloser", "path", ose.fullPath(), "algos", ose.owi.owo.InvCsAlgos)
	rr, err := ose.dss().GetReadCloser(ose.fullPath())
	if err != nil {
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	defer rr.Close()
	if eev.Checksums, err = common.ReaderChecksum(rr, ose.owi.owo.InvCsAlgos); err != nil {
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	if eev.Checksums != ose.le.InvChecksums {
		err := fmt.Errorf(
			"inventory checksum failed: inv %s actual %s",
			ose.le.InvChecksums, eev.Checksums)
		ose.newEvent(opelog.EVT_UNSPECIFIED, opelog.ORI_READ, err.Error())
		return nil
	}
	return nil
}
