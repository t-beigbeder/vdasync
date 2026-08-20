package opelogimpl

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/localfiles"
	"github.com/t-beigbeder/vdasync/opelog"
)

func TestM2fOpeLogs(t *testing.T) {
	std := t.TempDir()
	require.NoError(t, common.FileTreeGenerate(std, 250, 25000, 2, 32768, false, 0))
	lgr := common.GetLogger()
	dss := localfiles.MakeLocalFilesDssa()

	ltd := t.TempDir()
	ttd := t.TempDir()

	olm, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), std, ttd)
	require.NoError(t, err)
	require.NoError(t, olm.Init(std, ttd))
	require.NoError(t, olm.NewSession())
	startDe := func(pe *dssa.DssProcessedEntry, noLstatOnList bool) []*dssa.DataEntry {
		des, _ := dss.List(pe.DataEntry.Path)
		de := pe.DataEntry
		children := make([]string, len(des))
		for ic, child := range des {
			children[ic] = path.Base(child.Path)
		}
		olm.PutEntryLog(common.RelPath(de.Path, std),
			&opelog.OpeLogEntry{
				Code:      opelog.OPE_CODE_SOURCE_STAT,
				TimeStamp: pe.DataEntry.Mtime,
				Source: &opelog.StoredEntry{
					Size:          de.Size,
					Mtime:         de.Mtime,
					Uid:           uint32(de.User),
					Gid:           uint32(de.Group),
					Mode:          common.DataEntryMod(de),
					SymLinkTarget: de.SymLinkTarget,
					Children:      children,
				},
				SourceChecksums: "none",
			})
		return des
	}
	startNde := func(pe *dssa.DssProcessedEntry) {
		de := pe.DataEntry
		olm.PutEntryLog(common.RelPath(de.Path, std),
			&opelog.OpeLogEntry{
				Code:      opelog.OPE_CODE_SOURCE_STAT,
				TimeStamp: pe.DataEntry.Mtime,
				Source: &opelog.StoredEntry{
					Size:          de.Size,
					Mtime:         de.Mtime,
					Uid:           uint32(de.User),
					Gid:           uint32(de.Group),
					Mode:          common.DataEntryMod(de),
					SymLinkTarget: de.SymLinkTarget,
				},
				SourceChecksums: "none",
			})
	}
	walker := dssa.MakeWalker(lgr, 4, dss, startDe, startNde,
		nil, nil, nil, nil, "TestBasicWalker", std, dss)
	walker.Run(&dssa.DataEntry{Path: std, IsDir: true})
	require.NoError(t, olm.EndSession())

	olm2, err := MakeM2fManager(path.Join(ltd, "m2f.opl"), std, ttd)
	require.NoError(t, err)
	require.NoError(t, olm2.NewSession())
	le, err := olm2.GetEntryLog("")
	require.NoError(t, err)
	lgr.Debug("TestM2fOpeLogs: done", "path", ltd, "le", len(le.OpeLogEntries))
}
