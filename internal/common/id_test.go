package common

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenId(t *testing.T) {
	td := t.TempDir()
	for i := 0; i < 128; i++ {
		sid, err := GenId()
		require.Nil(t, err)
		require.Equal(t, 64, len(sid))
		path_ := Id2Path(sid)
		require.NotEmpty(t, path_)
		fp := path.Join(td, path_)
		sidAgain := Path2Id(fp)
		require.Equal(t, sid, sidAgain)
	}
}
