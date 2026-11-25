package common

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	mrand "math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

func doGetLogger(sll string) *slog.Logger {
	sl := slog.Level(-4)
	if sll != "" {
		sl.UnmarshalText([]byte(sll))
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: sl}))
}

func GetLogger() *slog.Logger {
	gtll := os.Getenv("GO_TEST_LOG_LEVEL")
	if gtll == "" {
		return GetNullLogger()
	}
	return doGetLogger(gtll)
}

func DbgLogger() *slog.Logger {
	return doGetLogger("DEBUG")
}

func InfoLogger() *slog.Logger {
	return doGetLogger("INFO")
}

func GetNullLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{}))
}

func GetTestOut() io.Writer {
	so := os.Stdout
	do := io.Discard
	_, _ = so, do
	return do
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

func MakeTextTestFile(tfPath string, size int) error {
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
		for i := 0; i < len(buf); i++ {
			buf[i] = byte(0x41 + (i % 26))
			if i%26 == 25 {
				buf[i] = byte(0x0a)
			}
		}
		nw, err := fd.Write(buf)
		if err != nil {
			return err
		}
		if nw != len(buf) {
			return fmt.Errorf("MakeTestFile: %s written %d != read %d", tfPath, nw, len(buf))
		}
	}
	return err
}

func makeRandomDir(pPath string, depth, maxDirs, maxFiles, filesPerDir, dirsPerDir, maxFileSize int) (int, int, error) {
	sumAddedDirs, sumAddedFiles := 0, 0
	for dx := 0; dx < dirsPerDir && sumAddedDirs < maxDirs && sumAddedFiles < maxFiles; dx++ {
		dn := fmt.Sprintf("d%02d", dx)
		fdn := path.Join(pPath, dn)
		if err := os.Mkdir(fdn, 0750); err != nil {
			return 0, 0, err
		}
		if depth < 3 {
			dpd := dirsPerDir
			if depth == 2 {
				dpd = 0
			}
			addedDirs, addedFiles, err := makeRandomDir(fdn, depth+1, maxDirs, maxFiles, filesPerDir, dpd, maxFileSize)
			if err != nil {
				return 0, 0, err
			}
			sumAddedDirs, sumAddedFiles = sumAddedDirs+addedDirs, sumAddedFiles+addedFiles
		}
		sumAddedDirs += 1
	}
	for fx := 0; fx < filesPerDir && sumAddedFiles < maxFiles; fx++ {
		fn := fmt.Sprintf("f%02d", fx)
		cfs := mrand.Intn(1 + maxFileSize)
		if err := MakeTestFile(path.Join(pPath, fn), cfs); err != nil {
			return 0, 0, err
		}
		sumAddedFiles += 1
	}
	return sumAddedDirs, sumAddedFiles, nil
}

func MakeTestFilesTree(tdPath string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (int, int, error) {
	filesPerDir := maxFiles / maxDirs
	dirsPerDir := childrenPerDir - filesPerDir
	sumAddedDirs, sumAddedFiles, err := makeRandomDir(tdPath, 0, maxDirs, maxFiles, filesPerDir, dirsPerDir, maxFileSize)
	return sumAddedDirs, sumAddedFiles, err
}

func AugmentTestFilesTree(td string) (sumAddedDirs int, sumAddedFiles int, err error) {
	for _, sd := range []string{"dLinks", "dAddFiles", "dMod", "dRO"} {
		if err = os.Mkdir(path.Join(td, sd), 0750); err != nil {
			return
		}
		sumAddedDirs += 1
	}
	dld := path.Join(td, "dLinks")
	if err = os.Symlink("..", path.Join(dld, "dotDot.lnk")); err != nil {
		return
	}
	if err = MakeTestFile(path.Join(dld, "fRef.dat"), 1024); err != nil {
		return
	}
	if err = os.Symlink("fRef.dat", path.Join(dld, "fRef.lnk")); err != nil {
		return
	}
	if err = os.Symlink("notYet.dat", path.Join(dld, "notYet.lnk")); err != nil {
		return
	}
	sumAddedFiles += 4
	dafd := path.Join(td, "dAddFiles")
	for _, sd := range []string{"dStay", "dRemoved"} {
		if err = os.Mkdir(path.Join(dafd, sd), 0750); err != nil {
			return
		}
	}
	sumAddedDirs += 2
	for _, sf := range []string{"fStay.dat", "fRemoved.dat"} {
		if err = MakeTestFile(path.Join(dafd, sf), 1024); err != nil {
			return
		}
	}
	sumAddedFiles += 2
	return
}

func MakeAugmentedTestFilesTree(td string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (sumAddedDirs int, sumAddedFiles int, err error) {
	if sumAddedDirs, sumAddedFiles, err = MakeTestFilesTree(td, maxDirs, maxFiles, childrenPerDir, maxFileSize); err != nil {
		return
	}
	if err = os.Mkdir(path.Join(td, "dau"), 0750); err != nil {
		return
	}
	sumAddedDirs += 1
	addedDirs, addedFiles, err := AugmentTestFilesTree(path.Join(td, "dau"))
	if err != nil {
		return
	}
	sumAddedDirs += addedDirs
	sumAddedFiles += addedFiles
	for _, sd := range []string{"dAddFiles/dRemoved", "dAddFiles/dStay", "dMod", "dRO"} {
		addedDirs, addedFiles, err = AugmentTestFilesTree(path.Join(td, "dau", sd))
		if err != nil {
			return
		}
		sumAddedDirs += addedDirs
		sumAddedFiles += addedFiles
	}
	return
}

func RecChtime(td string, mtime int64) error {
	mt := mtime - 60
	filepath.Walk(td, func(path_ string, info fs.FileInfo, err error) error {
		mt += 60
		return Lutimes(path_, mt)
	})
	return nil
}

func MakeSimpleTestFilesTree(td string) error {
	for _, testd := range []string{"td1", "td2", "td2/tr2", "td3", "td3/ta3", "td4", "td4/tr4", "td4/tr4/trs41"} {
		if err := os.MkdirAll(path.Join(td, testd), 0750); err != nil {
			return err
		}
		for _, tf := range []string{"tf1", "tf2", "tf3", "tf4"} {
			if err := MakeTestFile(path.Join(td, testd, tf), 256); err != nil {
				return err
			}
		}
	}
	if err := RecChtime(td, time.Now().Unix()-27*3600); err != nil {
		return err
	}
	return nil
}

func UpdateSimpleTestsFilesTree(td string) error {
	fi, err := os.Stat(path.Join(td, "td2/tr2"))
	if err != nil {
		return err
	}
	for _, tf := range []string{"tf2"} {
		if err := os.Remove(path.Join(td, "td2/tr2", tf)); err != nil {
			return err
		}
	}
	if err := os.RemoveAll(path.Join(td, "td4/tr4")); err != nil {
		return err
	}
	if err := Lutimes(path.Join(td, "td2/tr2"), fi.ModTime().Unix()); err != nil {
		return err
	}
	for _, tf := range []string{"tfa1", "tfa2"} {
		if err := MakeTestFile(path.Join(td, "td3/ta3", tf), 256); err != nil {
			return err
		}
	}
	return nil
}
