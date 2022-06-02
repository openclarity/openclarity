// Copyright © 2021 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"fmt"
	"os/exec"
)

func DescribeKubeClarityDeployment() {
	cmd := exec.Command("kubectl", "-n", KubeClarityNamespace, "describe", "deployments.apps", KubeClarityDeploymentName)

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DescribeKubeClarityDeployment failed. Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe deployments.apps -n kubeclarity kubeclarity-kubeclarity:\n %s\n", out)
}

func DescribeKubeClarityPods() {
	cmd := exec.Command("kubectl", "-n", KubeClarityNamespace, "describe", "pods")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("DescribeKubeClarityPods failed. Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl describe pods -n kubeclarity:\n %s\n", out)
}

func GetKubeClarityPods() {
	cmd := exec.Command("kubectl", "-n", KubeClarityNamespace, "get", "pods")

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("GetKubeClarityPods failed. Failed to execute command. %v, %s", err, out)
		return
	}
	fmt.Printf("kubectl get pods -n kubeclarity:\n %s\n", out)
}
