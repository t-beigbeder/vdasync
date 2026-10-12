package opelogimpl

import (
	"errors"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/opelog"
)

func makeTestM2fManager(path_ string, bckLimit int, testMarsh bool, std, ttd string) (*m2fOplMng, error) {
	oplm, err := MakeM2fManager(path_, bckLimit)
	if err != nil {
		return nil, err
	}
	oplmi, ok := oplm.(*m2fOplMng)
	if !ok {
		return nil, errors.New("not an *m2fOplMng")
	}
	oplmi.testMarsh = testMarsh
	if err = oplm.Create(std, ttd); err != nil {
		return nil, err
	}
	if err = oplm.Open(false); err != nil {
		return nil, err
	}
	_, err = oplm.AddSession("ds", &opelog.SessionInfo{})
	if err != nil {
		return nil, err
	}
	_, err = oplm.AddSession("di", &opelog.SessionInfo{IsInventory: true})
	if err != nil {
		return nil, err
	}
	return oplmi, nil
}

func openTestM2fManagerRO(path_ string, testMarsh bool) (*m2fOplMng, error) {
	oplm, err := MakeM2fManager(path_, 0)
	if err != nil {
		return nil, err
	}
	oplmi, ok := oplm.(*m2fOplMng)
	if !ok {
		return nil, errors.New("not an *m2fOplMng")
	}
	oplmi.testMarsh = testMarsh
	if err = oplm.Open(true); err != nil {
		return nil, err
	}
	return oplmi, nil
}

func TestInventoryCsvExport(t *testing.T) {
	ttd := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(ttd, 100, 3000, 2, 4096, false, 2))
	ctd := t.TempDir()
	err := InventoryCsvExport(ttd, path.Join(ctd, "invDump.csv"), "md5,sha256,sha512")
	require.NoError(t, err)
}
