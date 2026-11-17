package share

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type stTestSubSess struct {
	SessionManagerImpl
	subId int
}

func (stss *stTestSubSess) OpenSession() (int, error) {
	id, err := stss.SessionManagerImpl.OpenSession()
	id += 100
	stss.subId = id
	return id, err
}

func (stss *stTestSubSess) CloseSession(id int) error {
	err := stss.SessionManagerImpl.CloseSession(stss.subId - 100)
	return err
}

func TestStruct(t *testing.T) {
	stss := &stTestSubSess{}
	id, err := stss.OpenSession()
	require.NoError(t, err)
	err = stss.CloseSession(id)
	require.NoError(t, err)
}
