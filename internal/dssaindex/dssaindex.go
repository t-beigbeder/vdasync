package dssaindex

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
)

type nopWriteCloser struct {
}

// Close implements [io.WriteCloser].
func (n nopWriteCloser) Close() error {
	return nil
}

// Write implements [io.WriteCloser].
func (nopWriteCloser) Write(p []byte) (n int, err error) {
	return len(p), nil
}

type memIndexDssa struct {
	mx      sync.Mutex
	entries map[string]*dssa.DataEntry
	dirs    map[string]map[string]bool
}

// GetReadCloser implements [dssa.Dssa].
func (mid *memIndexDssa) GetReadCloser(path_ string) (io.ReadCloser, error) {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	_, ok := mid.entries[path_]
	if !ok {
		return nil, fs.ErrNotExist
	}
	var buffer bytes.Buffer
	return io.NopCloser(&buffer), nil
}

// GetWriteCloser implements [dssa.Dssa].
func (mid *memIndexDssa) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	de, ok := mid.entries[path_]
	pp := path.Dir(path_)
	if !ok {
		de := dssa.DataEntry{
			IsDir: false,
			Path:  path_,
			Mtime: time.Now().Unix(),
		}
		mid.entries[path_] = &de
		mid.entries[pp].Mtime = de.Mtime
		mid.dirs[pp][path_] = true
	} else {
		de.Mtime = time.Now().Unix()
		mid.entries[pp].Mtime = de.Mtime
	}
	return nopWriteCloser{}, nil
}

// List implements [dssa.Dssa].
func (mid *memIndexDssa) List(path_ string) ([]*dssa.DataEntry, error) {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	de, ok := mid.entries[path_]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if !de.IsDir {
		return nil, fmt.Errorf("memIndexDssa.List: %s: not a directory", path_)
	}
	var des []*dssa.DataEntry
	for cp := range mid.dirs[path_] {
		des = append(des, mid.entries[cp])
	}
	return des, nil
}

// Mkdir implements [dssa.Dssa].
func (mid *memIndexDssa) Mkdir(de *dssa.DataEntry) error {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	pp := path.Dir(de.Path)
	_, ok := mid.entries[de.Path]
	if ok {
		return fmt.Errorf("memIndexDssa.Mkdir: %s: already created", de.Path)
	}
	if !de.IsDir {
		return fmt.Errorf("memIndexDssa.Mkdir: %s: IsDir should be set", de.Path)
	}
	pde, ok := mid.entries[pp]
	if !ok {
		return fmt.Errorf("memIndexDssa.Mkdir: %s: no such file or directory", pp)
	}
	mid.entries[de.Path] = de
	mid.dirs[de.Path] = map[string]bool{}
	mid.dirs[pp][de.Path] = true
	pde.Mtime = time.Now().Unix()
	return nil
}

// Rm implements [dssa.Dssa].
func (mid *memIndexDssa) Rm(path_ string) error {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	de, ok := mid.entries[path_]
	if !ok {
		return fs.ErrNotExist
	}
	if de.IsDir && len(mid.dirs[de.Path]) > 0 {
		return fmt.Errorf("memIndexDssa.Rm: dir %s is not empty", path_)
	}
	if path_ == "/" {
		return fmt.Errorf("memIndexDssa.Rm: removing %s is forbidden", path_)
	}
	if de.IsDir {
		delete(mid.dirs, path_)
	}
	delete(mid.entries, path_)
	pp := path.Dir(de.Path)
	delete(mid.dirs[pp], path_)
	return nil
}

// SetStat implements [dssa.Dssa].
func (mid *memIndexDssa) SetStat(de *dssa.DataEntry, _ bool, _ bool) error {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	_, ok := mid.entries[de.Path]
	if !ok {
		return fs.ErrNotExist
	}
	mid.entries[de.Path] = de
	return nil
}

// Stat implements [dssa.Dssa].
func (mid *memIndexDssa) Stat(path_ string) (*dssa.DataEntry, error) {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	de, ok := mid.entries[path_]
	if !ok {
		err := fs.ErrNotExist
		return &dssa.DataEntry{Path: path_, Error: err, ErrNotExist: true}, err
	}
	return de, nil
}

// Symlink implements [dssa.Dssa].
func (mid *memIndexDssa) Symlink(old string, new_ string) error {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	pp := path.Dir(new_)
	_, ok := mid.entries[new_]
	if ok {
		return fmt.Errorf("memIndexDssa.Symlink: %s: already created", new_)
	}
	pde, ok := mid.entries[pp]
	if !ok {
		return fmt.Errorf("memIndexDssa.Symlink: %s: no such file or directory", pp)
	}
	mid.entries[new_] = &dssa.DataEntry{
		IsSymLink:     true,
		SymLinkTarget: old,
		Mtime:         time.Now().Unix(),
	}
	mid.dirs[pp][new_] = true
	pde.Mtime = time.Now().Unix()
	return nil
}

func (mid *memIndexDssa) init() {
	mid.mx.Lock()
	defer mid.mx.Unlock()
	mid.entries["/"] = &dssa.DataEntry{IsDir: true, Path: "/"}
	mid.dirs["/"] = map[string]bool{}
}

// Currently not used, would require actual Dssa as backend
func MakeMemIndexDssa() dssa.Dssa {
	mid := &memIndexDssa{
		entries: map[string]*dssa.DataEntry{},
		dirs:    map[string]map[string]bool{},
	}
	mid.init()
	return mid
}
