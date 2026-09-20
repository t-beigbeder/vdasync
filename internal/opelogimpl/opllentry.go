package opelogimpl

import (
	"errors"
	"fmt"
	"path"
	"slices"
	"strings"

	"github.com/t-beigbeder/vdasync/internal/common"
)

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
	if !ole.source().isPresent() || !ole.target().isAbsent() {
		return nil
	}
	return ole.target().create()
}

func (ole *oplLogicalEntry) process() error {
	ole.lgr().Debug("process: start")
	var (
		err error
	)
	for goal := range strings.SplitSeq("load,create,update,verify", ",") {
		if !ole.owi.impliesGoal(goal) {
			continue
		}
		switch goal {
		case "load":
			err = ole.load()
		case "create":
			err = ole.create()
		case "update":
			err = errors.ErrUnsupported
		case "verify":
			err = errors.ErrUnsupported
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
