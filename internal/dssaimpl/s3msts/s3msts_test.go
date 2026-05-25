package s3msts

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
)

func wtf(ds S3DssaWithMsts, path_ string) error {
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
	SkipIf(t)
	ds := GetRepo(t)
	require.NotNil(t, ds)
	require.NoError(t, Cleanup(ds))
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	require.NoError(t, wtf(ds, "/d1/f1.txt"))
	require.NoError(t, wtf(ds, "/d2/f2.txt"))
	require.NoError(t, wtf(ds, "/d2/f3.txt"))
	require.NoError(t, wtf(ds, "/f0.txt"))
	ls, err := ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	require.NoError(t, ds.Msts().EndSession())
	require.NoError(t, ds.Msts().NewSession())
	ls, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
}

func TestBisBasicDirsAndFiles(t *testing.T) {
	SkipIf(t)
	ds := GetRepo(t)
	require.NotNil(t, ds)
	require.NoError(t, Cleanup(ds))
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d1", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, ds.Mkdir(&dssa.DataEntry{Path: "/d2", IsDir: true}))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	require.NoError(t, wtf(ds, "/d1/f1.txt"))
	require.NoError(t, wtf(ds, "/d2/f2.txt"))
	require.NoError(t, wtf(ds, "/d2/f3.txt"))
	require.NoError(t, wtf(ds, "/f0.txt"))
	ls, err := ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	require.NoError(t, ds.EndSession())
	require.NoError(t, ds.NewSession())
	ls, err = ds.List("/")
	require.NoError(t, err)
	require.Equal(t, 3, len(ls))
	ls, err = ds.List("/d2")
	require.NoError(t, err)
	require.Equal(t, 2, len(ls))
}
