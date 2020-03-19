package orchestrator

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"kubei/common"
	"sync"
	"testing"
	"time"
)

func createScanData(imageName string) *scanData {
	return &scanData{
		k8sContext: []*common.K8ExtendedContext{
			{
				Namespace: imageName+"Namespace",
				Container: imageName+"Container",
				Pod:       imageName+"Pod",
				Secret:    imageName+"Secret",
			},
		},
		scanUUID:   uuid.NewV4(),
		result:     nil,
		resultChan: make(chan bool),
		imageName:  imageName,
	}
}

func createImageToDataScan(totalImages int) map[string]*scanData {
	ret := make(map[string]*scanData)
	for i:=0;i<totalImages;i++{
		imageName := fmt.Sprintf("%v",i)
		ret[imageName] = createScanData(imageName)
	}

	return ret
}

func TestOrchestrator_jobBatchManagement(t *testing.T) {
	type fields struct {
		ImageK8ExtendedContextMap common.ImageK8ExtendedContextMap
		DataUpdateLock            *sync.Mutex
		ExecutionConfig           *common.ExecutionConfiguration
		scanIssuesMessages        *[]string
		batchCompletedScansCount  *int32
		k8ContextService          common.K8ContextServiceInterface
		imageToScanData           map[string]*scanData
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "",
			fields: fields{
				ExecutionConfig: &common.ExecutionConfiguration{
					Parallelism:      10,
					KubeiNamespace:   "KubeiNamespace",
					TargetNamespace:  "TargetNamespace",
					ClairOutput:      "ClairOutput",
					WhitelistFile:    "WhitelistFile",
					IgnoreNamespaces: nil,
					KlarTrace:        false,
				},
				imageToScanData: createImageToDataScan(20),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orc := &Orchestrator{
				ImageK8ExtendedContextMap: tt.fields.ImageK8ExtendedContextMap,
				DataUpdateLock:            tt.fields.DataUpdateLock,
				ExecutionConfig:           tt.fields.ExecutionConfig,
				scanIssuesMessages:        tt.fields.scanIssuesMessages,
				batchCompletedScansCount:  tt.fields.batchCompletedScansCount,
				k8ContextService:          tt.fields.k8ContextService,
				imageToScanData:           tt.fields.imageToScanData,
			}

			go orc.jobBatchManagement()

			time.Sleep(5 * time.Second)

			for _, data := range orc.imageToScanData {
				go func() {
					data.resultChan <- true
				}()
				time.Sleep(1 * time.Second)
			}

			time.Sleep(10 * time.Second)
		})
	}
}
