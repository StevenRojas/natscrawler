package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
	"time"
)

// environmentPrefix prefix used to avoid environment variable names collisions
const environmentPrefix = "CT_R"

var (
	// Filename configuration file name.
	Filename string
	// App configuration struct
	App AppConfig

	// environmentVarList list of environment variables read by the app. The name should match with a struct field.
	// The dots will be replaced by underscores, it will be capitalized and the environmentPrefix will be added
	// 		i.e.: nats.host => CT_R_NATS_HOST
	environmentVarList = []string{
		"nats.host",
	}
)

// AppConfig struct
type AppConfig struct {
	General General    `mapstructure:"general"`
	Grpc    GrpcServer `mapstructure:"grpc"`
}

// General configuration
type General struct {
	SkipRows   int    `mapstructure:"skip_rows"`
	BufferSize int    `mapstructure:"buffer_size"`
	URLDomain  string `mapstructure:"url_domain"`
}

// GrpcServer GRPC server configuration
type GrpcServer struct {
	ListenerAddress        string `mapstructure:"listener_address"`
	DialTimeoutDuration    string `mapstructure:"dial_timeout"`
	RequestTimeoutDuration string `mapstructure:"request_timeout"`
	PingIntervalDuration   string `mapstructure:"ping_interval"`
	PingTimeoutDuration    string `mapstructure:"ping_timeout"`
	DialTimeout            time.Duration
	RequestTimeout         time.Duration
	PingInterval           time.Duration
	PingTimeout            time.Duration
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
	App.Grpc.DialTimeout, err = time.ParseDuration(App.Grpc.DialTimeoutDuration)
	if err != nil {
		return err
	}
	App.Grpc.RequestTimeout, err = time.ParseDuration(App.Grpc.RequestTimeoutDuration)
	if err != nil {
		return err
	}
	App.Grpc.PingTimeout, err = time.ParseDuration(App.Grpc.PingTimeoutDuration)
	if err != nil {
		return err
	}
	App.Grpc.PingInterval, err = time.ParseDuration(App.Grpc.PingIntervalDuration)
	if err != nil {
		return err
	}
	return nil
}
