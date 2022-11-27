package job_factory

import (
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/openclarity/kubeclarity/shared/pkg/job_manager"
)

var (
	createJobFuncs map[string]CreateJobFunc // scanner name to CreateJobFunc
	once           sync.Once
)

type CreateJobFunc func(conf job_manager.IsConfig, logger *logrus.Entry, resultChan chan job_manager.Result) job_manager.Job

func RegisterCreateJobFunc(name string, createJobFunc CreateJobFunc) {
	once.Do(func() {
		createJobFuncs = make(map[string]CreateJobFunc)
	})
	if _, ok := createJobFuncs[name]; ok {
		logrus.Fatalf("%q already registered", name)
	}
	createJobFuncs[name] = createJobFunc
}

func GetCreateJobFuncs() map[string]CreateJobFunc {
	return createJobFuncs
}
