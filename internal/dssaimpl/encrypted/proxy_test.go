package encrypted

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestProxy(t *testing.T) {
	lgr := common.GetLogger()
	td := t.TempDir()
	rec1, id1, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	dss, err := MakeProxyDssa(lgr, td, []string{id1})
	require.NoError(t, err)
	require.Error(t, dss.NewSession())
	rec2, id2, err := common.AgeNewKeyPair()
	require.NoError(t, err)
	idss, err := common.AgeEncryptMsg([]byte(id2), rec1)
	require.NoError(t, err)
	require.NoError(t, dss.GetValueSetCb()(KeyIds, idss))
	require.NoError(t, dss.GetValueSetCb()(KeyRecs, []byte(rec2)))
	require.NoError(t, dss.GetValueSetCb()(KeyOpen, nil))
	require.NoError(t, dss.NewSession())
	require.NoError(t, dss.Mkdir(&dssa.DataEntry{Path: "/d1"}))
	require.NoError(t, dss.EndSession())
	require.NoError(t, dss.NewSession())
	de, err := dss.Stat("/d1")
	require.NoError(t, err)
	require.Equal(t, "/d1", de.Path)
	require.NoError(t, dss.EndSession())
	require.NoError(t, dss.GetValueSetCb()(KeyClose, nil))
	require.Error(t, dss.NewSession())
}