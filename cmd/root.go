package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	ccmd "ntsc.ac.cn/tas/tas-commons/pkg/cmd"
)

var envs struct {
	certPath   string
	serverName string
}

var rootCmd = &cobra.Command{
	Use:   "ta-time-src",
	Short: "TAS time source monitor",
}

func init() {
	cobra.OnInitialize(func() {})
	viper.AutomaticEnv()
	viper.SetEnvPrefix("TA")
	rootCmd.PersistentFlags().StringVar(&ccmd.GlobalEnvs.LoggerLevel,
		"logger-level", "DEBUG", "logger level")
	rootCmd.PersistentFlags().StringVar(&envs.certPath,
		"cert-path", "/etc/ntsc/ta/certs", "TAS certificates root path")
	rootCmd.PersistentFlags().StringVar(&envs.serverName,
		"server-name", "ntsc.ac.cn", "TAS certificates server name")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
