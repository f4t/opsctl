package utils

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func GetMatchingPids(cmdArgs []string) ([]int, error) {
	cmdString := strings.Join(cmdArgs, " ")
	psString := fmt.Sprintf("ps aux | grep -v grep | grep -e '%s' | awk '{print $2}'", cmdString)
	psOutput, _ := exec.Command("sh", "-c", psString).Output()
	pids := make([]int, 0)
	for _, pid := range strings.Split(string(psOutput), "\n") {
		if pid != "" {
			pidVal, err := strconv.Atoi(pid)
			if err != nil {
				continue
			}
			pids = append(pids, pidVal)
		}
	}
	return pids, nil
}

func RunDetachedProcess(logPath string, cmdArgs []string) error {
	// Check that executable exists
	executable, err := os.Stat(cmdArgs[0])
	if err != nil {
		log.Printf("Package binary %s not found", cmdArgs[0])
		return err
	}

	if executable.IsDir() {
		errMsg := fmt.Sprintf("%s is not a file.", cmdArgs[0])
		log.Printf(errMsg)
		return errors.New(errMsg)
	}

	// Open log file (append if exists else create)
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		errMsg := fmt.Sprintf("Failed creating log file %s", logPath)
		log.Printf(errMsg)
		return errors.New(errMsg)
	}
	// Create command
	cmdR := exec.Command("nohup", cmdArgs...)
	// Redirect both stdout and stderr to the log file
	cmdR.Stdout = f
	cmdR.Stderr = f
	// Run the process
	err = cmdR.Start()
	if err != nil {
		log.Printf("Failed starting.")
		return err
	}
	return nil
}

func WaitForProcess(cmdArgs []string, gracePeriod time.Duration) (int, error) {
	for start := time.Now(); time.Since(start) < gracePeriod; {
		pids, _ := GetMatchingPids(cmdArgs)
		if len(pids) == 1 {
			return pids[0], nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return -1, errors.New("Instance hasn't started within grace period.")
}
