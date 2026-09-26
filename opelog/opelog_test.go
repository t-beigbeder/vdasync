package opelog

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestEnum(t *testing.T) {
	require.Equal(t, "EVC_CREATED", EVC_CREATED.String())
	require.Equal(t, "STC_DESC_ERROR", STC_DESC_ERROR.String())
	require.Equal(t, "md5", common.HAL_MD5.String())
	common.DbgLogger().Debug("TestEnum", "evt", EVC_CREATED, "hal", common.HAL_MD5, "stc", STC_DESC_ERROR)
}

func TestBackAndForth(t *testing.T) {
	le := NewLogicalEntry()
	h1 := "sha256:0b3b26c3b2e9c20ffa068810ed8badab23d77a423619c698d0ba96787ed83051"
	h2 := "md5:6ced4125b87379848fd3807d129b3dca"
	bss12, err := common.Checksums2TypedChecksums(fmt.Sprintf("%s,%s", h1, h2))
	require.NoError(t, err)
	bss21, err := common.Checksums2TypedChecksums(fmt.Sprintf("%s,%s", h2, h1))
	require.NoError(t, err)
	le.CreateEvent(1, false, EVC_LOADED, &StoredEntry{IsDir: true}, bss12)
	le.CreateEvent(1, true, EVC_LOADED, &StoredEntry{IsDir: true}, bss12)
	le.CreateEvent(1, true, EVC_CREATED, &StoredEntry{IsDir: true}, bss21)
	le.SetState(1, false, STC_DONE_PRESENT, &StoredEntry{Size: 256}, bss12, 0)
	le.SetState(1, true, STC_DONE_PRESENT, &StoredEntry{Size: 256}, bss12, 0)
	pbLe := LogicalEntry2ProtoBuf(le)
	leBack := ProtoBuf2LogicalEntry(pbLe)
	evsSource := leBack.GetEvents(1, false)
	require.Equal(t, 1, len(evsSource))
	require.True(t, evsSource[0].Se.IsDir)
	require.Equal(t, bss12, evsSource[0].Tcss)
	evsTarget := leBack.GetEvents(1, true)
	require.Equal(t, 2, len(evsTarget))
	require.Equal(t, bss21, evsTarget[1].Tcss)
	require.Equal(t, int64(256), leBack.GetState(1, false).Se.Size)
	require.Equal(t, int64(256), leBack.GetState(1, true).Se.Size)
}