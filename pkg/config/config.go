package config

import (
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const (
	Verbose               = "VERBOSE"
	MaxParallelism        = "MAX_PARALLELISM"
	TargetNamespace       = "TARGET_NAMESPACE"
	SeverityThreshold     = "SEVERITY_THRESHOLD"
	ClairAddress          = "CLAIR_ADDR"
	IgnoreNamespaces      = "IGNORE_NAMESPACES"
	KlarTrace             = "KLAR_TRACE"
	KlarImageName         = "KLAR_IMAGE_NAME"
	KlarResultServicePath = "KLAR_RESULT_SERVICE_PATH"
	KlarResultListenPort  = "KLAR_RESULT_LISTEN_PORT"
	JobResultTimeout      = "JOB_RESULT_TIMEOUT"
	ListeningPort         = "LISTENING_PORT"
)

type Config struct {
	Verbose               bool
	MaxScanParallelism    int
	TargetNamespace       string
	SeverityThreshold     string
	ClairAddress          string
	IgnoredNamespaces     []string
	KlarTrace             bool
	KlarImageName         string
	KlarResultServicePath string
	KlarResultListenPort  string
	JobResultTimeout      time.Duration
	WebappPort            string
}

func setDefaults() {
	viper.SetDefault(Verbose, "false")
	viper.SetDefault(MaxParallelism, "10")
	viper.SetDefault(TargetNamespace, corev1.NamespaceAll) // Scan all namespaces by default
	viper.SetDefault(SeverityThreshold, "MEDIUM")          // Minimum severity level to report
	viper.SetDefault(ClairAddress, "clair.kubei")
	viper.SetDefault(IgnoreNamespaces, "")
	viper.SetDefault(KlarTrace, "false")                   // Run Klar in more verbose mode
	viper.SetDefault(KlarImageName, "gcr.io/development-infra-208909/klar")
	viper.SetDefault(KlarResultServicePath, "http://kubei.kubei:8081/result/")
	viper.SetDefault(KlarResultListenPort, "8081")
	viper.SetDefault(JobResultTimeout, "10m")
	viper.SetDefault(ListeningPort, "8080")

	viper.AutomaticEnv()
}

func LoadConfig() (*Config, error) {
	setDefaults()

	config := &Config{
		Verbose:               viper.GetBool(Verbose),
		MaxScanParallelism:    viper.GetInt(MaxParallelism),
		TargetNamespace:       viper.GetString(TargetNamespace),
		SeverityThreshold:     viper.GetString(SeverityThreshold),
		ClairAddress:          viper.GetString(ClairAddress),
		IgnoredNamespaces:     strings.Split(viper.GetString(IgnoreNamespaces), ","),
		KlarTrace:             viper.GetBool(KlarTrace),
		KlarImageName:         viper.GetString(KlarImageName),
		KlarResultServicePath: viper.GetString(KlarResultServicePath),
		KlarResultListenPort:  viper.GetString(KlarResultListenPort),
		JobResultTimeout:      viper.GetDuration(JobResultTimeout),
		WebappPort:            viper.GetString(ListeningPort),
	}

	return config, nil
}
