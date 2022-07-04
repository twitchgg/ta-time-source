package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	ccmd "ntsc.ac.cn/ta-registry/pkg/cmd"
	"ntsc.ac.cn/ta-time-source/internal/service"
)

var serviceEnvs struct {
	tsmAddr string
}

var serviceCmd = &cobra.Command{
	Use:    "service",
	Short:  "TAS time source monitor data service",
	PreRun: _service_prerun,
	Run:    _service_run,
}

func init() {
	rootCmd.AddCommand(serviceCmd)
	serviceCmd.Flags().StringVar(&serviceEnvs.tsmAddr,
		"tsm-addr", "tcp://127.0.0.1:1358",
		"TAS time source monitor gRPC addr")
}

func _service_prerun(cmd *cobra.Command, args []string) {
	ccmd.InitGlobalVars()
	var err error
	if err = ccmd.ValidateStringVar(&serviceEnvs.tsmAddr,
		"tsm_addr", true); err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("check boot var failed: %s", err.Error())
	}
	go func() {
		ccmd.RunWithSysSignal(nil)
	}()
}

func _service_run(cmd *cobra.Command, args []string) {
	ds, err := service.NewDataService(&service.Config{
		ServerName:      envs.serverName,
		CertPath:        envs.certPath,
		ServiceEndpoint: serviceEnvs.tsmAddr,
		CVDataListener:  "127.0.0.1:1123",
	})
	if err != nil {
		logrus.WithField("prefix", "service.main").
			Fatalf("failed to create data service: %v", err)
	}
	logrus.WithField("prefix", "service.main").
		Fatalf("failed to run data service: %v", <-ds.Start())
}
