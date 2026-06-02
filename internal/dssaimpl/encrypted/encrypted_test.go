package encrypted

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
)

func wtf(ds dssa.Dssa, path_ string) error {
	wc, err := ds.GetWriteCloser(path_)
	if err != nil {
		return err
	}
	if _, err = wc.Write([]byte(path_ + "\n")); err != nil {
		return err
	}
	if err = wc.Close(); err != nil {
		return err
	}
	return nil
}

func TestBasicDirsAndFiles(t *testing.T) {
	recs, ids, err := common.NewKeyPair()
	require.NoError(t, err)
	td := t.TempDir()
	ds, _ := MakeEncryptedDssa(
		common.GetNullLogger(),
		localfiles.MakeLocalFilesDssa(),
		td,
		[]string{ids},
		[]string{recs},
	)
	require.NotNil(t, ds)
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: td + "/d1", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: td + "/d2", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, wtf(ds, td+"/d1/f1.txt"))
	require.NoError(t, wtf(ds, td+"/d2/f2.txt"))
	require.NoError(t, wtf(ds, td+"/d2/f3.txt"))
	require.NoError(t, wtf(ds, td+"/f0.txt"))
	ls, err := ds.List(td + "/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	ls, err = ds.List(td + "/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List(td + "/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
}
