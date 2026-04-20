package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func main() {
	var (
		isChild bool
	)
	flag.BoolVar(&isChild, "is-child", false, "")
	flag.Parse()
	log := common.GetLogger()
	exe, err := os.Executable()
	if err != nil {
		log.Error("error", "details", err)
		fmt.Fprintf(os.Stderr, "error %s", err)
		os.Exit(-1)
	}
	if !isChild {
		log.Info("main", "details", fmt.Sprintf("the executable is %s", exe))
		common.GetLogger().Info("main", "details", "let's run a sub-process")
		cmd := exec.Command(exe, "-is-child")
		cmd.Start()
		cmd.Wait()
	} else {
		log.Info("main", "details", "here is the child")
	}
	fmt.Fprintf(os.Stderr, "this is stderr\n")
}
