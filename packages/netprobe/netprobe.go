package netprobe

import (
	"path/filepath"
	"time"

	"github.com/f4t/opsctl/instance"
)

var mandatoryRcVars = []string{
	"NETPROBE_LISTEN_PORT",
}

type Netprobe struct {
	Instance instance.Instance
}

func (svc Netprobe) Self() instance.Instance {
	svc.Instance.Config.MandatoryRcVars = mandatoryRcVars
	return svc.Instance
}

func (svc Netprobe) Start() {
	instance := svc.Instance
	if !instance.State.Up {
		instance.LogMsg("starting")
		startupGracePeriod := 2 * time.Second
		instance.RunInstanceProcess(startupGracePeriod)
	} else {
		instance.LogMsg("already running")
	}
}

func (svc Netprobe) Stop() {
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
func (svc *Netprobe) SetStartupCmd() {
	instance := svc.Instance
	// Determine path of package binary
	serviceBin := filepath.Join(
		instance.OpsctlEnv.Home,
		"packages",
		instance.Config.Type,
		instance.Config.RcValues["INSTANCE_PACKAGE_VERSION"],
		"netprobe.linux_64",
	)

	// Define the command line
	cmdArgs := []string{
		serviceBin,
		"-port",
		instance.Config.RcValues["NETPROBE_LISTEN_PORT"],
	}
	svc.Instance.Config.StartupArgs = cmdArgs
}

// Defines the runtime pattern if different from startup command
// Example: logstash is started with 'bin/logstash' but process runtime is 'bin/java'
// Must return an array of arguments. Regex can be used, example:
// []string{"java.*","something", "--some-option", "value"}
func (svc *Netprobe) SetRuntimeCmd() {
	svc.Instance.Config.RuntimeArgs = svc.Instance.Config.StartupArgs
}
