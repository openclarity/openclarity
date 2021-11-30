package types

import (
	grype_models "github.com/anchore/grype/grype/presenter/models"

	dockle_types "github.com/Portshift/dockle/pkg/types"
	"github.com/Portshift/klar/docker"
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
	Vulnerabilities       *grype_models.Document
	DockerfileScanResults dockle_types.AssessmentMap
	LayerCommands         []*docker.FsLayerCommand
	Success               bool
	ScanErrors            []*ScanError
}

type ScanResults struct {
	ImageScanResults []*ImageScanResult
	Progress         ScanProgress
}
