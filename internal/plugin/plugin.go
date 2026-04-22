package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

func RunConfFile(confPath string) ([]*exec.Cmd, []error) {
	config, err := config.Load(confPath)
	if err != nil {
		return nil, []error{err}
	}
	cmds := []*exec.Cmd{}
	errs := []error{}
	for _, plugin := range config.Plugins {
		port := plugin.Port
		if port == 0 {
			port, err = common.GetFreePort()
			if err != nil {
				errs = append(errs, fmt.Errorf("GetFreePort for %s error %s", plugin.Name, err))
				continue
			}
		}
		args := []string{
			"-port", fmt.Sprint(port),
			"-name", plugin.Name,
			"-type", plugin.Type,
		}
		for _, addArg := range plugin.AddArgs {
			args = append(args, addArg)
		}
		cmd := exec.Command(plugin.ExecutablePath, args...)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err = cmd.Start()
		if err != nil {
			errs = append(errs, fmt.Errorf("child %s start error %s", cmd, err))
			continue
		}
		cmds = append(cmds, cmd)
	}
	return cmds, errs
}

func RunConfData(tempFilePath string, conf string) ([]*exec.Cmd, []error) {
	if err := common.WriteFile(tempFilePath, []byte(conf)); err != nil {
		return nil, []error{err}
	}
	return RunConfFile(tempFilePath)
}

func WaitFor(cmds []*exec.Cmd) []error {
	errs := []error{}
	for _, cmd := range cmds {
		err := cmd.Wait()
		if err != nil {
			errs = append(errs, fmt.Errorf("child %s wait error %s", cmd, err))
			continue
		}
	}
	return errs
}
