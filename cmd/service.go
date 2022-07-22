package cmd

import (
	"github.com/spf13/cobra"
	ccmd "ntsc.ac.cn/tas/tas-commons/pkg/cmd"
)

var serviceCmd = &cobra.Command{
	Use:    "service",
	Short:  "TAS time source monitor data service",
	PreRun: _service_prerun,
	Run:    _service_run,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
}

func _service_prerun(cmd *cobra.Command, args []string) {
	ccmd.InitGlobalVars()
	go func() {
		ccmd.RunWithSysSignal(nil)
	}()
}

func _service_run(cmd *cobra.Command, args []string) {

}
