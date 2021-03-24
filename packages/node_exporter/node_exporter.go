package node_exporter

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/f4t/opsctl/instance"
)

var mandatoryRcVars = []string{
	"NODE_EXPORTER_LISTEN_PORT",
}

type NodeExporter struct {
	Instance instance.Instance
}

func (svc NodeExporter) Self() instance.Instance {
	svc.Instance.Config.MandatoryRcVars = mandatoryRcVars
	return svc.Instance
}

func (svc NodeExporter) Start() {
	instance := svc.Instance
	if !instance.State.Up {
		instance.LogMsg("starting")
		startupGracePeriod := 2 * time.Second
		instance.RunInstanceProcess(startupGracePeriod)
	} else {
		instance.LogMsg("already running")
	}
}

func (svc NodeExporter) Stop() {
	instance := svc.Instance
	if instance.State.Up {
		instance.LogMsg("stopping")
		sigtermGracePeriod := 5 * time.Second
		sigkillGracePeriod := 5 * time.Second
		instance.TerminateInstanceProcess(sigtermGracePeriod, sigkillGracePeriod)
	} else {
		instance.LogMsg("already stopped")
	}
}

// Defines the startup command
func (svc *NodeExporter) SetStartupCmd() {
	instance := svc.Instance
	// Determine path of package binary
	serviceBin := filepath.Join(
		instance.OpsctlEnv.Home,
		"packages",
		instance.Config.Type,
		instance.Config.RcValues["INSTANCE_PACKAGE_VERSION"],
		"node_exporter",
	)

	// Define the command line
	cmdArgs := []string{
		serviceBin,
		fmt.Sprintf(
			"--web.listen-address=:%s",
			instance.Config.RcValues["NODE_EXPORTER_LISTEN_PORT"],
		),
		"--collector.systemd",
	}
	svc.Instance.Config.StartupArgs = cmdArgs
}

// Defines the runtime pattern if different from startup command
// Example: logstash is started with 'bin/logstash' but process runtime is 'bin/java'
// Must return an array of arguments. Regex can be used, example:
// []string{"java.*","something", "--some-option", "value"}
func (svc *NodeExporter) SetRuntimeCmd() {
	svc.Instance.Config.RuntimeArgs = svc.Instance.Config.StartupArgs
}
