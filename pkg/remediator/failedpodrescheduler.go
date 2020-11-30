package remediator

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"strings"
	"sync"
	"time"
)

type FailedPodRescheduler struct {
	Base
	informerFactory informers.SharedInformerFactory
}

func (p *FailedPodRescheduler) Setup(logger *zap.Logger, client k8s.ClientInterface) error {
	informerFactory, err := client.NewSharedInformerFactory("")
	if err != nil {
		return err // untested section
	}
	p.informerFactory = informerFactory
	p.logger = logger
	p.client = client
	return nil
}

func (p *FailedPodRescheduler) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	p.logStartAndStop(func(){
		// Check for any Failed Pods first
		p.reschedulePods()
		// TODO: filter failed pods here to avoid overhead
		informer := p.informerFactory.Core().V1().Pods().Informer()

		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			UpdateFunc: p.rescheduleIfNecessary,
		})
		informer.Run(ctx.Done())

		<-ctx.Done()
	})
}

func (p *FailedPodRescheduler) reschedulePods() {
	p.logger.Info("Reconcile")
	for _, pod := range *p.getFailedPods() {
		p.rescheduleIfNecessary(nil, &pod)
	}
}

func (p *FailedPodRescheduler) rescheduleIfNecessary(oldObj, newObj interface{}) {
	pod := newObj.(*v1.Pod)
	if p.shouldReschedule(pod) {
		p.deletePod(*pod)
	}
}

func (p *FailedPodRescheduler) getFailedPods() *[]v1.Pod {
	pods, err := p.client.GetPods("", metav1.ListOptions{FieldSelector: "status.phase=Failed"})
	if err != nil {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return &[]v1.Pod{}
	}
	return &pods.Items
}

func (p *FailedPodRescheduler) shouldReschedule(pod *v1.Pod) bool {
	reason := strings.ToLower(pod.Status.Reason) // we saw OutOfCPU, OutOfcpu and Outofmemory
	if pod.Status.Phase != "Failed" || (reason != "outofcpu" && reason != "outofmemory") {
		return false
	}

	// Pods that would not be recreated need to stay
	if len(pod.ObjectMeta.OwnerReferences) == 0 {
		return false
	}

	// Job pods are deleted by Kubernetes
	for _, ownerReference := range pod.ObjectMeta.OwnerReferences {
		if ownerReference.Kind == "Job" {
			return false
		}
	}

	// Keep pods for 5 mins to be able to debug and log pipeline to find out metadata
	if pod.ObjectMeta.CreationTimestamp.Time.After(time.Now().Add(-5 * time.Minute)) {
		return false
	}

	return true
}
