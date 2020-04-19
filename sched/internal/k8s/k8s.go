// Package k8s contains utilities to interact with Kubernetes.
// It provides a more focused interface over k8s.io/client-go, to
// reduce coupling and simplify stubs.
package k8s

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Client interface {
}

type k8s struct {
	clientset *kubernetes.Clientset
}

func New(kubeconfig string) (Client, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &k8s{clientset}, nil
}

// Below is the definition of the k8s client stub.
// It is only useful for unit testing and should not be used in
// the actual code.

type stub struct{}

func NewStub() Client {
	return &stub{}
}
