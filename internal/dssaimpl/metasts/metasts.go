package metasts

import "github.com/t-beigbeder/vdasync/dssa"

type MetaStorageSvc interface {
	NewSession() error
	EndSession() error
	SaveSession() error
	Put(*dssa.DataEntry) error
	Exists(string) (bool, error)
	Get(string) (*dssa.DataEntry, error)
	Del(string) error
	List(string) ([]*dssa.DataEntry, error)
}
