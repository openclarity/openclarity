package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"strings"
	"time"
)

const (
	MaxParallelism       = "MAX_PARALLELISM"
	TargetNamespace      = "TARGET_NAMESPACE"
	SeverityThreshold    = "SEVERITY_THRESHOLD"
	IgnoreNamespaces     = "IGNORE_NAMESPACES"
	JobResultTimeout     = "JOB_RESULT_TIMEOUT"
	KlarImageName        = "KLAR_IMAGE_NAME"
	DeleteJobPolicy      = "DELETE_JOB_POLICY"
	ShouldScanDockerFile = "SHOULD_SCAN_DOCKERFILE"
)

type ScanConfig struct {
	MaxScanParallelism   int
	TargetNamespace      string
	SeverityThreshold    string
	KlarImageName        string
	IgnoredNamespaces    []string
	JobResultTimeout     time.Duration
	DeleteJobPolicy      DeleteJobPolicyType
	ShouldScanDockerFile bool
}

func setScanConfigDefaults() {
	viper.SetDefault(MaxParallelism, "10")
	viper.SetDefault(TargetNamespace, corev1.NamespaceAll) // Scan all namespaces by default
	viper.SetDefault(SeverityThreshold, "MEDIUM")          // Minimum severity level to report
	viper.SetDefault(IgnoreNamespaces, "")
	viper.SetDefault(KlarImageName, "gcr.io/development-infra-208909/klar")
	viper.SetDefault(JobResultTimeout, "10m")
	viper.SetDefault(DeleteJobPolicy, DeleteJobPolicySuccessful)
	viper.SetDefault(ShouldScanDockerFile, "true")

	viper.AutomaticEnv()
}

func LoadScanConfig() *ScanConfig {
	setScanConfigDefaults()

	return &ScanConfig{
		MaxScanParallelism: viper.GetInt(MaxParallelism),
		TargetNamespace:    viper.GetString(TargetNamespace),
		SeverityThreshold:  viper.GetString(SeverityThreshold),
		KlarImageName:      viper.GetString(KlarImageName),
		IgnoredNamespaces:  strings.Split(viper.GetString(IgnoreNamespaces), ","),
		JobResultTimeout:   viper.GetDuration(JobResultTimeout),
		DeleteJobPolicy:    getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
		ShouldScanDockerFile: viper.GetBool(ShouldScanDockerFile),
	}
}

func getDeleteJobPolicyType(policyType string) DeleteJobPolicyType {
	deleteJobPolicy := DeleteJobPolicyType(policyType)
	if !deleteJobPolicy.IsValid() {
		log.Warnf("Invalid %s type - using default `%s`", DeleteJobPolicy, DeleteJobPolicySuccessful)
		deleteJobPolicy = DeleteJobPolicySuccessful
	}

	return deleteJobPolicy
}
