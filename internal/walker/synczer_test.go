package walker

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/config"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/s3msts"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func runSyncTest(lgr *slog.Logger, sDss, tDss dssa.Dssa, sde *dssa.DataEntry, tRoot string, so *config.SyncOptionsType) (syncRes map[string]*SyncEntryStatus, err error) {
	walker := NewSynchronizer(lgr, 4, so, sDss, tDss, tRoot)
	if err = walker.Run(sde); err != nil {
		return
	}
	syncRes = SyncResult(walker)
	if syncRes != nil {
		_ = io.Discard
		_ = os.Stderr
		DisplaySyncResult(syncRes, io.Discard, true, false)
	}
	return
}

func getTestDss(t *testing.T, hasS3 bool) (dssa.Dssa, dssa.Dssa, s3msts.S3DssaWithMsts, context.CancelFunc) {
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.NoError(t, err)
	dss1 := localfiles.MakeLocalFilesDssa()
	dss2 := grpcclient.MakeGrpcClient(context.Background(), cli)
	var dss3 s3msts.S3DssaWithMsts
	if hasS3 {
		s3msts.SkipIf(t)
		dss3 = s3msts.GetRepo(t)
		require.NoError(t, s3msts.Cleanup(dss3))
		require.NoError(t, dss3.Msts().NewSession())
	}
	require.NoError(t, err)
	return dss1, dss2, dss3, cFunc
}

func TestBasicDryrunSynczer(t *testing.T) {
	rLgr := common.GetLogger()
	dss1, dss2, _, cFunc := getTestDss(t, false)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss1, dss2} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestBasicActualSynczer(t *testing.T) {
	type syncTestConfig struct {
		sDss dssa.Dssa
		tDss dssa.Dssa
	}
	rLgr := common.GetLogger()
	lDss, rDss, _, cFunc := getTestDss(t, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{sDss: lDss, tDss: lDss},
		{sDss: lDss, tDss: rDss},
		{sDss: rDss, tDss: lDss},
		{sDss: rDss, tDss: rDss},
	} {
		sDss := tsCfg.sDss
		tDss := tsCfg.tDss
		lgr := rLgr.With("sDss", fmt.Sprintf("%T", sDss), "tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := sDss.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		lgr.Debug("TestBasicActualSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		err = sDss.Mkdir(&dssa.DataEntry{Path: path.Join(td1, "d00", "d99"), UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
		require.Nil(t, err)
		sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024*1024)
		require.Nil(t, err)
		newSubTotal := sad2 + saf2 + 1

		sr, err = runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestBaseAugmentedTestDataSynczer(t *testing.T) {
	rLgr := common.GetLogger()
	dss1, dss2, _, cFunc := getTestDss(t, false)
	defer cFunc()

	for _, tDss := range []dssa.Dssa{dss1, dss2} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := PrepareAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		defer SetTestDirRW(td1, "source")
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		defer SetTestDirRW(td2, "target")
		lgr.Debug("TestBaseAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestModAugmentedTestDataSynczer(t *testing.T) {
	type syncTestConfig struct {
		doRm    bool
		doCheck bool
		tDss    dssa.Dssa
	}
	rLgr := common.GetLogger()
	dss1, dss2, _, cFunc := getTestDss(t, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{doRm: false, doCheck: false, tDss: dss1},
		{doRm: true, doCheck: false, tDss: dss1},
		{doRm: false, doCheck: false, tDss: dss2},
		{doRm: true, doCheck: false, tDss: dss2},
		{doRm: true, doCheck: true, tDss: dss1},
		{doRm: true, doCheck: true, tDss: dss2},
	} {
		doRm := tsCfg.doRm
		doCheck := tsCfg.doCheck
		tDss := tsCfg.tDss

		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss)).With("doRm", doRm).With("doCheck", doCheck)
		td1 := t.TempDir()
		sad, saf, err := PrepareAugmentedTestFilesTree(td1, 7, 100, 16, 17*1024)
		defer SetTestDirRW(td1, "source")
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		defer SetTestDirRW(td2, "target")
		lgr.Debug("TestModAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sad2, saf2, err := UpdateAugmentedTestFilesTree(td1, 5, 10, 3, 11*1024)
		require.Nil(t, err)
		_ = sad2 + saf2 + 1
		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: false, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.LessOrEqual(t, sr[""].AggregatedModChanged, 1)
	}
}

func TestNoTarget(t *testing.T) {
	dss := localfiles.MakeLocalFilesDssa()
	lgr := common.GetLogger()
	td1 := t.TempDir()
	td2 := t.TempDir()
	tRoot := path.Join(td2, "noSuchRoot")
	sde, err := dss.Stat(td1)
	require.Nil(t, err)
	so := &config.SyncOptionsType{Dryrun: true}
	walker := NewSynchronizer(lgr, 0, so, dss, dss, tRoot)
	err = walker.Run(sde)
	require.Nil(t, err)
	syncRes := SyncResult(walker)
	require.NotNil(t, syncRes[""].Error)

	so = &config.SyncOptionsType{Dryrun: false}
	walker = NewSynchronizer(lgr, 0, so, dss, dss, tRoot)
	err = walker.Run(sde)
	require.Nil(t, err)
	syncRes = SyncResult(walker)
	require.NotNil(t, syncRes[""].Error)

}

func TestBasicS3DryrunSynczer(t *testing.T) {
	rLgr := common.GetLogger()
	dss1, _, dss3, cFunc := getTestDss(t, true)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss3} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestBasicS3ActualSynczer(t *testing.T) {
	type syncTestConfig struct {
		sDss dssa.Dssa
		tDss dssa.Dssa
	}
	rLgr := common.GetLogger()
	lDss, _, rDss, cFunc := getTestDss(t, true)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{sDss: lDss, tDss: rDss},
	} {
		sDss := tsCfg.sDss
		tDss := tsCfg.tDss
		lgr := rLgr.With("sDss", fmt.Sprintf("%T", sDss), "tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := sDss.Stat(td1)
		require.Nil(t, err)
		lgr.Debug("TestBasicActualSynczer", "td1", td1, "sad", sad, "saf", saf, "dbg", 2)

		sr, err := runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NoError(t, rDss.Msts().EndSession())
		require.NoError(t, rDss.Msts().NewSession())

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		err = sDss.Mkdir(&dssa.DataEntry{Path: path.Join(td1, "d00", "d99"), UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
		require.Nil(t, err)
		sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024*1024)
		require.Nil(t, err)
		newSubTotal := sad2 + saf2 + 1

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, newSubTotal, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NoError(t, rDss.Msts().EndSession())
		require.NoError(t, rDss.Msts().NewSession())

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}
