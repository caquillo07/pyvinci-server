package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/caquillo07/pyvinci-server/database"
	"github.com/caquillo07/pyvinci-server/pkg/conf"
	"github.com/caquillo07/pyvinci-server/pkg/server"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "server",
		Short: "Run the PyVinci server",
		Run:   runServer,
	})
}

func runServer(cmd *cobra.Command, args []string) {
	config, err := conf.LoadConfig(viper.GetViper())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	db, err := database.Open(config.Database)
	if err != nil {
		log.Fatalln(err)
	}

	log.Fatal(server.NewServer(config, db).Serve())
}
