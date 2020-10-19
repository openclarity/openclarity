package types

import (
	"github.com/Portshift/klar/clair"
	dockle_types "gitlab.com/portshift/dockle/pkg/types"
)

type ScanProgress struct {
	ImagesToScan          uint32
	ImagesStartedToScan   uint32
	ImagesCompletedToScan uint32
}

type ImageScanResult struct {
	PodName               string
	PodNamespace          string
	ImageName             string
	ContainerName         string
	ImageHash             string
	PodUid                string
	Vulnerabilities       []*clair.Vulnerability
	DockerfileScanResults dockle_types.AssessmentMap
	Success               bool
	ScanErrMsg            string
}

type ScanResults struct {
	ImageScanResults []*ImageScanResult
	Progress         ScanProgress
}
