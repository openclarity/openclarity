// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
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

package k8s

import (
	"context"
	"fmt"
	"strings"

	"github.com/containers/image/v5/docker/reference"
	"github.com/google/go-containerregistry/pkg/authn"
	authnk8s "github.com/google/go-containerregistry/pkg/authn/kubernetes"
	"github.com/google/go-containerregistry/pkg/name"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const MaxK8sJobName = 63

func GetPodImagePullSecrets(clientset kubernetes.Interface, pod corev1.Pod) []*corev1.Secret {
	secrets := make([]*corev1.Secret, 0, len(pod.Spec.ImagePullSecrets))
	for _, secretName := range pod.Spec.ImagePullSecrets {
		secret, err := clientset.CoreV1().Secrets(pod.Namespace).Get(context.TODO(), secretName.Name, metav1.GetOptions{})
		if err != nil {
			log.Warnf("Failed to get secret %s in namespace %s. %+v", secretName.Name, pod.Namespace, err)
			continue
		}
		secrets = append(secrets, secret)
	}

	return secrets
}

func GetMatchingSecretName(secrets []*corev1.Secret, imageName string) string {
	named, err := reference.ParseNormalizedNamed(imageName)
	if err != nil {
		log.Warnf("failed to normalized image name: %v", err)
		return ""
	}

	for _, secret := range secrets {
		authorization, err := GetAuthConfig([]corev1.Secret{*secret}, named)
		if err != nil {
			log.Debugf("Failed to get auth config - skipping secret: %v", err)
			continue
		}

		if authorization.Password != "" && authorization.Username != "" {
			log.Debugf("Matching secret was found. name=%v", secret.Name)
			return secret.Name
		}

		// TODO: should we try to fetch the image to see that the username and password really match?
	}

	log.Debugf("No matching secret found.")
	return ""
}

func GetAuthConfig(secrets []corev1.Secret, named reference.Named) (*authn.AuthConfig, error) {
	keychain, err := authnk8s.NewFromPullSecrets(context.TODO(), secrets)
	if err != nil {
		return nil, fmt.Errorf("failed to create keychain: %v", err)
	}

	repository, err := name.NewRepository(reference.FamiliarName(named))
	if err != nil {
		return nil, fmt.Errorf("failed to create repository: %v", err)
	}

	authenticator, err := keychain.Resolve(repository)
	if err != nil {
		return nil, fmt.Errorf("failed to create authenticator: %v", err)
	}

	authorization, err := authenticator.Authorization()
	if err != nil {
		return nil, fmt.Errorf("failed to create authorization: %v", err)
	}

	return authorization, nil
}

// ParseImageHash extracts image hash from image ID
// input: docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
// output: 6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
func ParseImageHash(imageID string) string {
	index := strings.LastIndex(imageID, ":")
	if index == -1 {
		return ""
	}

	return imageID[index+1:]
}

// ParseImageID remove "docker-pullable://" prefix from imageID if exists
// https://github.com/kubernetes/kubernetes/issues/95968
// input: docker-pullable://gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
// output: gcr.io/development-infra-208909/kubeclarity@sha256:6d5d0e4065777eec8237cefac4821702a31cd5b6255483ac50c334c057ffecfa
func ParseImageID(imageID string) string {
	return strings.TrimPrefix(imageID, "docker-pullable://")
}
