package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/cisco-open/kubei/pkg/config"
	"github.com/cisco-open/kubei/pkg/webapp"
)

func initLog(verbose bool) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	conf := config.LoadConfig()
	scanConfig := config.LoadScanConfig()

	initLog(conf.Verbose)

	log.Debugf("config=%+v", conf)
	log.Debugf("scanConfig=%+v", scanConfig)

	kubeiWebapp, err := webapp.Init(conf, scanConfig)
	if err != nil {
		log.Fatalf("Failed to init webapp: %v", err)
	}

	kubeiWebapp.Run()
}
