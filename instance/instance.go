package instance

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/f4t/opsctl/utils"
	"github.com/joho/godotenv"
	"github.com/olekukonko/tablewriter"
)

type Instance struct {
	OpsctlEnv utils.OpsctlEnv // Generic
	Config    InstanceConfig
	State     InstanceState
	Errors    InstanceErrors
}

type InstanceConfig struct {
	Type            string   // Generic
	Name            string   // Generic
	Workdir         string   // Generic
	MandatoryRcVars []string // Specific
	RcValues        map[string]string
	PackageVer      string   // Specific
	StartupArgs     []string // Specific
	RuntimeArgs     []string // Specific
}

type InstanceState struct {
	Exists  bool // Generic
	Enabled bool // Generic
	Up      bool // Specific
	PID     int  // Specific
}

type InstanceErrors struct {
	Exists  error
	Enabled error
	Config  error
}

func MakeGenericInstance(Type string, Name string) Instance {
	env := utils.LoadOpsctlEnv()
	config := InstanceConfig{
		Name:    Name,
		Type:    Type,
		Workdir: filepath.Join(env.Home, "instances", Type, Name),
	}
	state := InstanceState{}
	instance := Instance{
		OpsctlEnv: env,
		Config:    config,
		State:     state,
	}

	// Check if workdir exists
	exists, err := instance.WorkdirExists()
	instance.State.Exists = exists
	instance.Errors.Exists = err

	if exists {
		// Check if instance is enabled
		enabled, err := instance.isEnabled()
		instance.State.Enabled = enabled
		instance.Errors.Enabled = err
	}
	return instance
}

func (instance Instance) WorkdirExists() (bool, error) {
	servicesHome := instance.OpsctlEnv.Home
	// First check that a directory exists for instances of this type
	instanceTypeDir := filepath.Join(servicesHome, "instances", instance.Config.Type)
	stat, err := os.Stat(instanceTypeDir)
	if err != nil {
		errMsg := fmt.Sprintf("No such directory %s for instances of type '%s'.", instanceTypeDir, instance.Config.Type)
		return false, errors.New(errMsg)
	}

	if !stat.IsDir() {
		errMsg := fmt.Sprintf("%s exists but is not a directory.", instanceTypeDir)
		return false, errors.New(errMsg)
	}

	// Check the actual instance directory
	instanceDir := instance.Config.Workdir
	stat, err = os.Stat(instanceDir)
	if err != nil {
		errMsg := fmt.Sprintf("Instance directory %s not found.", instanceDir)
		return false, errors.New(errMsg)
	}

	if !stat.IsDir() {
		errMsg := fmt.Sprintf("%s exists but is not a directory.", instanceDir)
		return false, errors.New(errMsg)
	}
	return true, nil
}

func (instance Instance) isEnabled() (bool, error) {
	enabledFlag := filepath.Join(instance.Config.Workdir, "enabled.flag")
	stat, err := os.Stat(enabledFlag)
	if err != nil {
		errMsg := fmt.Sprintf("enabled.flag not found at %s. Please enable instance.", instance.Config.Workdir)
		return false, errors.New(errMsg)
	}

	if stat.IsDir() {
		errMsg := fmt.Sprintf("enabled.flag exists at %s but is a directory !", instance.Config.Workdir)
		return false, errors.New(errMsg)
	}
	return true, nil
}

func (instance Instance) IsUp() (bool, int) {
	pids, _ := utils.GetMatchingPids(instance.Config.RuntimeArgs)
	if len(pids) == 1 {
		return true, pids[0]
	}
	return false, -1
}

const packageVersionVar = "INSTANCE_PACKAGE_VERSION"

func (instance *Instance) LoadRcConfig() error {
	// Unset variables if already present
	// Necessary as godotenv won't override on its own
	os.Unsetenv(packageVersionVar)
	for _, v := range instance.Config.MandatoryRcVars {
		os.Unsetenv(v)
	}

	// Source rc file for instance
	rcFile := filepath.Join(
		instance.Config.Workdir,
		fmt.Sprintf("%s.rc", instance.Config.Type),
	)

	err := godotenv.Load(rcFile)
	if err != nil {
		errMsg := fmt.Sprintf("Unable to load instance config at %s", rcFile)
		return errors.New(errMsg)
	}

	// Load mandatory variables
	vars := make(map[string]string)
	for _, v := range instance.Config.MandatoryRcVars {
		val := os.Getenv(v)
		if val == "" {
			errMsg := fmt.Sprintf("%s definition missing in %s", v, filepath.Join(
				instance.Config.Workdir,
				fmt.Sprintf("%s.rc", instance.Config.Type),
			))
			return errors.New(errMsg)
		}
		vars[v] = val
	}

	// Load package version if not found in rc file
	packageVersion := os.Getenv(packageVersionVar)
	if packageVersion == "" {
		vars[packageVersionVar] = "active_prod"
	} else {
		vars[packageVersionVar] = packageVersion
	}

	instance.Config.RcValues = vars

	return nil
}

func (instance Instance) Preflight() error {
	if !instance.State.Exists {
		instance.LogMsg("Does not exist.")
		os.Exit(1)
	}

	if instance.Errors.Config != nil {
		instance.LogMsg(instance.Errors.Config.Error())
		return instance.Errors.Config
	}

	if !instance.State.Enabled {
		errMsg := "Instance is not enabled"
		instance.LogMsg(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (instance Instance) RunInstanceProcess(startupGracePeriod time.Duration) error {

	if instance.Errors.Config != nil {
		return instance.Errors.Config
	}

	// Define the log path to write on
	logPath := filepath.Join(
		instance.Config.Workdir,
		fmt.Sprintf("%s.log", instance.Config.Type),
	)

	// Run process detached
	err := utils.RunDetachedProcess(logPath, instance.Config.StartupArgs)
	if err != nil {
		return err
	}

	// Wait for the process to be started
	time.Sleep(100 * time.Millisecond)
	pid, err := utils.WaitForProcess(instance.Config.RuntimeArgs, startupGracePeriod)
	if err != nil {
		instance.LogMsg(err.Error())
		return err
	}

	instance.LogMsg(fmt.Sprintf("Started with pid=%d", pid))

	return nil
}

func (instance Instance) TerminateInstanceProcess(sigtermGracePeriod time.Duration, sigkillGracePeriod time.Duration) error {
	pid := instance.State.PID
	syscall.Kill(pid, syscall.SIGTERM)
	// Wait for grace period
	for start := time.Now(); time.Since(start) < sigtermGracePeriod; {
		time.Sleep(50 * time.Millisecond)
		isUp, _ := instance.IsUp()
		if !isUp {
			log.Printf("Terminated pid=%d with SIGTERM", pid)
			return nil
		}
	}
	syscall.Kill(pid, syscall.SIGKILL)
	// Wait for grace period
	for start := time.Now(); time.Since(start) < sigkillGracePeriod; {
		time.Sleep(50 * time.Millisecond)
		isUp, _ := instance.IsUp()
		if !isUp {
			log.Printf("Terminated pid=%d with SIGKILL", pid)
			return nil
		}
	}
	return errors.New("Failed to terminate process within grace period.")
}

func DiscoverInstanceTypes() []string {
	env := utils.LoadOpsctlEnv()
	instancesBase := filepath.Join(env.Home, "instances")
	files, err := ioutil.ReadDir(instancesBase)
	if err != nil {
		log.Fatal(err)
	}
	instanceTypes := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			instanceTypes = append(instanceTypes, f.Name())
		}
	}
	return instanceTypes
}

func DiscoverInstances(instanceType string) []string {
	env := utils.LoadOpsctlEnv()
	instancesBase := filepath.Join(env.Home, "instances", instanceType)
	files, err := ioutil.ReadDir(instancesBase)
	if err != nil {
		log.Fatal(err)
	}
	instances := make([]string, 0)
	for _, f := range files {
		if f.IsDir() {
			instances = append(instances, f.Name())
		}
	}
	return instances
}

func (instance Instance) LogMsg(msg string) {
	log.Println(instance.Desc(), msg)
}

func (instance Instance) Desc() string {
	return fmt.Sprintf("type=%s name=%s", instance.Config.Type, instance.Config.Name)
}

func (instance Instance) ToolkitRow() []string {
	// Status value
	state := ""
	if instance.State.Enabled {
		state = "DOWN"
	}
	if instance.State.Up {
		state = "UP"
	}
	if !instance.State.Enabled {
		state = "DISABLED"
	}

	// Pid value
	pid := ""
	if instance.State.Up {
		pid = fmt.Sprintf("%d", instance.State.PID)
	}
	// Threads, starttime, uptime hours fields (from procfs)
	threads := ""
	startTimeStr := ""
	uptimeHours := ""
	if instance.State.Up {
		stat, err := utils.GetProcStats(instance.State.PID)
		if err == nil {
			threads = fmt.Sprintf("%d", stat.NumThreads)
			startEpochNanos, _ := stat.StartTime()
			startEpoch := int64(startEpochNanos)
			startTime := time.Unix(startEpoch, 0)
			startTimeStr = fmt.Sprintf("%s", startTime)
			uptimeHours = fmt.Sprintf("%d", int(time.Now().Sub(startTime).Hours()))
		}
	}

	instanceDirSizeStr := ""
	instanceSize, err := utils.DirSizeBytes(instance.Config.Workdir)
	if err == nil {
		instanceDirSizeStr = fmt.Sprintf("%d MB", instanceSize/1e6)
	}

	dataDirSizeStr := ""
	// + "/" allows DirSizeBytes to follow symlink: data -> /path/to/actual/data + "/"
	dataSize, err := utils.DirSizeBytes(filepath.Join(instance.Config.Workdir, "data") + "/")
	if err == nil {
		dataDirSizeStr = fmt.Sprintf("%d MB", dataSize/1e6)
	}

	// Build columns:
	row := make([]string, 0)
	// Row name column (instance)
	row = append(row, fmt.Sprintf("%s - %s", instance.Config.Type, instance.Config.Name))
	// Status column
	row = append(row, state)
	// Port column
	row = append(row, "TODO")
	// Instance type column
	row = append(row, instance.Config.Type)
	// Instance name column
	row = append(row, instance.Config.Name)
	// Start time column
	row = append(row, startTimeStr)
	// Uptime hours column
	row = append(row, uptimeHours)
	// PID column
	row = append(row, pid)
	// Threads column
	row = append(row, threads)
	// Dir size column
	row = append(row, instanceDirSizeStr)
	// Data size column
	row = append(row, dataDirSizeStr)
	// cmdline column
	// row = append(row, instance.Config.RuntimeArgs...)
	row = append(row, "---")
	return row
}

func (instance Instance) PrintSummary() {
	tableData := make([][]string, 0)
	tableData = append(tableData, []string{"Type", instance.Config.Type})
	tableData = append(tableData, []string{"Name", instance.Config.Name})
	tableData = append(tableData, []string{"Path", instance.Config.Workdir})
	enabled := "N"
	if instance.State.Enabled {
		enabled = "Y"
	}
	tableData = append(tableData, []string{"Enabled", enabled})
	state := ""
	if instance.State.Enabled {
		state = "DOWN"
	}
	if instance.State.Up {
		state = "UP"
	}
	tableData = append(tableData, []string{"State", state})
	if instance.State.Up {
		pid := fmt.Sprintf("%d", instance.State.PID)
		tableData = append(tableData, []string{"PID", pid})
	}

	if len(instance.Config.RcValues) > 0 {
		tableData = append(tableData, []string{"", ""})

		for k, v := range instance.Config.RcValues {
			rcVal := fmt.Sprintf("%s=%s", k, v)
			tableData = append(tableData, []string{"", rcVal})
		}
	}

	if instance.Errors.Config != nil {
		tableData = append(tableData, []string{"", ""})
		tableData = append(tableData, []string{"Error", instance.Errors.Config.Error()})
	}

	table := tablewriter.NewWriter(os.Stdout)
	for _, v := range tableData {
		table.Append(v)
	}
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.Render()
}

func (instance Instance) StatusRow() []string {
	enabled := "N"
	if instance.State.Enabled {
		enabled = "Y"
	}
	up := ""
	if instance.State.Enabled {
		up = "DOWN"
	}
	if instance.State.Up {
		up = "UP"
	}
	pid := ""
	if instance.State.PID > 0 {
		pid = fmt.Sprintf("%d", instance.State.PID)
	}

	errors := ""

	if instance.Errors.Config != nil {
		errors = "Invalid config"
	}

	return []string{
		instance.Config.Type,
		instance.Config.Name,
		enabled,
		up,
		pid,
		errors,
	}
}
