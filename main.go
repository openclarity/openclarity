package main

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"kubei/common"
	"kubei/webapp"
	"os"
	"strconv"
	"strings"
)


func getKubeClient() *kubernetes.Clientset {
	// creates the in-cluster config
	log.Infof("connecting....")
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	log.Infof("CONNECTED!!")

	return clientset
}

// Helper to read an environment variable into a string slice or return default value
func getEnvVariableAsSlice(name string, defaultVal []string, sep string) []string {
	valStr := os.Getenv(name)

	if valStr == "" {
		return defaultVal
	}

	val := strings.Split(valStr, sep)

	return val
}



func  getVerbose() bool {
	value := os.Getenv("VERBOSE")
	if value == "" {
		return false
	}
	verbose, err := strconv.ParseBool(value)
	if err != nil {
		log.Infof("VERBOSE is invalid. defaulting to false")
		return false
	}
	return verbose
}

func initLog() {
	if getVerbose() {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}
}

func initArgs() *common.ExecutionConfiguration {
	var err error
	executionConfig, err := getArgs()
	if err != nil {
		log.Error("Failed to get args. ", err)
		os.Exit(1)
	}

	return executionConfig
}

func getArgs() (*common.ExecutionConfiguration, error) {
	var err error

	currentNamespace := os.Getenv("MY-POD-NAMESPACE")
	if currentNamespace == "" {
		err = errors.New("MY-POD-NAMESPACE is mandatory but missing! Execution failed")
		return nil, err
	}

	var parallelism = 1 //default
	maxParallelismString := os.Getenv("MAX_PARALLELISM")
	if maxParallelismString == "" {
		log.Debugf("MAX_PARALLELISM is empty. defaulting to 1")
	} else {
		parallelism, err = strconv.Atoi(maxParallelismString)
		if err != nil {
			log.Infof("MAX_PARALLELISM is invalid. defaulting to 1")
		}
	}

	targetNamespace := os.Getenv("TARGET_NAMESPACE")
	if targetNamespace == "" {
		log.Debugf("TARGET_NAMESPACE is empty. Scanning ALL namespaces")
	}

	clairOutput := os.Getenv("CLAIR_OUTPUT")
	if clairOutput == "" {
		log.Debugf("CLAIR_OUTPUT is empty. defaulting to MEDIUM")
		clairOutput = "MEDIUM" //default
	}

	ignoreKubeSystem := true //default
	ignoreKubeSystemString := os.Getenv("IGNORE_KUBE_SYSTEM")
	if ignoreKubeSystemString == "" {
		log.Debugf("IGNORE_KUBE_SYSTEM is empty. defaulting to true")
	} else {
		ignoreKubeSystem, err = strconv.ParseBool(ignoreKubeSystemString)
		if err != nil {
			log.Warnf("IGNORE_KUBE_SYSTEM is invalid. defaulting to true")
		}
	}

	ignoreNamespaces := getEnvVariableAsSlice("IGNORE_NAMESPACES", []string{}, ",")

	klarTrace := false //default
	klarTraceString := os.Getenv("KLAR_TRACE")
	if klarTraceString == "" {
		log.Debugf("KLAR_TRACE is empty. defaulting to false")
	} else {
		klarTrace, err = strconv.ParseBool(klarTraceString)
		if err != nil {
			log.Warnf("KLAR_TRACE is invalid. defaulting to false")
		}
	}

	whitelistFile := os.Getenv("WHITELIST_FILE")
	if whitelistFile == "" {
		log.Debugf("WHITELIST_FILE is empty. Defaulting to no white-listing")
		whitelistFile = "" //default
	} else {
		data, err := ioutil.ReadFile(whitelistFile)
		if err != nil {
			log.Warnf("WHITELIST_FILE is invalid. Defaulting to no white-listing")
		} else {
			whitelistFile = "/usr/local/portshift/whitelist.txt"
			f, err := os.Create(whitelistFile)
			if err != nil {
				log.Warnf("failed to use create whitelist file: %s", err)
			} else {
				defer f.Close()
				_, err := f.Write(data)
				if err != nil {
					log.Warnf("failed to use create whitelist file: %s", err)
				}
			}
		}

	}

	clientset := getKubeClient()

	executionConfig := &common.ExecutionConfiguration{
		Clientset:        clientset,
		Parallelism:      parallelism,
		KubeiNamespace:   os.Getenv("MY-POD-NAMESPACE"),
		TargetNamespace:  targetNamespace,
		ClairOutput:      clairOutput,
		IgnoreKubeSystem: ignoreKubeSystem,
		IgnoreNamespaces: ignoreNamespaces,
		KlarTrace:        klarTrace,
		WhitelistFile:    whitelistFile,
	}
	return executionConfig, nil
}

func main() {
	initLog()
	executionConfiguration := initArgs()
	kubeiWebapp := webapp.Init(executionConfiguration)
	kubeiWebapp.Run()
}
