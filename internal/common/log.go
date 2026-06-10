package common

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path"
)

func CliLogger(cmd, sll, file string) (lgr *slog.Logger, err error) {
	sl := slog.LevelError
	lgr = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: sl}))
	if sll != "" {
		if err = sl.UnmarshalText([]byte(sll)); err != nil {
			return
		}
	}
	var wr io.Writer
	if file != "stderr" {
		if file == "" {
			file = path.Join(os.TempDir(), fmt.Sprintf("%s-%06d.log", cmd, os.Getpid()))
		}
		wr, err = os.Create(file)
		if err != nil {
			return
		}
	} else {
		wr = os.Stderr
	}
	lgr = slog.New(slog.NewTextHandler(wr, &slog.HandlerOptions{Level: sl})).With("app", cmd)
	return
}
