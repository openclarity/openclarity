package types

type ScanError struct {
	ErrMsg    string
	ErrType   string
	ErrSource ScanErrorSource
}

type ScanErrorType string

const (
	JobRun     ScanErrorType = "errorJobRun"
	JobTimeout ScanErrorType = "errorJobTimeout"
)

type ScanErrorSource string

const (
	ScanErrSourceDockle ScanErrorSource = "ScanErrSourceDockle"
	ScanErrSourceVul    ScanErrorSource = "ScanErrSourceVulnerability"
	ScanErrSourceJob    ScanErrorSource = "ScanErrSourceJob"
)
