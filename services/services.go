package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/f4t/opsctl/instance"

	// All packages modules need to be imported here:
	"github.com/f4t/opsctl/packages/logstash"
	"github.com/f4t/opsctl/packages/netprobe"
	"github.com/f4t/opsctl/packages/node_exporter"
)

type ServiceInterface interface {
	Self() instance.Instance
	Start()
	Stop()
	SetStartupCmd()
	SetRuntimeCmd()
}

// All package mappings need to be implemented here:
func packageSelector(instance instance.Instance) (ServiceInterface, error) {
	switch instance.Config.Type {
	case "netprobe":
		return &netprobe.Netprobe{Instance: instance}, nil
	case "logstash":
		return &logstash.Logstash{Instance: instance}, nil
	case "node_exporter":
		return &node_exporter.NodeExporter{Instance: instance}, nil
	default:
		err := fmt.Sprintf("Unsupported instance type '%s'", instance.Config.Type)
		return nil, errors.New(err)
	}
}

// MakeInstance returns a fully conigured / inspected instance struct
func MakeInstance(Type string, Name string) (ServiceInterface, error) {
	// Initialize generic instance data (name, type, exists, etc..)
	instance := instance.MakeGenericInstance(Type, Name)
	svc, err := loadSpecifics(instance)
	if err != nil {
		return nil, err
	}
	instance = svc.Self() // Reflect to get instance details
	svc = inspectRuntime(instance)
	return svc, nil
}

func loadSpecifics(instance instance.Instance) (ServiceInterface, error) {
	// Create package-specific instance
	svc, err := packageSelector(instance)
	if err != nil {
		return nil, err
	}
	// Reflect to get editable instance details
	instance = svc.Self()
	// Populate package-specifics instance data
	// Load RC file
	err = instance.LoadRcConfig()
	instance.Errors.Config = err
	svc, _ = packageSelector(instance)

	// Set Startup and Runtime command
	svc.SetStartupCmd()
	svc.SetRuntimeCmd()

	// Re-build instance with all specifics populated
	instance = svc.Self() // Reflect to get instance details
	svc, _ = packageSelector(instance)
	return svc, nil
}

func inspectRuntime(instance instance.Instance) ServiceInterface {
	// Populate runtime checks
	isUp, pid := instance.IsUp()
	instance.State.Up = isUp
	instance.State.PID = pid
	svc, err := packageSelector(instance)
	if err != nil {
		log.Printf(err.Error())
		os.Exit(1)
	}
	return svc
}
