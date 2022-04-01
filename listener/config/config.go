package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
	"time"
)

// environmentPrefix prefix used to avoid environment variable names collisions
const environmentPrefix = "CT_L"

var (
	// Filename configuration file name.
	Filename string
	// App configuration struct
	App AppConfig

	// environmentVarList list of environment variables read by the app. The name should match with a struct field.
	// The dots will be replaced by underscores, it will be capitalized and the environmentPrefix will be added
	// 		i.e.: nats.host => CT_L_NATS_HOST
	environmentVarList = []string{
		"nats.host",
	}
)

// AppConfig struct
type AppConfig struct {
	Grpc       GrpcServer `mapstructure:"grpc"`
	Nats       NatsServer `mapstructure:"nats"`
	Queue      Queue      `mapstructure:"queue"`
	Monitoring Monitoring `mapstructure:"monitor"`
}

// GrpcServer GRPC server configuration
type GrpcServer struct {
	Address string `mapstructure:"address"`
}

// NatsServer NATS server configuration
type NatsServer struct {
	Host                  string `mapstructure:"host"`
	AllowReconnect        bool   `mapstructure:"allow_reconnect"`
	MaxReconnectAttempts  int    `mapstructure:"max_reconnect_attempts"`
	ReconnectWaitDuration string `mapstructure:"reconnect_wait"`
	TimeoutDuration       string `mapstructure:"timeout"`
	ReconnectWait         time.Duration
	Timeout               time.Duration
}

// Queue NATS queue configuration
type Queue struct {
	Topic string `mapstructure:"topic"`
	Group string `mapstructure:"group"`
}

// Monitoring configuration
type Monitoring struct {
	Topic string `mapstructure:"topic"`
}

// Setup bind command flags and environment variables
// The precedence to override a configuration is: flag -> environment variable -> configuration field
func Setup(cmd *cobra.Command, _ []string) error {
	v := viper.New()
	v.SetConfigFile(Filename)
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	v.SetEnvPrefix(environmentPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	err := v.ReadInConfig()
	if err != nil {
		return err
	}

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		_ = v.BindPFlag(flag.Name, cmd.Flags().Lookup(flag.Name))
	})
	for _, env := range environmentVarList {
		_ = v.BindPFlag(env, cmd.Flags().Lookup(env))
	}

	err = v.Unmarshal(&App)
	if err != nil {
		return err
	}
	err = parseDurations()
	if err != nil {
		return err
	}
	return nil
}

// parseDurations parse string durations (1s, 2h, etc.) into time.Duration values
func parseDurations() error {
	var err error
	App.Nats.ReconnectWait, err = time.ParseDuration(App.Nats.ReconnectWaitDuration)
	if err != nil {
		return err
	}
	App.Nats.Timeout, err = time.ParseDuration(App.Nats.TimeoutDuration)
	if err != nil {
		return err
	}

	return nil
}
