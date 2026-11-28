package opelogimpl

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/t-beigbeder/vdasync/internal/common"
)

type bckFile struct {
	bckDir string
	fileCount int
	bf *os.File
}

func (bf *bckFile) create() error {
	var err error
	if bf.bf != nil {
		if err = bf.close(); err != nil {
			return err
		}
	}
	bf.bf, err = os.Create(path.Join(bf.bckDir, fmt.Sprintf("bck-%03d.pbfl", bf.fileCount)))
	if err != nil {
		return err
	}
	bf.fileCount++
	return nil
}

func (bf *bckFile) close() error {
	if bf.bf == nil {
		return errors.New("bckFile.close: not opened")
	}
	err := bf.bf.Close()
	bf.bf = nil
	return err
}

// Backup implements [PbfManager].
func (m *pbOplMng) Backup(bckDir string, partSize int) error {
	if !m.isOpen {
		return fmt.Errorf("pbOplMng.Backup: %s source is not opened", m.path_)
	}
	bckFi, err := os.Stat(bckDir)
	if err != nil || !bckFi.IsDir() {
		return fmt.Errorf("pbOplMng.Backup: %s directory does not exist", bckDir)
	}
	bckMf := path.Join(bckDir, path.Base(m.path_))
	if common.FileExists(bckMf) {
		return fmt.Errorf("pbOplMng.Backup: %s already exists", bckMf)
	}
	if err := m.saveMasterTo(bckMf); err != nil {
		return err
	}
	bf := &bckFile{bckDir: bckDir}
	if err = bf.create(); err != nil {
		return err
	}
	defer bf.close()

	var (
		hasErrors bool
		curPartSize int
	)

	err = m.doWalk(func(path_ string) error {
		bole, err := common.UnsafeLoadFile(path_)
		if err != nil {
			m.lgr.Error("pbOplMng.Backup", "path", path_, "err", err)
			hasErrors = true
			return nil
		}
		bs4 := make([]byte, 4)
		binary.BigEndian.PutUint32(bs4, uint32(len(bole)))
		if _, err = bf.bf.Write(bs4); err != nil {
			return err
		}
		if _, err = bf.bf.Write(bole); err != nil {
			return err
		}
		curPartSize++
		if partSize > 0 && curPartSize >= partSize {
			if err = bf.create(); err != nil {
				return err
			}
			curPartSize = 0
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("pbOplMng.Backup: internal error: %s", err)
	}
	if err = bf.close(); err != nil {
		return fmt.Errorf("pbOplMng.Backup: internal error: %s", err)
	}
	if hasErrors {
		return errors.New("pbOplMng.Backup: some entries have errors, please consult the logs")
	}
	return nil
}

// Restore implements [PbfManager].
func (m *pbOplMng) Restore(path_ string) error {
	panic("unimplemented")
}
