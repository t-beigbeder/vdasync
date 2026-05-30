package sftpc

import (
	"errors"
	"io"
	"path"

	"github.com/pkg/sftp"
	"github.com/t-beigbeder/vdasync/dssa"
	"github.com/t-beigbeder/vdasync/internal/common"
)

type sftpClient struct {
	sfc  *sftp.Client
	root string
}

// EndSession implements [dssa.Dssa].
func (sf *sftpClient) EndSession() error {
	panic("unimplemented")
}

// GetReadCloser implements [dssa.Dssa].
func (sf *sftpClient) GetReadCloser(string) (io.ReadCloser, error) {
	panic("unimplemented")
}

// GetWriteCloser implements [dssa.Dssa].
func (sf *sftpClient) GetWriteCloser(string) (io.WriteCloser, error) {
	panic("unimplemented")
}

// List implements [dssa.Dssa].
func (sf *sftpClient) List(path_ string) ([]*dssa.DataEntry, error) {
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
		rights := common.Mod2Rights(sfi.FileMode())
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
	return sf.sfc.Mkdir(path.Join(sf.root, de.Path))
}

// NewSession implements [dssa.Dssa].
func (sf *sftpClient) NewSession() error {
	panic("unimplemented")
}

// Rm implements [dssa.Dssa].
func (sf *sftpClient) Rm(path_ string) error {
	return sf.sfc.Remove(path.Join(sf.root, path_))
}

// SetStat implements [dssa.Dssa].
func (sf *sftpClient) SetStat(_ *dssa.DataEntry, noPerm bool, noMtime bool) error {
	panic("unimplemented")
}

// Stat implements [dssa.Dssa].
func (sf *sftpClient) Stat(string) (*dssa.DataEntry, error) {
	panic("unimplemented")
}

// Symlink implements [dssa.Dssa].
func (sf *sftpClient) Symlink(old string, new_ string) error {
	panic("unimplemented")
}

func MakeSftpClientDssa(sfc *sftp.Client, root string) dssa.Dssa {
	return &sftpClient{sfc: sfc, root: root}
}
