package common

import (
	"fmt"
	"os/exec"
)

func DescribeKubeClarityDeployment() {
	cmd := exec.Command("kubectl", "-n", "kubeclarity", "describe", "deployments.apps", KubeClarityDeploymentName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe deployments.apps -n kubeclarity kubeclarity-kubeclarity:\n %s\n", out)
}

func DescribeKubeClarityPods() {
	cmd := exec.Command("kubectl", "-n", "kubeclarity", "describe", "pods")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe pods -n kubeclarity:\n %s\n", out)
}

func GetKubeClarityPods() {
	cmd := exec.Command("kubectl", "-n", "kubeclarity", "get", "pods")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl get pods -n kubeclarity:\n %s\n", out)
}


