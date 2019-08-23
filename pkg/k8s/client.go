package k8s

import (
	"go.uber.org/zap"
	apiv1 "k8s.io/api/core/v1"
	v1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	clientSet *kubernetes.Clientset
}

func (c *Client) GetPods(namespace string) (*apiv1.PodList , error) {
	return c.clientSet.CoreV1().Pods(namespace).List(metav1.ListOptions{})
}

func (c *Client) DeletePod(name string, namespace string) error {
	return c.clientSet.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
}

func (c *Client) CreatePod(pod *apiv1.Pod) (*apiv1.Pod, error) {
	return c.clientSet.CoreV1().Pods(pod.ObjectMeta.Namespace).Create(pod)
}

func (c *Client) GetPodDisruptionBudget(name string, namespace string) (*v1beta1.PodDisruptionBudget, error) {
	podDisruptionBudget, err := c.clientSet.PolicyV1beta1().PodDisruptionBudgets(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return podDisruptionBudget, err
}

func (c *Client) GetPodDisruptionBudgets(namespace string) (*v1beta1.PodDisruptionBudgetList, error) {
	podDisruptionBudgets, err := c.clientSet.PolicyV1beta1().PodDisruptionBudgets(namespace).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return podDisruptionBudgets, err
}

func GetNewClient(logger *zap.Logger) (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &Client{clientSet: clientSet}, err
}
