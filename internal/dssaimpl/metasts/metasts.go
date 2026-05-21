package metasts

import "github.com/t-beigbeder/vdasync/dssa"

/*
	*List(string) ([]*DataEntry, error) => List
	*Mkdir(*DataEntry) error => Put
	*Stat(string) (*DataEntry, error) => Exists/Get
	*SetStat(_ *DataEntry, noPerm, noMtime bool) error => Put
	*GetReadCloser(string) (io.ReadCloser, error) => none
	*GetWriteCloser(string) (io.WriteCloser, error) => Exists/Get/Put
	*Rm(string) error => Exists/Get/Del
	*Symlink(old, new_ string) error => Exists/Get/Put

	Implementation basic all in one index file

*/

type MetaStorageSvc interface {
	NewSession() error
	EndSession() error
	Put(*dssa.DataEntry) error
	Exists(string) (bool, error)
	Get(string) (*dssa.DataEntry, error)
	Del(string) error
	List(string) ([]*dssa.DataEntry, error)
}
