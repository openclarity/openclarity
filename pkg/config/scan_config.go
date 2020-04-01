package config

import (
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const (
	MaxParallelism        = "MAX_PARALLELISM"
	TargetNamespace       = "TARGET_NAMESPACE"
	SeverityThreshold     = "SEVERITY_THRESHOLD"
	IgnoreNamespaces      = "IGNORE_NAMESPACES"
	JobResultTimeout      = "JOB_RESULT_TIMEOUT"
	KlarImageName         = "KLAR_IMAGE_NAME"
)

type ScanConfig struct {
	MaxScanParallelism    int
	TargetNamespace       string
	SeverityThreshold     string
	KlarImageName         string
	IgnoredNamespaces     []string
	JobResultTimeout      time.Duration
}

func setScanConfigDefaults() {
	viper.SetDefault(MaxParallelism, "10")
	viper.SetDefault(TargetNamespace, corev1.NamespaceAll) // Scan all namespaces by default
	viper.SetDefault(SeverityThreshold, "MEDIUM")          // Minimum severity level to report
	viper.SetDefault(IgnoreNamespaces, "")
	viper.SetDefault(KlarImageName, "gcr.io/development-infra-208909/klar")
	viper.SetDefault(JobResultTimeout, "10m")

	viper.AutomaticEnv()
}

func LoadScanConfig() *ScanConfig {
	setScanConfigDefaults()

	return &ScanConfig{
		MaxScanParallelism:    viper.GetInt(MaxParallelism),
		TargetNamespace:       viper.GetString(TargetNamespace),
		SeverityThreshold:     viper.GetString(SeverityThreshold),
		IgnoredNamespaces:     strings.Split(viper.GetString(IgnoreNamespaces), ","),
		JobResultTimeout:      viper.GetDuration(JobResultTimeout),
		KlarImageName:         viper.GetString(KlarImageName),
	}
}
