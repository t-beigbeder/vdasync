package synpoc

import (
	"testing"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestIt(t *testing.T) {
	log := common.GetLogger()
	gen := data_entry_generator(20)
	pq := make(chan *process_entry, 5)
	rootIsDone := make(chan bool)
	done := func() {
		log.Debug("test is done")
		rootIsDone <- true
	}
	go func() {
		pq <- &process_entry{
			de:   &dssa.DataEntry{Path: "root", IsDir: true},
			done: done,
		}
	}()
	log.Info("TestIt start")
LOOP:
	for {
		select {
		case <-rootIsDone:
			log.Info("main processing, rootIsDone")
			break LOOP
		case pe := <-pq:
			log.Info("main processing, pulling", "name", pe.de.Path[0])
			go process_dnde(log, gen, pq, pe)
		}
	}
	log.Info("TestIt stop")
}
