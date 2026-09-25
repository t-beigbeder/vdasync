package opelog

import (
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
