package main

import (
	log "github.com/sirupsen/logrus"
	"kubei/pkg/config"
	"kubei/pkg/webapp"
)

func initLog(verbose bool) {
	if verbose {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func main() {
	config, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	initLog(config.Verbose)
	kubeiWebapp := webapp.Init(config)
	kubeiWebapp.Run()
}
