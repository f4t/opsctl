package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/f4t/opsctl/instance"
	"github.com/f4t/opsctl/services"
	"github.com/f4t/opsctl/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// toolkitCmd represents the toolkit command
var toolkitCmd = &cobra.Command{
	Use:   "toolkit",
	Short: "toolkit shows services status in CSV format for ITRS",
	Long:  `toolkit shows services status in CSV format for ITRS`,
	Run: func(cmd *cobra.Command, args []string) {
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

	opsctlEnv := utils.LoadOpsctlEnv()
	archivedLogsDirSizeStr := ""
	// + "/" allows DirSizeBytes to follow symlink: data -> /path/to/actual/data + "/"
	archiveSize, err := utils.DirSizeBytes(filepath.Join(opsctlEnv.Home, "archived_logs"))
	if err == nil {
		archivedLogsDirSizeStr = fmt.Sprintf("%d MB", archiveSize/1e6)
	}

	// Build headlines:
	headlines := make(map[string]string)

	headlines["archived_logs"] = archivedLogsDirSizeStr
	headlines["services_home"] = opsctlEnv.Home

	for k, v := range headlines {
		fmt.Printf("<!>%s,%s\n", k, v)
	}

	toolkitRows := make([][]string, 0)
	for _, instance := range instances {
		row := instance.Self().ToolkitRow()
		toolkitRows = append(toolkitRows, row)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"instance", "status", "port", "type", "name", "start_time", "uptime_hours", "pid", "threads", "dir_size", "data_size", "cmdline"})
	for _, v := range toolkitRows {
		table.Append(v)
	}
	table.Render()
}

func init() {
	rootCmd.AddCommand(toolkitCmd)
}
