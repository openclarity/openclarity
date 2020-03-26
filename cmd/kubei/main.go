package main

import (
	"github.com/Portshift/kubei/pkg/config"
	"github.com/Portshift/kubei/pkg/webapp"
	log "github.com/sirupsen/logrus"
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
	kubeiWebapp := webapp.Init(conf, scanConfig)
	kubeiWebapp.Run()
}
