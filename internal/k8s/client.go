package k8s

import (
	"fmt"
	"os"
	"strings"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient(path string) (*kubernetes.Clientset, error) {
	config, err := getConfig(path)
	if err != nil {
		return nil, fmt.Errorf("generate kube config: %w", err)
	}
	return kubernetes.NewForConfig(config)
}

func getConfig(kubeConfigPath string) (*rest.Config, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		path, err := resolveConfigPath(kubeConfigPath)
		if err != nil {
			return nil, err
		}
		// Grab it from kubeconfig if we can't get an in-cluster config
		config, err = clientcmd.BuildConfigFromFlags("", path)

		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func resolveConfigPath(path string) (string, error) {
	if strings.Contains(path, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = strings.Replace(path, "~", home, -1)
	}

	return path, nil
}
