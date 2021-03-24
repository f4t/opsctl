package utils

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/mitchellh/go-homedir"
)

const (
	dotOpsctlFilename = ".opsctl"
	servicesHomeVar   = "OPSCTL_HOME"
	packageVersionVar = "SERVICE_PACKAGE_VERSION"
)

type OpsctlEnv struct {
	Home string
}

func LoadOpsctlEnv() OpsctlEnv {

	// Find system home dir path
	home, err := homedir.Dir()
	if err != nil {
		log.Fatal("Unable to locate $HOME directory !")
	}

	// Load global opsctl environment
	dotOpsctl := filepath.Join(home, dotOpsctlFilename)
	err = godotenv.Load(dotOpsctl)
	if err != nil {
		log.Fatal(err)
	}

	// Load opsctl home directory env variable
	servicesHome := os.Getenv(servicesHomeVar)
	if servicesHome == "" {
		log.Printf("Error: %s value not set in %s\n", servicesHomeVar, dotOpsctl)
		os.Exit(1)
	}

	// Make sure the path exists
	statRes, err := os.Stat(servicesHome)
	if err != nil {
		log.Printf("Error: %s in %s points to inexisting path=%s", servicesHomeVar, dotOpsctl, servicesHome)
		os.Exit(1)
	}

	// Make sure it points to a directory
	if !statRes.IsDir() {
		log.Printf("Error: %s in %s does not point to a directory path=%s", servicesHomeVar, dotOpsctl, servicesHome)
		os.Exit(1)
	}

	return OpsctlEnv{
		Home: servicesHome,
	}
}
