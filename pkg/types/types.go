package types

import (
	"github.com/Portshift/klar/clair"
	"github.com/Portshift/klar/docker"
	dockle_types "github.com/Portshift/dockle/pkg/types"
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
	LayerCommands         []*docker.FsLayerCommand
	Success               bool
	ScanErrMsg            string
}

type ScanResults struct {
	ImageScanResults []*ImageScanResult
	Progress         ScanProgress
}
