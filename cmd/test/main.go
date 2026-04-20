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
		isFatal bool
	)
	flag.BoolVar(&isChild, "is-child", false, "")
	flag.BoolVar(&isFatal, "is-fatal", false, "")
	flag.Parse()
	log := common.GetLogger()
	exe, err := os.Executable()
	if err != nil {
		common.Fatal(log, err)
	}
	if !isChild {
		log = log.With("app", "main")
		log.Info("started", "details", fmt.Sprintf("the executable is %s", exe))
		log.Info("just to say", "details", "let's run a sub-process")
		var cmd *exec.Cmd
		if isFatal {
			cmd = exec.Command(exe, "-is-child", "-is-fatal")
		} else {
			cmd = exec.Command(exe, "-is-child")
		}
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Start()
		if err != nil {
			common.Fatal(log, err)
		}
		err = cmd.Wait()
		if err != nil {
			common.Fatal(log, fmt.Errorf("child wait error %s", err))
		}
	} else {
		log = log.With("app", "child")
		log.Info("started", "details", "here is the child")
		if isFatal {
			common.Fatal(log, fmt.Errorf("child failing on demand"))
		}
	}
}
