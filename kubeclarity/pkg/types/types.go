package types

import "github.com/Portshift/klar/clair"

type ScanProgress struct {
	ImagesToScan          uint32
	ImagesStartedToScan   uint32
	ImagesCompletedToScan uint32
}

type ImageScanResult struct {
	PodName         string
	PodNamespace    string
	ImageName       string
	ContainerName   string
	ImageHash       string
	PodUid          string
	Vulnerabilities []*clair.Vulnerability
	Success         bool
}

type ScanResults struct {
	ImageScanResults []*ImageScanResult
	Progress         ScanProgress
}
