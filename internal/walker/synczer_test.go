package walker

import (
	"context"
	"errors"
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
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/encrypted"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/grpcclient"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/s3msts"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/sftpc"
	"github.com/t-beigbeder/vdasync/internal/remote"
)

func runSyncTest(lgr *slog.Logger, sDss, tDss dssa.Dssa, sde *dssa.DataEntry, tRoot string, so *config.SyncOptionsType) (syncRes map[string]*SyncEntryStatus, err error) {
	var walker Walker
	if walker, err = NewSynchronizer(lgr, 4, so, sDss, tDss, tRoot); err != nil {
		return
	}
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

func getTestDss(t *testing.T, hasS3, hasSftp, hasEncrypt, hasRencrypt bool) (dssa.Dssa, dssa.Dssa, s3msts.S3DssaWithMsts, dssa.Dssa, encrypted.EncryptedDssa, encrypted.EncryptedDssa, context.CancelFunc) {
	cli, cFunc, err := remote.GrpcGetTestClient(nil)
	require.NoError(t, err)
	dss1 := localfiles.MakeLocalFilesDssa()
	dss2 := grpcclient.MakeGrpcClient(common.GetNullLogger(), context.Background(), cli)
	var dss3 s3msts.S3DssaWithMsts
	var dss4 dssa.Dssa
	var dss5 encrypted.EncryptedDssa
	if hasS3 {
		s3msts.SkipIf(t)
		dss3 = s3msts.GetRepo(t)
		require.NoError(t, s3msts.Cleanup(dss3))
		require.NoError(t, dss3.Msts().NewSession())
	}
	if hasSftp {
		sftpc.SkipIf(t)
		dss4 = sftpc.GetSftpDss(t)
	}
	require.NoError(t, err)
	if hasEncrypt {
		recs, ids, err := common.AgeNewKeyPair()
		require.NoError(t, err)
		td := t.TempDir()
		dss5, _ = encrypted.MakeEncryptedDssa(
			common.GetNullLogger(),
			localfiles.MakeLocalFilesDssa(),
			td,
			[]string{ids},
			[]string{recs},
		)
		require.NotNil(t, dss5)
		require.NoError(t, dss5.NewSession())
	}
	return dss1, dss2, dss3, dss4, dss5, nil, cFunc
}

func TestBasicDryrunSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, dss5, _, cFunc := getTestDss(t, false, false, true, false)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss1, dss2, dss5} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		if tDss == dss5 {
			td2 = "/"
		}
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
	rLgr := common.GetNullLogger()
	lDss, rDss, _, _, dss5, _, cFunc := getTestDss(t, false, false, true, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{sDss: lDss, tDss: lDss},
		{sDss: lDss, tDss: rDss},
		{sDss: rDss, tDss: lDss},
		{sDss: rDss, tDss: rDss},
		{sDss: rDss, tDss: dss5},
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
		if tDss == dss5 {
			td2 = "/"
		}
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
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, dss5, _, cFunc := getTestDss(t, false, false, true, false)
	defer cFunc()

	for _, tDss := range []dssa.Dssa{dss1, dss2, dss5} {
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
		if tDss == dss5 {
			td2 = "/"
		}
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
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, dss5, _, cFunc := getTestDss(t, false, false, true, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{doRm: false, doCheck: false, tDss: dss1},
		{doRm: true, doCheck: false, tDss: dss1},
		{doRm: false, doCheck: false, tDss: dss2},
		{doRm: true, doCheck: false, tDss: dss2},
		{doRm: true, doCheck: true, tDss: dss1},
		{doRm: true, doCheck: true, tDss: dss2},
		{doRm: true, doCheck: true, tDss: dss5},
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
		if tDss == dss5 {
			td2 = "/"
		}
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
	walker, err := NewSynchronizer(lgr, 0, so, dss, dss, tRoot)
	require.NoError(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	syncRes := SyncResult(walker)
	require.NotNil(t, syncRes[""].Error)

	so = &config.SyncOptionsType{Dryrun: false}
	walker, err = NewSynchronizer(lgr, 0, so, dss, dss, tRoot)
	require.NoError(t, err)
	err = walker.Run(sde)
	require.Nil(t, err)
	syncRes = SyncResult(walker)
	require.NotNil(t, syncRes[""].Error)

}

func TestBasicS3DryrunSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, _, dss3, _, _, _, cFunc := getTestDss(t, true, false, false, false)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss3} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024)
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
	rLgr := common.GetNullLogger()
	lDss, _, rDss, _, _, _, cFunc := getTestDss(t, true, false, false, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{sDss: lDss, tDss: rDss},
	} {
		sDss := tsCfg.sDss
		tDss := tsCfg.tDss
		lgr := rLgr.With("sDss", fmt.Sprintf("%T", sDss), "tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024)
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
		sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024)
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

func TestBaseAugmentedTestS3DataSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, _, dss3, _, _, _, cFunc := getTestDss(t, true, false, false, false)
	defer cFunc()

	for _, tDss := range []dssa.Dssa{dss3} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := PrepareAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024)
		defer SetTestDirRW(td1, "source")
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		lgr.Debug("TestBaseAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NoError(t, dss3.Msts().EndSession())
		require.NoError(t, dss3.Msts().NewSession())

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestModAugmentedTestS3DataSynczer(t *testing.T) {
	type syncTestConfig struct {
		doRm    bool
		doCheck bool
		tDss    dssa.Dssa
	}
	rLgr := common.GetNullLogger()
	dss1, _, dss3, _, _, _, cFunc := getTestDss(t, true, false, false, false)
	defer cFunc()

	for _, tsCfg := range []syncTestConfig{
		{doRm: true, doCheck: true, tDss: dss3},
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
		lgr.Debug("TestModAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sad2, saf2, err := UpdateAugmentedTestFilesTree(td1, 5, 10, 3, 11*1024)
		require.Nil(t, err)
		_ = sad2 + saf2 + 1
		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: false, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.LessOrEqual(t, sr[""].AggregatedModChanged, 1)
	}
}

func TestBasicSftpDryrunSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, _, _, dss4, _, _, cFunc := getTestDss(t, false, true, false, false)
	defer cFunc()
	RecChmodRW(rLgr, 2, dss4, "/dau", "sftp")
	require.NoError(t, sftpc.Cleanup(dss4))
	for _, tDss := range []dssa.Dssa{dss4} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024)
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

func TestBasicSftpActualSynczer(t *testing.T) {
	type syncTestConfig struct {
		sDss dssa.Dssa
		tDss dssa.Dssa
	}
	rLgr := common.GetNullLogger()
	lDss, _, _, rDss, _, _, cFunc := getTestDss(t, false, true, false, false)
	defer cFunc()
	RecChmodRW(rLgr, 2, rDss, "/dau", "sftp")
	require.NoError(t, sftpc.Cleanup(rDss))

	for _, tsCfg := range []syncTestConfig{
		{sDss: lDss, tDss: rDss},
	} {
		sDss := tsCfg.sDss
		tDss := tsCfg.tDss
		lgr := rLgr.With("sDss", fmt.Sprintf("%T", sDss), "tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 101*1024)
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

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		err = sDss.Mkdir(&dssa.DataEntry{Path: path.Join(td1, "d00", "d99"), UserRights: dssa.Rights{Read: true, Write: true, Execute: true}})
		require.Nil(t, err)
		sad2, saf2, err := common.MakeTestFilesTree(path.Join(td1, "d00", "d99"), 5, 10, 3, 6*1024)
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

		sr, err = runSyncTest(lgr, sDss, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total+newSubTotal-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 0, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestBaseAugmentedTestSftpDataSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, _, _, dss4, _, _, cFunc := getTestDss(t, false, true, false, false)
	defer cFunc()
	RecChmodRW(rLgr, 2, dss4, "/dau", "sftp")
	require.NoError(t, sftpc.Cleanup(dss4))

	for _, tDss := range []dssa.Dssa{dss4} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sad, saf, err := PrepareAugmentedTestFilesTree(td1, 7, 100, 16, 6*1024)
		defer SetTestDirRW(td1, "source")
		require.Nil(t, err)
		total := sad + saf + 1
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		lgr.Debug("TestBaseAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError) // SFTP specific

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 3, sr[""].AggregatedUpdated) // SFTP specific
		require.Equal(t, 0, sr[""].AggregatedError)   // SFTP specific
	}
}

func TestModAugmentedTestSftpDataSynczer(t *testing.T) {
	type syncTestConfig struct {
		doRm    bool
		doCheck bool
		tDss    dssa.Dssa
	}
	rLgr := common.GetNullLogger()
	dss1, _, _, dss4, _, _, cFunc := getTestDss(t, false, true, false, false)
	defer cFunc()
	RecChmodRW(rLgr, 2, dss4, "/dau", "sftp")
	require.NoError(t, sftpc.Cleanup(dss4))

	for _, tsCfg := range []syncTestConfig{
		{doRm: true, doCheck: true, tDss: dss4},
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
		lgr.Debug("TestModAugmentedTestDataSynczer", "td1", td1, "sad", sad, "saf", saf)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, total-1, sr[""].AggregatedCreated)
		require.Equal(t, 1, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError) // SFTP specific

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true})
		require.Nil(t, err)
		require.Equal(t, total-1, sr[""].AggregatedChildrenNumber)
		require.Equal(t, 0, sr[""].AggregatedCreated)
		require.Equal(t, 3, sr[""].AggregatedUpdated)
		require.Equal(t, 0, sr[""].AggregatedError)

		sad2, saf2, err := UpdateAugmentedTestFilesTree(td1, 5, 10, 3, 11*1024)
		require.Nil(t, err)
		_ = sad2 + saf2 + 1
		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: false, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.NotEqual(t, 0, sr[""].AggregatedModChanged)

		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: doRm, Check: doCheck})
		require.Nil(t, err)
		require.Equal(t, 0, sr[""].AggregatedError)
		require.LessOrEqual(t, sr[""].AggregatedModChanged, 3)
	}

	RecChmodRW(rLgr, 2, dss4, "/dau", "sftp")
}

func TestFix01Synczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, _, _, cFunc := getTestDss(t, false, false, false, false)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss2} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		td2 := t.TempDir()
		require.NoError(t, os.Mkdir(path.Join(td1, "ds"), 0755))
		require.NoError(t, os.Mkdir(path.Join(td2, "ds"), 0755))
		require.NoError(t, os.Mkdir(path.Join(td2, "dd"), 0755))
		common.MakeTestFile(path.Join(td1, "ds", "fileSource.dat"), 100)
		common.MakeTestFile(path.Join(td2, "dd", "fileDestD.dat"), 100)
		common.MakeTestFile(path.Join(td2, "ds", "fileDestS.dat"), 100)

		sr, err := runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Dryrun: true, Rm: true})
		require.Equal(t, 0, sr[""].AggregatedError)
		DisplaySyncResult(sr, os.Stderr, true, true)
		sr, err = runSyncTest(lgr, dss1, tDss, sde, td2, &config.SyncOptionsType{Rm: true})
		DisplaySyncResult(sr, os.Stderr, true, true)
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestFix02Synczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, _, _, dss4, _, _, cFunc := getTestDss(t, false, true, false, false)
	defer cFunc()
	for _, tDss := range []dssa.Dssa{dss4} {
		lgr := rLgr.With("tDss", fmt.Sprintf("%T", tDss))
		td1 := t.TempDir()
		sde, err := dss1.Stat(td1)
		require.Nil(t, err)
		require.NoError(t, os.Mkdir(path.Join(td1, "ds"), 0755))
		common.MakeTestFile(path.Join(td1, "ds", "fileSource.dat"), 100)
		require.NoError(t, dss1.Symlink(path.Join(td1, "ds", "fileSource.dat"), path.Join(td1, "ds", "fileSource.sl")))
		require.NoError(t, dss1.Symlink(path.Join(td1, "ds", "notThat.dat"), path.Join(td1, "ds", "notThat.sl")))
		sr, err := runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: true})
		require.Equal(t, 0, sr[""].AggregatedError)
		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Rm: true})
		require.Equal(t, 0, sr[""].AggregatedError)
		sr, err = runSyncTest(lgr, dss1, tDss, sde, "/", &config.SyncOptionsType{Dryrun: true, Rm: true})
		require.Equal(t, 0, sr[""].AggregatedError)
	}
}

func TestExclSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, _, _, cFunc := getTestDss(t, false, false, false, false)
	defer cFunc()
	lgr := rLgr.With("tDss", fmt.Sprintf("%T", dss2))
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	total := sad + saf + 1
	sde, err := dss1.Stat(td1)
	require.Nil(t, err)
	td2 := t.TempDir()
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	elp := path.Join(td1, "el.txt")
	common.Lines2file(
		[]string{
			"^d00/d00/d00/f1.$",
			"^d00/d00/d01$",
		},
		elp)
	sr, err := runSyncTest(lgr, dss1, dss2, sde, td2, &config.SyncOptionsType{ExclListPath: elp})
	require.Greater(t, total, sr[""].AggregatedChildrenNumber)
	require.Greater(t, total, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

}

func TestInclSynczer(t *testing.T) {
	rLgr := common.GetNullLogger()
	dss1, dss2, _, _, _, _, cFunc := getTestDss(t, false, false, false, false)
	defer cFunc()
	lgr := rLgr.With("tDss", fmt.Sprintf("%T", dss2))
	td1 := t.TempDir()
	sad, saf, err := common.MakeTestFilesTree(td1, 7, 100, 16, 6*1024*1024)
	require.Nil(t, err)
	total := sad + saf + 1
	sde, err := dss1.Stat(td1)
	require.Nil(t, err)
	td2 := t.TempDir()
	lgr.Debug("TestBasicWalker", "td1", td1, "sad", sad, "saf", saf)

	elp := path.Join(td1, "el.txt")
	common.Lines2file(
		[]string{
			"^d00/d00/d00/f1.$",
			"^d00/d00/d01/",
		},
		elp)
	ilp := path.Join(td1, "il.txt")
	common.Lines2file(
		[]string{
			"^d00/d00/d00/f12",
			"^d00/d00/d01/f0.$",
		},
		ilp)

	sr, err := runSyncTest(lgr, dss1, dss2, sde, td2, &config.SyncOptionsType{ExclListPath: elp, InclListPath: ilp})
	require.Greater(t, total+1, sr[""].AggregatedChildrenNumber)
	require.Greater(t, total+1, sr[""].AggregatedCreated)
	require.Equal(t, 1, sr[""].AggregatedUpdated)
	require.Equal(t, 0, sr[""].AggregatedError)

}

type simpleStepFunc func(string, *simpleStepsDesc, string, string) error

type simpleStep struct {
	ssn string
	ssf simpleStepFunc
}

type simpleStepsDesc struct {
	omit                bool
	dispRes             bool
	rLgr                *slog.Logger
	concurrency         int
	syncOptions         *config.SyncOptionsType
	sDss                dssa.Dssa
	srGet               func() string
	tDss                dssa.Dssa
	trGet               func() string
	tdGet               func() string
	simpleSteps         []simpleStep
	cLgr                *slog.Logger
	gotSr, gotTr, gotTd string
	lastWk              Walker
}

func runSyncAndCheck(
	ssn string, ssd *simpleStepsDesc,
	syncOptions *config.SyncOptionsType,
	sourceDs dssa.Dssa, sourceRoot string,
	targetDs dssa.Dssa, targetRoot string,
) (*SyncEntryStatus, error) {
	wk, err := RunSynchronizer(ssd.cLgr.With("subStep", ssn), ssd.concurrency, syncOptions, sourceDs, sourceRoot, targetDs, targetRoot)
	if err != nil {
		return nil, err
	}
	sr := SyncResult(wk)
	if sr == nil {
		return nil, errors.New("SyncResult is nil")
	}
	if sr[""].AggregatedError != 0 {
		DisplaySyncResult(sr, os.Stderr, true, false)
		return nil, fmt.Errorf("runSyncAndCheck: AggregatedError is %d", sr[""].AggregatedError)
	}
	ssd.lastWk = wk
	return sr[""], nil
}

func checkSrRef(chkSr, refSr *SyncEntryStatus) error {
	chkSrv := SyncEntryStatus{
		AggregatedSize:           chkSr.AggregatedSize,
		AggregatedChildrenNumber: chkSr.AggregatedChildrenNumber,
		AggregatedCreated:        chkSr.AggregatedCreated,
		AggregatedUpdated:        chkSr.AggregatedUpdated,
		AggregatedRemoved:        chkSr.AggregatedRemoved,
		AggregatedModChanged:     chkSr.AggregatedModChanged,
		AggregatedError:          chkSr.AggregatedError,
	}
	refSrv := SyncEntryStatus{
		AggregatedSize:           refSr.AggregatedSize,
		AggregatedChildrenNumber: refSr.AggregatedChildrenNumber,
		AggregatedCreated:        refSr.AggregatedCreated,
		AggregatedUpdated:        refSr.AggregatedUpdated,
		AggregatedRemoved:        refSr.AggregatedRemoved,
		AggregatedModChanged:     refSr.AggregatedModChanged,
		AggregatedError:          refSr.AggregatedError,
	}
	if chkSrv != refSrv {
		return fmt.Errorf("checkSr: checked %+v reference %+v", chkSrv, refSrv)
	}
	return nil
}

func checkStep(sn string, ssf simpleStepFunc, ssd *simpleStepsDesc) error {
	ssd.cLgr = ssd.rLgr.With("tDss", fmt.Sprintf("%T", ssd.tDss), "step", sn)
	if ssd.sDss == nil {
		ssd.sDss = localfiles.MakeLocalFilesDssa()
	}
	if ssd.gotSr == "" {
		if ssd.srGet != nil {
			ssd.gotSr = ssd.srGet()
		} else {
			ssd.gotSr = "/"
		}
	}
	if ssd.tDss == nil {
		ssd.tDss = localfiles.MakeLocalFilesDssa()
	}
	if ssd.gotTr == "" {
		if ssd.trGet != nil {
			ssd.gotTr = ssd.trGet()
		} else {
			ssd.gotTr = "/"
		}
	}
	if ssd.gotTd == "" {
		ssd.gotTd = ssd.tdGet()
	}
	ssd.cLgr.Info("checkStep")
	if err := ssf(sn, ssd, ssd.gotSr, ssd.gotTr); err != nil {
		return err
	}
	drSo := *ssd.syncOptions
	drSo.Dryrun = true
	drSr, err := runSyncAndCheck("dryrun1", ssd, &drSo, ssd.sDss, ssd.gotSr, ssd.tDss, ssd.gotTr)
	if err != nil {
		return err
	}
	acSr, err := runSyncAndCheck("actual", ssd, ssd.syncOptions, ssd.sDss, ssd.gotSr, ssd.tDss, ssd.gotTr)
	if err != nil {
		return err
	}
	if err := checkSrRef(acSr, drSr); err != nil {
		return err
	}
	dr2Sr, err := runSyncAndCheck("dryrun2", ssd, &drSo, ssd.sDss, ssd.gotSr, ssd.tDss, ssd.gotTr)
	if err != nil {
		return err
	}
	dr2Sr.AggregatedSize = 0
	dr2Sr.AggregatedChildrenNumber = 0
	if err := checkSrRef(dr2Sr, &SyncEntryStatus{}); err != nil {
		return err
	}
	bckSr, err := runSyncAndCheck("backward", ssd, ssd.syncOptions, ssd.tDss, ssd.gotTr, ssd.sDss, ssd.gotTd)
	if err != nil {
		return err
	}
	// if err := checkSrRef(bckSr, acSr); err != nil {
	// 	return err
	// }
	_ = bckSr
	dr3Sr, err := runSyncAndCheck("dryrun3", ssd, &drSo, ssd.sDss, ssd.gotSr, ssd.sDss, ssd.gotTd)
	if err != nil {
		return err
	}
	dr3Sr.AggregatedSize = 0
	dr3Sr.AggregatedChildrenNumber = 0
	if err := checkSrRef(dr3Sr, &SyncEntryStatus{}); err != nil {
		return err
	}
	if ssd.dispRes {
		DisplaySyncResult(SyncResult(ssd.lastWk), os.Stderr, true, true)
	}
	return nil
}

func stepMakeTestFilesTree(ssn string, ssd *simpleStepsDesc, sr, tr string) error {
	_, _, err := common.MakeTestFilesTree(sr, 7, 100, 16, 6*1024)
	return err
}

func stepUtilMkdir(ssd *simpleStepsDesc, root, dp string) error {
	return common.MakeParents(ssd.sDss, path.Join(root, dp))
}

func stepUtilMkfile(ssd *simpleStepsDesc, root, fp string) error {
	ffp := path.Join(root, fp)
	if err := common.MakeParents(ssd.sDss, path.Dir(ffp)); err != nil {
		return err
	}
	return common.MakeTestFile(ffp, 6*1024)
}

func stepUtilRmdir(lgr *slog.Logger, ssd *simpleStepsDesc, root, dp string) error {
	_, err := RemoveAll(lgr, 2, ssd.sDss, path.Join(root, dp), "source", false)
	return err
}

func stepMakeTest1Base(ssn string, ssd *simpleStepsDesc, sr, tr string) error {
	if err := stepUtilMkfile(ssd, sr, "d1/d11/f111.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d11/d111/f1111.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d12/f121.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d13/f131.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d13/f132.dat"); err != nil {
		return err
	}
	if err := stepUtilMkdir(ssd, sr, "d1/d13e"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d14/f141.dat"); err != nil {
		return err
	}
	return nil
}

func stepMakeTest1Step2(ssn string, ssd *simpleStepsDesc, sr, tr string) error {
	if err := stepUtilMkfile(ssd, sr, "d1/d11/f112.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d11/d111/f1112.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d11/d112/f1121.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d12/f122.dat"); err != nil {
		return err
	}
	if err := stepUtilMkfile(ssd, sr, "d1/d14/f142.dat"); err != nil {
		return err
	}
	if err := stepUtilRmdir(ssd.cLgr.With("subStep", ssn), ssd, sr, "d1/d14"); err != nil {
		return err
	}
	return nil
}

func TestSimpleSteps(t *testing.T) {
	getTd := func() string {
		return t.TempDir()
	}
	nullLgr := common.GetNullLogger()
	dbgLgr := common.DbgLogger()
	infoLgr := common.InfoLogger()
	_, _, _ = nullLgr, dbgLgr, infoLgr
	_, rDss, _, _, eDss, _, cFunc := getTestDss(t, false, true, true, false)
	defer cFunc()

	testSet := []simpleStepsDesc{
		{
			omit: false,
			rLgr: nullLgr, syncOptions: &config.SyncOptionsType{Rm: true},
			srGet: getTd, trGet: getTd, tdGet: getTd,
			simpleSteps: []simpleStep{
				{"stepMakeTestFilesTree", stepMakeTestFilesTree},
			},
		},
		{
			omit: false,
			rLgr: nullLgr, syncOptions: &config.SyncOptionsType{Rm: true},
			srGet: getTd, trGet: getTd, tdGet: getTd,
			simpleSteps: []simpleStep{
				{"stepMakeTest1Base", stepMakeTest1Base},
				{"stepMakeTest1Step2", stepMakeTest1Step2},
			},
		},
		{
			omit: false,
			rLgr: nullLgr, syncOptions: &config.SyncOptionsType{Rm: true},
			srGet: getTd,
			tDss:  rDss,
			trGet: getTd,
			tdGet: getTd,
			simpleSteps: []simpleStep{
				{"stepMakeTest1Base", stepMakeTest1Base},
				{"stepMakeTest1Step2", stepMakeTest1Step2},
			},
		},
		{
			omit:    false,
			dispRes: true,
			rLgr:    nullLgr, syncOptions: &config.SyncOptionsType{Rm: true},
			srGet: getTd,
			tDss:  eDss,
			tdGet: getTd,
			simpleSteps: []simpleStep{
				{"stepMakeTest1Base", stepMakeTest1Base},
				{"stepMakeTest1Step2", stepMakeTest1Step2},
			},
		},
	}
	for _, test := range testSet {
		if test.omit {
			continue
		}
		for _, sst := range test.simpleSteps {
			require.NoError(t, checkStep(sst.ssn, sst.ssf, &test))
		}
		require.True(t, true)
	}
}
