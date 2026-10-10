package common

import (
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPeriodicJob(t *testing.T) {
	if os.Getenv("OTVL_TEST_FULL") == "" {
		//t.Skip("OTVL_TEST_FULL not set")
	}
	var (
		lgr *slog.Logger
		err error
	)
	lgr = GetLogger()
	lgr = InfoLogger()
	// lgr, err = CliLogger("TestPeriodicJob", "DEBUG+2", "stderr")
	lgr = DbgLogger()
	require.NoError(t, err)
	pb := NewPeriodicJob(lgr, 1, func() { lgr.Debug("job") })
	go pb.Start()
	go func() {
		time.Sleep(5 * time.Second)
		lgr.Debug("will stop")
		pb.Stop()
	}()
	lgr.Debug("waiting for done")
	<-pb.Done()
	lgr.Debug("done raised")
}
