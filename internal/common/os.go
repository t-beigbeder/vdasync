package common

import (
	"fmt"
	"log/slog"
	"os"
)

func Fatal(lgr *slog.Logger, err error) {
	if lgr != nil {
		lgr.Error("fatal error", "details", err)
		lgr.Debug("fatal", "args", os.Args)
	}
	fmt.Fprintf(os.Stderr, "fatal error %s\n", err)
	os.Exit(-1)
}
