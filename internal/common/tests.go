package common

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
)

func GetLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func MakeTestFile(tfPath string, size int) error {
	buf := make([]byte, 32*1024)
	fd, err := os.Create(tfPath)
	if err != nil {
		return err
	}
	defer fd.Close()
	bw := len(buf)
	for written := 0; written < size; written += bw {
		if size-written < len(buf) {
			bw = size - written
			buf = make([]byte, bw)
		}
		nr, err := rand.Read(buf)
		if err != nil {
			return err
		}
		nw, err := fd.Write(buf)
		if err != nil {
			return err
		}
		if nw != nr {
			return fmt.Errorf("MakeTestFile: %s written %d != read %d", tfPath, nw, nr)
		}
	}
	return err
}
