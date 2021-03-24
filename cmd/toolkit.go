package cmd

import (
	"fmt"
	"os"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// toolkitCmd represents the toolkit command
var toolkitCmd = &cobra.Command{
	Use:   "toolkit",
	Short: "toolkit shows services status in CSV format for ITRS",
	Long:  `toolkit shows services status in CSV format for ITRS`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("toolkit called")
		toolkitAll()
	},
}

func toolkitAll() {
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
		row := instance.Self().ToolkitRow()
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
	rootCmd.AddCommand(toolkitCmd)
}
