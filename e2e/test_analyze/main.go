package main

import (
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

func main(){

	config := api.Config{}

	log.Debugf("THIS IS VUL: %v", config)

	return
}
