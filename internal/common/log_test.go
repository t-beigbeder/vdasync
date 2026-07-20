package common

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCliLog(t *testing.T) {
	lgr, err := CliLogger("TestCliLog", "DEBUG", "")
	require.Nil(t, err)
	lgr.Debug("1")
	lgr, err = CliLogger("TestCliLog", "DEBUG+2", "")
	require.Nil(t, err)
	lgr.Log(context.Background(), slog.LevelDebug+2, "DETAIL")
	lgr, err = CliLogger("TestCliLog", "INFO", "")
	require.Nil(t, err)
	lgr.Debug("2")
	lgr.Info("2")
	lgr, err = CliLogger("TestCliLog", "ERROR", "")
	require.Nil(t, err)
	lgr.Debug("3")
	lgr.Info("3")
	lgr.Error("3")
	lgr, err = CliLogger("TestCliLog", "", "")
	require.Nil(t, err)
	lgr.Debug("4")
	lgr.Info("4")
	lgr.Error("4")
	lgr, err = CliLogger("TestCliLog", "", "stderr")
	require.Nil(t, err)
	lgr.Error("5")
	lgr, err = CliLogger("TestCliLog", "", "stdout")
	require.Nil(t, err)
	lgr.Error("6")
	fmt.Fprintf(os.Stderr, "the end\n")
}
