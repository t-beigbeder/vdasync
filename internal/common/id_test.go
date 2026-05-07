package common

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenId(t *testing.T) {
	for i := 0; i < 128; i++ {
		sid, err := GenId()
		require.Nil(t, err)
		require.Equal(t, 64, len(sid))
		path_ := Id2Path(sid)
		require.NotNil(t, path_)
		sidAgain := Path2Id(path_)
		require.Equal(t, sid, sidAgain)
	}
}
