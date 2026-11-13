package share

import (
	"errors"
	"fmt"
	"sync"

	"github.com/t-beigbeder/vdasync/dssa"
)

type SessionManagerImpl struct {
	mx     sync.Mutex
	nextId int
	ids    map[int]bool
	trMx   sync.Mutex
	trId   int
}

// OpenSession implements [dssa.SessionManager].
func (s *SessionManagerImpl) OpenSession() (int, error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	if s.ids == nil {
		s.ids = map[int]bool{}
	}
	s.nextId++
	s.ids[s.nextId] = true
	return s.nextId, nil
}

// CloseSession implements [dssa.SessionManager].
func (s *SessionManagerImpl) CloseSession(id int) error {
	s.mx.Lock()
	defer s.mx.Unlock()
	if s.ids == nil {
		s.ids = map[int]bool{}
	}
	op, ok := s.ids[id]
	if !ok {
		return fmt.Errorf("session %d does not exist", id)
	}
	if !op {
		return fmt.Errorf("session %d already closed", id)
	}
	s.ids[id] = false
	return nil
}

// NewSessionTr implements [dssa.SessionManager].
func (s *SessionManagerImpl) NewSessionTr() (err error) {
	s.trMx.Lock()
	defer s.trMx.Unlock()
	if s.trId > 0 {
		return errors.New("SessionManagerImpl: transition session already created")
	}
	s.trId, err = s.OpenSession()
	return
}

// EndSessionTr implements [dssa.SessionManager].
func (s *SessionManagerImpl) EndSessionTr() error {
	s.trMx.Lock()
	defer s.trMx.Unlock()
	if s.trId == 0 {
		return errors.New("SessionManagerImpl: transition session already ended")
	}
	if err := s.CloseSession(s.trId); err != nil {
		return err
	}
	s.trId = 0
	return nil
}

var _ dssa.SessionManager = &SessionManagerImpl{}
