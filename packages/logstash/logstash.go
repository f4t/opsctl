package logstash

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/f4t/opsctl/instance"
)

var mandatoryRcVars = []string{
	"LOGSTASH_HTTP_API_PORT",
}

type Logstash struct {
	Instance instance.Instance
}

func (svc Logstash) Self() instance.Instance {
	svc.Instance.Config.MandatoryRcVars = mandatoryRcVars
	return svc.Instance
}

func (svc Logstash) Start() {
	instance := svc.Instance
	if !instance.State.Up {
		instance.LogMsg("starting")
		startupGracePeriod := 10 * time.Second
		instance.RunInstanceProcess(startupGracePeriod)
	} else {
		instance.LogMsg("already running")
	}

	// TODO : logstash_exporter sidecar startup
}

func (svc Logstash) Stop() {
	instance := svc.Instance
	if instance.State.Up {
		instance.LogMsg("stopping")
		sigtermGracePeriod := 10 * time.Second
		sigkillGracePeriod := 10 * time.Second
		instance.TerminateInstanceProcess(sigtermGracePeriod, sigkillGracePeriod)
	} else {
		instance.LogMsg("already stopped")
	}
}

func (svc *Logstash) SetStartupCmd() {
	instance := svc.Instance
	// Determine path of package binary
	serviceBin := filepath.Join(
		instance.OpsctlEnv.Home,
		"packages",
		instance.Config.Type,
		instance.Config.RcValues["INSTANCE_PACKAGE_VERSION"],
		"bin",
		"logstash",
	)

	// Define the command line
	cmdArgs := []string{
		serviceBin,
		fmt.Sprintf("--path.config=%s", filepath.Join(instance.Config.Workdir, "logstash.conf")),
		fmt.Sprintf("--path.data=%s", filepath.Join(instance.Config.Workdir, "data")),
		fmt.Sprintf("--path.logs=%s", filepath.Join(instance.Config.Workdir, "logs")),
		"--config.reload.automatic",
		"--http.host=0.0.0.0",
		fmt.Sprintf("--http.port=%s", instance.Config.RcValues["LOGSTASH_HTTP_API_PORT"]),
	}
	svc.Instance.Config.StartupArgs = cmdArgs
}

func (svc *Logstash) SetRuntimeCmd() {
	runtimeArgs := append([]string{"bin/java .*"}, svc.Instance.Config.StartupArgs[1:]...)
	svc.Instance.Config.RuntimeArgs = runtimeArgs
}
