package synpoc

import (
	"fmt"
	"testing"

	"github.com/t-beigbeder/otvl_dtacsy/dssa"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func TestIt(t *testing.T) {
	log := common.GetLogger()
	gen := data_entry_generator(100)
	processing_list := make(chan *dssa.DataEntry, 7)
	done := func() {
		log.Debug("test is done")
	}
	
	process_dir(log, gen, processing_list, done, &dssa.DataEntry{Name: "root", IsDir: true})
}
