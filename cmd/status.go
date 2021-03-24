package cmd

import (
	"os"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status [<instance type> <instance name>]",
	Short: "Show instances status.",
	Long: `Show instances status.

# Show a summary of all instances:
opsctl status

# Show detailed summary of a specific instance:
opsctl status <instance type> <instance name>
`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 || (len(args) == 1 && args[0] == "all") {
			statusAll()
		} else if len(args) == 2 {
			instanceStatus(args[0], args[1])
		} else {
			cmd.Help()
		}
	},
}

func instanceStatus(instanceType string, instanceName string) {
	svc, _ := services.MakeInstance(instanceType, instanceName)
	instance := svc.Self()
	if instance.State.Exists {
		instance.PrintSummary()
	} else {
		instance.LogMsg("does not exist")
	}

}

func statusAll() {
	instances := make([]services.ServiceInterface, 0)
	for _, instanceType := range instance.DiscoverInstanceTypes() {
		for _, instanceName := range instance.DiscoverInstances(instanceType) {
			svc, err := services.MakeInstance(instanceType, instanceName)
			if err != nil {
				continue
			}
			instances = append(instances, svc)
		}
	}

	tableData := make([][]string, 0)
	for _, instance := range instances {
		row := instance.Self().StatusRow()
		tableData = append(tableData, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Type", "Name", "Enabled", "State", "PID", "Errors"})
	for _, v := range tableData {
		table.Append(v)
	}
	table.Render()
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
