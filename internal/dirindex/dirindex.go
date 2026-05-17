package dirindex

import (
	"fmt"
	"io/fs"
	"path"
	"sync"

	"github.com/t-beigbeder/vdasync/dssa"
)

type memDirIndex struct {
	mx      sync.Mutex
	entries map[string]*dssa.DataEntry
	dirs    map[string]map[string]bool
}

func (mdi *memDirIndex) Put(de *dssa.DataEntry) error {
	mdi.mx.Lock()
	defer mdi.mx.Unlock()
	_, ok := mdi.entries[de.Path]
	if !ok {
		mdi.entries[de.Path] = de
		pp := path.Dir(de.Path)
		pde, ok := mdi.entries[pp]
		if !ok || !pde.IsDir {
			return fmt.Errorf("memDirIndex.Put: %s: %s no such directory", de.Path, pp)
		}
		mdi.dirs[pp][de.Path] = true
		if de.IsDir {
			mdi.dirs[de.Path] = map[string]bool{}
		}
	}
	mdi.entries[de.Path] = de

	return nil
}

func (mdi *memDirIndex) Get(path_ string) (*dssa.DataEntry, error) {
	mdi.mx.Lock()
	defer mdi.mx.Unlock()
	de, ok := mdi.entries[path_]
	if !ok {
		err := fs.ErrNotExist
		return &dssa.DataEntry{Path: path_, Error: err, ErrNotExist: true}, err
	}
	return de, nil
}

func (mdi *memDirIndex) List(path_ string) ([]*dssa.DataEntry, error) {
	mdi.mx.Lock()
	defer mdi.mx.Unlock()
	de, ok := mdi.entries[path_]
	if !ok {
		err := fs.ErrNotExist
		return nil, err
	}
	if !de.IsDir {
		err := fmt.Errorf("memDirIndex.List: %s not a directory", path_)
		return nil, err
	}
	var des []*dssa.DataEntry
	for cp := range mdi.dirs[path_] {
		des = append(des, mdi.entries[cp])
	}
	return des, nil
}

func (mdi *memDirIndex) Del(path_ string) error {
	mdi.mx.Lock()
	defer mdi.mx.Unlock()
	de, ok := mdi.entries[path_]
	if !ok {
		return fs.ErrNotExist
	}
	if de.IsDir && len(mdi.dirs[de.Path]) > 0 {
		return fmt.Errorf("memDirIndex.Rm: dir %s is not empty", path_)
	}
	if path_ == "/" {
		return fmt.Errorf("memDirIndex.Rm: removing %s is forbidden", path_)
	}
	if de.IsDir {
		delete(mdi.dirs, path_)
	}
	delete(mdi.entries, path_)
	pp := path.Dir(de.Path)
	delete(mdi.dirs[pp], path_)
	return nil
}

func (mdi *memDirIndex) init() {
	mdi.mx.Lock()
	defer mdi.mx.Unlock()
	mdi.entries["/"] = &dssa.DataEntry{IsDir: true, Path: "/"}
	mdi.dirs["/"] = map[string]bool{}
}

type DirIndex interface {
	Put(*dssa.DataEntry) error
	Get(string) (*dssa.DataEntry, error)
	List(string) ([]*dssa.DataEntry, error)
	Del(string) error
}

// Currently not used, would require actual Dssa as backend
func MakeMemIndexDssa() DirIndex {
	mdi := &memDirIndex{
		entries: map[string]*dssa.DataEntry{},
		dirs:    map[string]map[string]bool{},
	}
	mdi.init()
	return mdi
}
