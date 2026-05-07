package common

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	mrand "math/rand"
	"os"
	"path"
)

func doGetLogger(sll string) *slog.Logger {
	sl := slog.Level(-4)
	if sll != "" {
		sl.UnmarshalText([]byte(sll))
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: sl}))
}

func GetLogger() *slog.Logger {
	return doGetLogger(os.Getenv("GO_TEST_LOG_LEVEL"))
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

func AugmentTestFilesTree(td string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (sumAddedDirs int, sumAddedFiles int, err error) {
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
	return sumAddedDirs, sumAddedFiles, err
}

func MakeAugmentedTestFilesTree(td string, maxDirs, maxFiles, childrenPerDir, maxFileSize int) (sumAddedDirs int, sumAddedFiles int, err error) {
	if sumAddedDirs, sumAddedFiles, err = MakeTestFilesTree(td, maxDirs, maxFiles, childrenPerDir, maxFileSize); err != nil {
		return
	}
	if err = os.Mkdir(path.Join(td, "dau"), 0750); err != nil {
		return
	}
	sumAddedDirs += 1
	return sumAddedDirs, sumAddedFiles, err
}
