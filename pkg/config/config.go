package config

import (
	"github.com/spf13/viper"
)

const (
	Verbose               = "VERBOSE"
	ListeningPort         = "LISTENING_PORT"
	ClairAddress          = "CLAIR_ADDR"
	KlarTrace             = "KLAR_TRACE"
	KlarResultServicePath = "KLAR_RESULT_SERVICE_PATH"
	KlarResultListenPort  = "KLAR_RESULT_LISTEN_PORT"
)

type Config struct {
	Verbose               bool
	WebappPort            string
	ClairAddress          string
	KlarTrace             bool
	KlarResultServicePath string
	KlarResultListenPort  string
}

func setConfigDefaults() {
	viper.SetDefault(Verbose, "false")
	viper.SetDefault(ListeningPort, "8080")
	viper.SetDefault(KlarResultServicePath, "http://kubei.kubei:8081/result/")
	viper.SetDefault(KlarResultListenPort, "8081")
	viper.SetDefault(KlarTrace, "false")                   // Run Klar in more verbose mode
	viper.SetDefault(ClairAddress, "clair.kubei")

	viper.AutomaticEnv()
}

func LoadConfig() *Config {
	setConfigDefaults()

	config := &Config{
		Verbose:               viper.GetBool(Verbose),
		WebappPort:            viper.GetString(ListeningPort),
		KlarResultServicePath: viper.GetString(KlarResultServicePath),
		KlarResultListenPort:  viper.GetString(KlarResultListenPort),
		ClairAddress:          viper.GetString(ClairAddress),
		KlarTrace:             viper.GetBool(KlarTrace),
	}

	return config
}
