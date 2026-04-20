package common

import (
	"fmt"
	"os"
	"log/slog"
)

func Fatal(log *slog.Logger, err error) {
		log.Error("error", "details", err)
		fmt.Fprintf(os.Stderr, "error %s", err)
		os.Exit(-1)
}
