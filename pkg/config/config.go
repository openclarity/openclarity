package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
	"strings"
)

const (
	Verbose   = "VERBOSE"
	MaxParallelism   = "MAX_PARALLELISM"
	TargetNamespace  = "TARGET_NAMESPACE"
	ClairOutput      = "CLAIR_OUTPUT"
	IgnoreNamespaces = "IGNORE_NAMESPACES"
	KlarTrace        = "KLAR_TRACE"
	WhitelistFile    = "WHITELIST_FILE"
	MyPodNamespace   = "MY-POD-NAMESPACE"
)

type Config struct {
	Verbose            bool
	MaxScanParallelism int
	TargetNamespace    string
	ClairScanThreshold string
	IgnoredNamespaces  []string
	KlarTrace          bool
	WhitelistFilePath  string
	KubeiNamespace     string
}

func convertStringEnvToInt(envName string) (int, error) {
	val, err := strconv.Atoi(viper.GetString(envName))
	if err != nil {
		return 0, fmt.Errorf("failed to convert %s to int. value=%v", envName, viper.GetString(envName))
	}

	return val, nil
}

func setDefaults() {
	viper.SetDefault(Verbose, "false")
	viper.SetDefault(MaxParallelism, "1")
	viper.SetDefault(TargetNamespace, "")   // Scan all namespaces by default
	viper.SetDefault(ClairOutput, "MEDIUM") // Clair severity level scan threshold
	viper.SetDefault(IgnoreNamespaces, "")
	viper.SetDefault(KlarTrace, "false")      // Run Klar in more verbose mode
	viper.SetDefault(WhitelistFile, "")       // No white list file by default
	viper.SetDefault(MyPodNamespace, "kubei") // Default namespace of kubei

	viper.AutomaticEnv()
}

func LoadConfig() (*Config, error) {
	setDefaults()

	var err error
	config := &Config{}

	config.Verbose = viper.GetBool(Verbose)
	config.TargetNamespace = viper.GetString(TargetNamespace)
	config.ClairScanThreshold = viper.GetString(ClairOutput)
	config.IgnoredNamespaces = strings.Split(viper.GetString(IgnoreNamespaces), ",")
	config.KlarTrace = viper.GetBool(KlarTrace)
	config.WhitelistFilePath = viper.GetString(WhitelistFile)
	config.KubeiNamespace = viper.GetString(MyPodNamespace)

	config.MaxScanParallelism, err = convertStringEnvToInt(MaxParallelism)
	if err != nil {
		return nil, err
	}

	return config, nil
}
