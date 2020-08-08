package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/caquillo07/pyvinci-server/pkg/conf"
)

var configFile string

// rootCmd represents the base command when called without any sub-commands
var rootCmd = &cobra.Command{
	Use:   "pyvinci-server",
	Short: "Server in charge of managing all things pyvinci related",
}

func init() {
	cobra.OnInitialize(func() { conf.InitViper(configFile) })
	cobra.OnInitialize(initLogging)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(
		&configFile,
		"config",
		"",
		"config file (default is $HOME/.config.yaml)",
	)
	rootCmd.PersistentFlags().Bool("dev-log", false, "Development logging")
}

// Execute adds all child commands to the root command and sets flags
// appropriately. This is called by main.main(). It only needs to happen once
// to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// initLogging will initialize the global zap logger
func initLogging() {
	var logger *zap.Logger
	if val, _ := rootCmd.PersistentFlags().GetBool("dev-log"); val == true {
		logger, _ = zap.NewDevelopment()
		logger.Info("Development logging enabled")
	} else {
		logger, _ = zap.NewProduction()
	}

	logger.Info("Server started")

	zap.ReplaceGlobals(logger)
}
