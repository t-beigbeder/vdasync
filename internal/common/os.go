package common

import (
	"fmt"
	"log/slog"
	"os"
)

func Fatal(log *slog.Logger, err error) {
	log.Error("fatal error", "details", err)
	fmt.Fprintf(os.Stderr, "fatal error %s\n", err)
	os.Exit(-1)
}
