package sftpc

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"path"
	"time"

	"github.com/pkg/sftp"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type sftpClient struct {
	tokens chan bool
	sfc    *sftp.Client
	root   string
}

// EndSession implements [dssa.Dssa].
func (sf *sftpClient) EndSession() error {
	return nil
}

// GetReadCloser implements [dssa.Dssa].
func (sf *sftpClient) GetReadCloser(path_ string) (io.ReadCloser, error) {
	rr, err := sf.sfc.Open(path.Join(sf.root, path_))
	if err != nil {
		return nil, err
	}
	sf.tokens <- true
	return &sftpReader{reader: rr, cb: func() { <-sf.tokens }}, nil
}

// GetWriteCloser implements [dssa.Dssa].
func (sf *sftpClient) GetWriteCloser(path_ string) (io.WriteCloser, error) {
	wr, err := sf.sfc.Create(path.Join(sf.root, path_))
	if err != nil {
		return nil, err
	}
	sf.tokens <- true
	return &sftpWriter{writer: wr, cb: func() { <-sf.tokens }}, nil
}

// List implements [dssa.Dssa].
func (sf *sftpClient) List(path_ string) ([]*dssa.DataEntry, error) {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	fis, err := sf.sfc.ReadDir(path.Join(sf.root, path_))
	if err != nil {
		return nil, err
	}
	des := []*dssa.DataEntry{}
	for _, fi := range fis {
		sfi, ok := fi.Sys().(*sftp.FileStat)
		if !ok {
			return nil, errors.New("sftpClient.List: not a sftp.FileStat")
		}
		rights := common.Perm2Rights(sfi.FileMode())
		des = append(des, &dssa.DataEntry{
			Path:        fi.Name(),
			IsDir:       fi.IsDir(),
			IsSymLink:   false, // FIXME
			Size:        fi.Size(),
			Mtime:       fi.ModTime().Unix(),
			User:        int(sfi.UID),
			UserRights:  rights[0],
			Group:       int(sfi.GID),
			GroupRights: rights[1],
			OtherRights: rights[2],
		})
	}
	return des, nil
}

// Mkdir implements [dssa.Dssa].
func (sf *sftpClient) Mkdir(de *dssa.DataEntry) error {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	return sf.sfc.Mkdir(path.Join(sf.root, de.Path))
}

// NewSession implements [dssa.Dssa].
func (sf *sftpClient) NewSession() error {
	return nil
}

// Rm implements [dssa.Dssa].
func (sf *sftpClient) Rm(path_ string) error {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	return sf.sfc.Remove(path.Join(sf.root, path_))
}

// SetStat implements [dssa.Dssa].
func (sf *sftpClient) SetStat(de *dssa.DataEntry, noPerm bool, noMtime bool) error {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	fp := path.Join(sf.root, de.Path)
	if !noPerm && !de.IsSymLink {
		ugor := []dssa.Rights{de.UserRights, de.GroupRights, de.OtherRights}
		if err := sf.sfc.Chmod(fp, common.Rights2Mod([3]dssa.Rights(ugor))); err != nil {
			return err
		}
	}
	if !noMtime {
		t := time.Unix(de.Mtime, 0)
		if err := sf.sfc.Chtimes(fp, t, t); err != nil {
			return err
		}
	}
	return nil
}

// Stat implements [dssa.Dssa].
func (sf *sftpClient) Stat(path_ string) (*dssa.DataEntry, error) {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	fp := path.Join(sf.root, path_)
	fi, err := sf.sfc.Lstat(fp)
	if err != nil {
		return &dssa.DataEntry{Path: path_, Error: err, ErrNotExist: os.IsNotExist(err)}, err
	}
	isSymlink := fi.Mode().Type()&fs.ModeSymlink != 0
	var linkTarget string
	if isSymlink {
		linkTarget, err = sf.sfc.ReadLink(fp)
		if err != nil {
			return nil, err
		}
	}
	sfi, ok := fi.Sys().(*sftp.FileStat)
	if !ok {
		return nil, errors.New("sftpClient.Stat: not a sftp.FileStat")
	}
	ugoRights := common.Perm2Rights(sfi.FileMode().Perm())
	return &dssa.DataEntry{
		IsDir:         fi.IsDir(),
		Path:          path_,
		Size:          fi.Size(),
		Mtime:         fi.ModTime().Unix(),
		User:          int(sfi.UID),
		UserRights:    ugoRights[0],
		Group:         int(sfi.GID),
		GroupRights:   ugoRights[1],
		OtherRights:   ugoRights[2],
		IsSymLink:     isSymlink,
		SymLinkTarget: linkTarget,
	}, nil
}

// Symlink implements [dssa.Dssa].
func (sf *sftpClient) Symlink(old string, new_ string) error {
	sf.tokens <- true
	defer func() {
		<-sf.tokens
	}()
	return sf.sfc.Symlink(old, path.Join(sf.root, new_))
}

func MakeSftpClientDssa(sfc *sftp.Client, root string, concurrency int) dssa.Dssa {
	tokens := make(chan bool, concurrency+1)
	return &sftpClient{sfc: sfc, root: root, tokens: tokens}
}
