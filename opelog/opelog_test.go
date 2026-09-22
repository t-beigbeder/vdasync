package opelog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/t-beigbeder/vdasync/internal/common"
)

func TestEnum(t *testing.T) {
	require.Equal(t, "EVT_CREATED", EVT_CREATED.String())
	require.Equal(t, "PRC_CREATING", PRC_CREATING.String())
	require.Equal(t, "HAL_MD5", HAL_MD5.String())
	common.DbgLogger().Debug("TestEnum", "evt", EVT_CREATED, "hal", HAL_MD5, "prc", PRC_CREATING)
}
