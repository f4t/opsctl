package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop (<instance type> <instance name>|all --confirm)",
	Short: "Stop service instances.",
	Long: `Stop service instances.
Example:

# Stop a specific instance
stop <instance type> <instance name>

# Stop all instances at once
stop all --confirm
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "all" {
			if confirm {
				doStopAllInstances()
			} else {
				fmt.Println("--confirm is required when stopping all services at once")
				os.Exit(1)
			}
		} else if len(args) == 2 {
			instanceType := args[0]
			instanceName := args[1]
			doStopInstance(instanceType, instanceName)
		} else {
			cmd.Help()
		}
	},
}

func doStopAllInstances() {
	for _, instanceType := range instance.DiscoverInstanceTypes() {
		for _, instanceName := range instance.DiscoverInstances(instanceType) {
			doStopInstance(instanceType, instanceName)
		}
	}
}

func doStopInstance(instanceType string, instanceName string) {
	svc, err := services.MakeInstance(instanceType, instanceName)
	if err != nil {
		log.Fatal(err)
	}
	instance := svc.Self()
	instance.LogMsg("attempting stop")
	instance.Preflight() // Exit if anything wrong.
	svc.Stop()
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm flag is only required when stopping all services at once.")
}
