package creds

import (
	"context"
	log "github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func isSecretExists(clientset kubernetes.Interface, name, namespace string) bool {
	_, err := clientset.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return false
		}

		log.Errorf("Failed to get secret. secret=%v: %v", name, err)
		return false
	}

	return true
}
