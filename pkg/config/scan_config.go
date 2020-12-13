package config

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	"strconv"
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
	DeleteJobPolicy       = "DELETE_JOB_POLICY"
	ShouldScanDockerFile  = "SHOULD_SCAN_DOCKERFILE"
	ScannerServiceAccount = "SCANNER_SERVICE_ACCOUNT"
	RegistryInsecure      = "REGISTRY_INSECURE"
)

type ScanConfig struct {
	MaxScanParallelism    int
	TargetNamespace       string
	SeverityThreshold     string
	KlarImageName         string
	IgnoredNamespaces     []string
	JobResultTimeout      time.Duration
	DeleteJobPolicy       DeleteJobPolicyType
	ShouldScanDockerFile  bool
	ScannerServiceAccount string
	RegistryInsecure      string
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
	viper.SetDefault(RegistryInsecure, "false")

	viper.AutomaticEnv()
}

func LoadScanConfig() *ScanConfig {
	setScanConfigDefaults()

	shouldScanDockerFile := viper.GetBool(ShouldScanDockerFile)
	registryInsecure, _ := strconv.ParseBool(viper.GetString(RegistryInsecure))
	// Disable DockerFile scan if insecure registry is set - currently not supported
	if registryInsecure {
		shouldScanDockerFile = false
	}

	return &ScanConfig{
		MaxScanParallelism:    viper.GetInt(MaxParallelism),
		TargetNamespace:       viper.GetString(TargetNamespace),
		SeverityThreshold:     viper.GetString(SeverityThreshold),
		KlarImageName:         viper.GetString(KlarImageName),
		IgnoredNamespaces:     strings.Split(viper.GetString(IgnoreNamespaces), ","),
		JobResultTimeout:      viper.GetDuration(JobResultTimeout),
		DeleteJobPolicy:       getDeleteJobPolicyType(viper.GetString(DeleteJobPolicy)),
		ShouldScanDockerFile:  shouldScanDockerFile,
		ScannerServiceAccount: viper.GetString(ScannerServiceAccount),
		RegistryInsecure:      viper.GetString(RegistryInsecure),
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
