package walker

import (
	"io"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
)

func TestBasicDryrunSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	dssa1 := localfiles.MakeLocalFilesDssa()
	// _, err = dssa1.List(common.OsPath2DssPath(td1))
	// require.Nil(t, err)
	td2 := t.TempDir()

	walker := NewSynchronizer(lgr, 5, &config.SyncOptionsType{Dryrun: true},
		dssa1, dssa1, td2)
	sde, err := dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr := SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
}

func TestBasicActualSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	total := sad + saf + 1
	require.Nil(t, err)
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	dssa1 := localfiles.MakeLocalFilesDssa()
	// _, err = dssa1.List(common.OsPath2DssPath(td1))
	// require.Nil(t, err)
	td2 := t.TempDir()

	walker := NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true},
		dssa1, dssa1, td2)
	sde, err := dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr := SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	DisplaySyncResult(sr, io.Discard, true)
	_ = os.Stderr
	res := sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, total-1, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{},
		dssa1, dssa1, td2)
	sde, err = dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	res = sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, total-1, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true},
		dssa1, dssa1, td2)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	res = sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, 0, res.AggregatedCreated)
	require.Equal(t, 0, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	err = dssa1.Mkdir(&dssa.DataEntry{Path: path.Join(td1, "d00", "d99"), UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
	require.Nil(t, err)
	sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024*1024)
	require.Nil(t, err)
	newSubTotal := sad2 + saf2 + 1

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true},
		dssa1, dssa1, td2)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, total+newSubTotal, len(sr))
	res = sr[""]
	require.Equal(t, len(sr)-1, res.AggregatedChildrenNumber)
	require.Equal(t, newSubTotal, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{},
		dssa1, dssa1, td2)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, total+newSubTotal, len(sr))
	res = sr[""]
	require.Equal(t, len(sr)-1, res.AggregatedChildrenNumber)
	require.Equal(t, newSubTotal, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true},
		dssa1, dssa1, td2)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, total+newSubTotal, len(sr))
	res = sr[""]
	require.Equal(t, len(sr)-1, res.AggregatedChildrenNumber)
	require.Equal(t, 0, res.AggregatedCreated)
	require.Equal(t, 0, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)
}

func TestAugmentedTestDataSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	total := sad + saf + 1
	require.Nil(t, err)
	lgr.Debug("TestAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

	dssa1 := localfiles.MakeLocalFilesDssa()
	td2 := t.TempDir()

	walker := NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true}, dssa1, dssa1, td2)
	sde, err := dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr := SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	DisplaySyncResult(sr, io.Discard, true)
	res := sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, total-1, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{},
		dssa1, dssa1, td2)
	sde, err = dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	DisplaySyncResult(sr, io.Discard, true)
	res = sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, total-1, res.AggregatedCreated)
	require.Equal(t, 1, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

	walker = NewSynchronizer(lgr, 4, &config.SyncOptionsType{Dryrun: true}, dssa1, dssa1, td2)
	sde, err = dssa1.Stat(td1)
	require.Nil(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	sr = SyncResult(walker)
	require.NotNil(t, sr)
	require.Equal(t, sad+saf+1, len(sr))
	DisplaySyncResult(sr, io.Discard, true)
	res = sr[""]
	require.Equal(t, total-1, res.AggregatedChildrenNumber)
	require.Equal(t, 0, res.AggregatedCreated)
	require.Equal(t, 0, res.AggregatedUpdated)
	require.Equal(t, 0, res.AggregatedError)

}
