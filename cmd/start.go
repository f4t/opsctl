package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start (<instance type> <instance name>|all --confirm)",
	Short: "Start service instances.",
	Long: `Start service instances.
Example:

# Start a specific instance
start <instance type> <instance name>

# Start all instances at once
start all --confirm
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "all" {
			if confirm {
				doStartAllInstances()
			} else {
				fmt.Println("--confirm is required when starting all services at once")
				os.Exit(1)
			}
		} else if len(args) == 2 {
			instanceType := args[0]
			instanceName := args[1]
			doStartInstance(instanceType, instanceName)
		} else {
			cmd.Help()
		}
	},
}

func doStartAllInstances() {
	for _, instanceType := range instance.DiscoverInstanceTypes() {
		for _, instanceName := range instance.DiscoverInstances(instanceType) {
			doStartInstance(instanceType, instanceName)
		}
	}
}

func doStartInstance(instanceType string, instanceName string) {
	svc, err := services.MakeInstance(instanceType, instanceName)
	if err != nil {
		log.Fatal(err)
	}
	instance := svc.Self()
	instance.LogMsg("attempting start")
	err = instance.Preflight() // Exit if anything wrong.
	if err != nil {
		return
	}
	svc.Start()
}

func init() {
	rootCmd.AddCommand(startCmd)
	startCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm flag is only required when starting all services at once.")
}
