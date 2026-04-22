package plugin

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/t-beigbeder/otvl_dtacsy/config"
	"github.com/t-beigbeder/otvl_dtacsy/internal/common"
)

type RunningPlugin struct {
	plugin *config.PluginType
	port   int
	cmd    *exec.Cmd
	err    error
}

func RunConfFile(confPath string) ([]*RunningPlugin, error) {
	config, err := config.Load(confPath)
	if err != nil {
		return nil, err
	}
	rps := []*RunningPlugin{}
	for _, plugin := range config.Plugins {
		crp := RunningPlugin{plugin: plugin, port: plugin.Port}
		if crp.port == 0 {
			port, err := common.GetFreePort()
			if err != nil {
				crp.err = fmt.Errorf("GetFreePort for %s error %s", plugin.Name, err)
				rps = append(rps, &crp)
				continue
			}
			crp.port = port
		}
		args := []string{
			"-port", fmt.Sprint(crp.port),
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
		crp.cmd = cmd
		if err != nil {
			crp.err = fmt.Errorf("child %s start error %s", cmd, err)
		}
		rps = append(rps, &crp)
	}
	return rps, nil
}

func Errors(rps []*RunningPlugin) []error {
	errs := []error{}
	for _, rp := range rps {
		if rp.err != nil {
			errs = append(errs, rp.err)
		}
	}
	return errs
}

func RunConfData(tempFilePath string, conf string) ([]*RunningPlugin, error) {
	if err := common.WriteFile(tempFilePath, []byte(conf)); err != nil {
		return nil, err
	}
	return RunConfFile(tempFilePath)
}

func WaitFor(rps []*RunningPlugin) {
	for _, rp := range rps {
		if rp.cmd == nil || rp.err != nil {
			continue
		}
		err := rp.cmd.Wait()
		if err != nil {
			rp.err = fmt.Errorf("child %s wait error %s", rp.cmd, err)
		}
	}
}
