package client

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Clientset struct {
	*kubernetes.Clientset
}

func NewClientset() *Clientset {
	return &Clientset{
		connect(),
	}
}

func connect() *kubernetes.Clientset {
	home, exists := os.LookupEnv("HOME")
	if !exists {
		home = "/root"
	}

	configPath := filepath.Join(home, ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", configPath)

	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err)
	}

	return clientset
}
