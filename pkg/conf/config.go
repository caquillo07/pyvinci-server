package conf

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"github.com/caquillo07/pyvinci-server/database"
)

// Config is the application configuration
type Config struct {

	// REST service settings
	REST struct {

		// Port the service will be listening on
		Port int

		// Allow CORS
		AllowCORS bool
	}
	Auth struct {

		// Whether or not tokens will be verified
		Enabled bool

		// Secret used to create tokens
		TokenSecret string
	}

	Database database.Config

	S3 struct {
		ImageBucket string
		AccessKey   string
		SecretKey   string
	}
}

func InitViper(configFile string) {
	if configFile == "" {

		// default to one in present directory named 'config'
		configFile = "config.yaml"
	}
	viper.SetConfigFile(configFile)

	// Default settings
	viper.SetDefault("auth.enabled", true)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv() // read in environment variables that match
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}
	zap.L().Info(fmt.Sprintf("Using config file: %s", viper.ConfigFileUsed()))
}

// LoadConfig will load the configuration from the provided viper instance
func LoadConfig(v *viper.Viper) (*Config, error) {
	config := &Config{}
	if err := v.Unmarshal(config); err != nil {
		return nil, err
	}
	return config, nil
}
