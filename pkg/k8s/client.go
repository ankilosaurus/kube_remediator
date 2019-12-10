package k8s

import (
	"go.uber.org/zap"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type ClientInterface interface {
	GetPods(namespace string) (*apiv1.PodList, error)
	DeletePod(pod *apiv1.Pod) error
}

type Client struct {
	logger    *zap.Logger
	clientSet *kubernetes.Clientset
}

func (c *Client) GetPods(namespace string) (*apiv1.PodList, error) {
	return c.clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func (c *Client) DeletePod(pod *apiv1.Pod) error {
	return c.clientSet.CoreV1().Pods(pod.ObjectMeta.Namespace).Delete(pod.ObjectMeta.Name, &metav1.DeleteOptions{})
}

func newClientSet() (*kubernetes.Clientset, error) {
	var err error
	var config *restclient.Config
	if os.Getenv("KUBERNETES_SERVICE_HOST") == "" {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = filepath.Join(os.Getenv("HOME"), ".kube", "config")
		}
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// Reads config when in cluster
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func NewClient(logger *zap.Logger) (*Client, error) {
	clientSet, err := newClientSet()
	if err != nil {
		return nil, err
	}

	return &Client{clientSet: clientSet, logger: logger}, err
}
