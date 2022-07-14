package cmd

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"ntsc.ac.cn/ta-time-source/internal/app"
	ccmd "ntsc.ac.cn/tas/tas-commons/pkg/cmd"
)

var srcEnvs struct {
	cvSerialPath string
	gpSerialPath string
	gbSerialPath string
	rpcListener  string
	wsListener   string
	counterAddr  string
}
var srcCmd = &cobra.Command{
	Use:    "ntsc",
	Short:  "TAS time source monitor",
	PreRun: _src_prerun,
	Run:    _src_run,
}

func init() {
	rootCmd.AddCommand(srcCmd)
	srcCmd.Flags().StringVar(&srcEnvs.cvSerialPath,
		"cv-serial-path", "/dev/ttyS1",
		"common view device serial path")
	srcCmd.Flags().StringVar(&srcEnvs.gpSerialPath,
		"gp-serial-path", "/dev/ttyUSB1",
		"GPS satellite receiver serial path")
	srcCmd.Flags().StringVar(&srcEnvs.gbSerialPath,
		"gb-serial-path", "/dev/ttyUSB0",
		"BD satellite receiver serial path")
	srcCmd.Flags().StringVar(&srcEnvs.rpcListener,
		"rpc-listener", "tcp://0.0.0.0:1358",
		"common view data rpc listener")
	srcCmd.Flags().StringVar(&srcEnvs.wsListener,
		"ws-listener", "0.0.0.0:8788",
		"ws service listener")
	srcCmd.Flags().StringVar(&srcEnvs.counterAddr,
		"counter-addr", "192.168.1.94:45454",
		"multiple channel couner address")
}

func _src_run(cmd *cobra.Command, args []string) {
	s, err := app.NewTimeSourceApp(&app.Config{
		CVConfig: &app.CommonViewDeviceConfig{
			SerialPath: srcEnvs.cvSerialPath,
		},
		SRConfig: &app.SatelliteReceiverConfig{
			GPSSerialPath: srcEnvs.gpSerialPath,
			BDSerialPath:  srcEnvs.gbSerialPath,
		},
		CounerConfig: &app.MultipleCounterConfig{
			Endpoint: srcEnvs.counterAddr,
		},
		CertPath:    envs.certPath,
		RPCListener: srcEnvs.rpcListener,
		WSListener:  srcEnvs.wsListener,
	})
	if err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("failed to create app: %v", err)
	}
	logrus.WithField("prefix", "cmd.root").
		Fatalf("failed to run app: %v", <-s.Start())
}

func _src_prerun(cmd *cobra.Command, args []string) {
	ccmd.InitGlobalVars()
	var err error
	if err = ccmd.ValidateStringVar(&srcEnvs.cvSerialPath,
		"cv_serial_path", true); err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("check boot var failed: %s", err.Error())
	}
	if err = ccmd.ValidateStringVar(&srcEnvs.gpSerialPath,
		"gp_serial_path", true); err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("check boot var failed: %s", err.Error())
	}
	if err = ccmd.ValidateStringVar(&srcEnvs.gbSerialPath,
		"gb_serial_path", true); err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("check boot var failed: %s", err.Error())
	}
	if err = ccmd.ValidateStringVar(&envs.certPath,
		"cert_path", true); err != nil {
		logrus.WithField("prefix", "cmd.root").
			Fatalf("check boot var failed: %s", err.Error())
	}
	go func() {
		ccmd.RunWithSysSignal(nil)
	}()
}
