package s3msts

import (
	"fmt"
	"io/fs"
	"path"
	"sync"
	"time"

	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/dssagrpc"
	"github.com/t-beigbeder/vdasync/internal/common"
	"github.com/t-beigbeder/vdasync/internal/dssaimpl/metasts"
	"github.com/t-beigbeder/vdasync/internal/s3common"
	"google.golang.org/protobuf/proto"
)

type m2s3svc struct {
	rootPrefix string
	s3repo     *s3common.S3RepoClient
	mx         sync.Mutex
	entries    map[string]*dssa.DataEntry
	dirs       map[string]map[string]bool
}

// Del implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) Del(path_ string) error {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	de, ok := msts.entries[path_]
	if !ok {
		return fs.ErrNotExist
	}
	if de.IsDir && len(msts.dirs[de.Path]) > 0 {
		return fmt.Errorf("m2s3svc.Del: dir %s is not empty", path_)
	}
	if path_ == "/" {
		return fmt.Errorf("m2s3svc.Del: removing %s is forbidden", path_)
	}
	if de.IsDir {
		delete(msts.dirs, path_)
	}
	delete(msts.entries, path_)
	pp := path.Dir(de.Path)
	delete(msts.dirs[pp], path_)
	return nil
}

// EndSession implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) EndSession() error {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	var metaEntries dssagrpc.MetaEntries
	entries := map[string]*dssagrpc.DataEntry{}
	metaEntries.Entries = entries
	for p, de := range msts.entries {
		entries[p] = common.DssDte2GrpcDte(de)
	}
	dirs := map[string]*dssagrpc.Paths{}
	metaEntries.Dirs = dirs
	for pp, pcs := range msts.dirs {
		paths := dssagrpc.Paths{}
		for pc := range pcs {
			paths.Paths = append(paths.Paths, pc)
		}
		dirs[pp] = &paths
	}
	bs, err := proto.Marshal(&metaEntries)
	if err != nil {
		return err
	}
	return msts.s3repo.PutObject(path.Join(msts.rootPrefix, "/.vdasync/m2s3msts.meta"), bs)
}

// Exists implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) Exists(path_ string) (bool, error) {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	_, ok := msts.entries[path_]
	return ok, nil
}

// Get implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) Get(path_ string) (*dssa.DataEntry, error) {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	de, ok := msts.entries[path_]
	if !ok {
		return &dssa.DataEntry{Path: path_, Error: fs.ErrNotExist, ErrNotExist: true}, fs.ErrNotExist
	}
	return de, nil
}

// List implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) List(path_ string) ([]*dssa.DataEntry, error) {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	de, ok := msts.entries[path_]
	if !ok {
		return nil, fs.ErrNotExist
	}
	if !de.IsDir {
		return nil, fmt.Errorf("%s: not a directory", path_)
	}
	var cdes []*dssa.DataEntry
	for pc := range msts.dirs[path_] {
		cde, ok := msts.entries[pc]
		if !ok {
			return nil, fmt.Errorf("%s: child %s does not exist", path_, pc)
		}
		cdes = append(cdes, cde)
	}
	return cdes, nil
}

// NewSession implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) NewSession() error {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	key := path.Join(msts.rootPrefix, "/.vdasync/m2s3msts.meta")
	ok, err := msts.s3repo.ObjectExists(key)
	if err != nil {
		return err
	}
	msts.entries = map[string]*dssa.DataEntry{}
	msts.dirs = map[string]map[string]bool{}
	if !ok {
		msts.dirs["/.."] = map[string]bool{}
		msts.dirs["/.."]["/"] = true
		msts.entries["/.."] = &dssa.DataEntry{}
		msts.dirs["/"] = map[string]bool{}
		msts.entries["/"] = &dssa.DataEntry{
			Path:  "/",
			IsDir: true,
			Mtime: time.Now().Unix(),
		}
		return nil
	}
	bs, err := msts.s3repo.GetObject(key)
	var gme dssagrpc.MetaEntries
	if err = proto.Unmarshal(bs, &gme); err != nil {
		return err
	}
	for gp, gde := range gme.Entries {
		msts.entries[gp] = common.GrpcDte2DssDte(gde)
	}
	for gp, gpcs := range gme.Dirs {
		msts.dirs[gp] = map[string]bool{}
		for _, gpc := range gpcs.Paths {
			msts.dirs[gp][gpc] = true
		}
	}
	return nil
}

// Put implements [metasts.MetaStorageSvc].
func (msts *m2s3svc) Put(de *dssa.DataEntry) error {
	msts.mx.Lock()
	defer msts.mx.Unlock()
	pp := path.Dir(de.Path)
	if de.Path == "/" {
		pp = "/.."
	}
	pde, ok := msts.entries[pp]
	if !ok {
		return fmt.Errorf("parent %s for entry %s to be created does not exist", pp, de.Path)
	}
	pde.Mtime = time.Now().Unix()
	msts.dirs[pp][de.Path] = true
	ede, ok := msts.entries[de.Path]
	if ok {
		if ede.IsDir {
			if !de.IsDir {
				return fmt.Errorf("cannot replace existing directory by %s file", de.Path)
			}
		} else if de.IsDir {
			return fmt.Errorf("cannot replace existing file by %s directory", de.Path)
		}
	} else if de.IsDir {
		msts.dirs[de.Path] = map[string]bool{}
	}
	msts.entries[de.Path] = de
	return nil
}

func MakeM2S3MetaStorageSvc(profileName, bucketName, rootPrefix string) (metasts.MetaStorageSvc, error) {
	s3repo, err := s3common.NewS3RepoClient(profileName, bucketName)
	if err != nil {
		return nil, err
	}
	return &m2s3svc{rootPrefix: rootPrefix, s3repo: s3repo}, nil
}
