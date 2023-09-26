package trivy

import (
	"fmt"

	"github.com/aquasecurity/trivy/pkg/flag"
)

func GetTrivyDBOptions() (flag.DBOptions, error) {
	// Get the Trivy CVE DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	dbRepoDefaultValue, ok := flag.DBRepositoryFlag.Default.(string)
	if !ok {
		return flag.DBOptions{}, fmt.Errorf("unable to get trivy DB repo config")
	}

	// Get the Trivy JAVA DB URL default value from the trivy
	// configuration, we may want to make this configurable in the
	// future.
	javaDBRepoDefaultValue, ok := flag.JavaDBRepositoryFlag.Default.(string)
	if !ok {
		return flag.DBOptions{}, fmt.Errorf("unable to get trivy java DB repo config")
	}

	return flag.DBOptions{
		DBRepository:     dbRepoDefaultValue,     // Use the default trivy source for the vuln DB
		JavaDBRepository: javaDBRepoDefaultValue, // Use the default trivy source for the java DB
		NoProgress:       true,                   // Disable the interactive progress bar
	}, nil
}
