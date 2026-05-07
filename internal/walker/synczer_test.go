package walker

import (
	"io"
	"log/slog"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
	"github.com/t-beigbeder/otvl_dtacsy/internal/dssaimpl/localfiles"
)

func runSyncTest(lgr *slog.Logger, dss dssa.Dssa, sde *dssa.DataEntry, tRoot string, so *config.SyncOptionsType) (syncRes map[string]*SyncEntryStatus, err error) {
	walker := NewSynchronizer(lgr, 4, so, dss, dss, tRoot)
	if err = walker.Run(sde); err != nil {
		return
	}
	syncRes = SyncResult(walker)
	if syncRes != nil {
		DisplaySyncResult(syncRes, io.Discard, true)
	}
	return
}

func TestBasicDryrunSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	total := sad + saf + 1
	dss := localfiles.MakeLocalFilesDssa()
	sde, err := dss.Stat(td1)
	require.Nil(t, err)
	td2 := t.TempDir()
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	sr, err := runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, total-1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)
}

func TestBasicActualSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	total := sad + saf + 1
	dss := localfiles.MakeLocalFilesDssa()
	sde, err := dss.Stat(td1)
	require.Nil(t, err)
	td2 := t.TempDir()
	lgr.Debug("TestBasicActualSynczer", "td1", td1, "sad", sad, "saf", saf)

	sr, err := runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, total-1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{})
	require.Nil(t, err)
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, total-1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, 0, sr[""].AggregatedCreated)
	require.Equal(t, 0, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	err = dss.Mkdir(&dssa.DataEntry{Path: path.Join(td1, "d00", "d99"), UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
	require.Nil(t, err)
	sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024*1024)
	require.Nil(t, err)
	newSubTotal := sad2 + saf2 + 1

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{})
	require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, 0, sr[""].AggregatedCreated)
	require.Equal(t, 0, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)
}

func TestAugmentedTestDataSynczer(t *testing.T) {
	lgr := common.GetLogger()
	td1 := t.TempDir()
	sad, saf, err := common.MakeAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	total := sad + saf + 1
	dss := localfiles.MakeLocalFilesDssa()
	sde, err := dss.Stat(td1)
	require.Nil(t, err)
	td2 := t.TempDir()
	lgr.Debug("TestAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

	sr, err := runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, total-1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{})
	require.Nil(t, err)
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, total-1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

	sr, err = runSyncTest(lgr, dss, sde, td2, &config.SyncOptionsType{Dryrun: true})
	require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
	require.Equal(t, 0, sr[""].AggregatedCreated)
	require.Equal(t, 0, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

}
