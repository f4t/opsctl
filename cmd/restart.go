package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/spf13/cobra"
)

// restartCmd represents the restart command
var restartCmd = &cobra.Command{
	Use:   "restart (<instance type> <instance name>|all --confirm)",
	Short: "Restart service instances.",
	Long: `Restart service instances.
Example:

# Restart a specific instance
restart <instance type> <instance name>

# Restart all instances at once
restart all --confirm
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 1 && args[0] == "all" {
			if confirm {
				doRestartAllInstances()
			} else {
				fmt.Println("--confirm is required when restartping all services at once")
				os.Exit(1)
			}
		} else if len(args) == 2 {
			instanceType := args[0]
			instanceName := args[1]
			doRestartInstance(instanceType, instanceName)
		} else {
			cmd.Help()
		}
	},
}

func doRestartAllInstances() {
	for _, instanceType := range instance.DiscoverInstanceTypes() {
		for _, instanceName := range instance.DiscoverInstances(instanceType) {
			doRestartInstance(instanceType, instanceName)
		}
	}
}

func doRestartInstance(instanceType string, instanceName string) {
	svc, err := services.MakeInstance(instanceType, instanceName)
	if err != nil {
		log.Fatal(err)
	}
	instance := svc.Self()
	instance.LogMsg("attempting restart")
	err = instance.Preflight()
	svc.Stop()
	// Re-load state
	svc, _ = services.MakeInstance(instanceType, instanceName)
	if err != nil {
		return
	}
	svc.Start()
}

func init() {
	rootCmd.AddCommand(restartCmd)
	restartCmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm flag is only required when restarting all services at once.")
}
