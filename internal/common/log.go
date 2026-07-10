package common

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

func CliLogger(cmd, sll, pathOrKw string) (lgr *slog.Logger, err error) {
	sl := slog.LevelError
	lgr = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: sl}))
	if sll != "" {
		if err = sl.UnmarshalText([]byte(sll)); err != nil {
			return
		}
	}
	var wr io.Writer
	if pathOrKw != "stdout" && pathOrKw != "stderr" {
		if pathOrKw == "" {
			pathOrKw = path.Join(os.TempDir(), fmt.Sprintf("%s-%06d.log", cmd, os.Getpid()))
		}
		wr, err = StdWriter(pathOrKw)
		if err != nil {
			return
		}
	} else {
		wr, _ = StdWriter(pathOrKw)
	}
	lgr = slog.New(slog.NewTextHandler(wr, &slog.HandlerOptions{Level: sl})).With("app", cmd)
	return
}
