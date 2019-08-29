package remediator

import (
	"time"
	"go.uber.org/zap"

	v1 "k8s.io/api/core/v1"
	//v1beta1 "k8s.io/api/policy/v1beta1"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
)

type PodFilter struct {
	annotation string
	failureThreshold int32
	namespace string
}

type PodRemediator struct {
	client *k8s.Client
	logger *zap.Logger
	frequency int // in minutes
	filter PodFilter
}

// Entrypoint
func (p  *PodRemediator) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(p.frequency) * time.Minute)
	MainLoop:
	for {
		select {
		case <-ticker.C:
			p.rescheduleUnhealthyPods()
		case s := <-stopCh:
			p.logger.Sugar().Infof("Pod remediator received a signal (%v) to terminate", s)
			break MainLoop
		}
	}
}

func (p *PodRemediator) rescheduleUnhealthyPods() {
	unHealthyPods, err := p.getUnhealthyPods()
	if err != nil { return }

	// reschedule pods
	for _, pod := range unHealthyPods.Items {
		p.logger.Sugar().Infof("Pod (%v) in namespace (%v)", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
		if p.canRecoverPod(&pod) {
			if p.podHasController(&pod) == true {
				p.logger.Sugar().Infof("Pod (%v) in namespace (%v) is marked for deletion",
					pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
				p.client.DeletePod(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			} else {
				//TODO: restart the pod
				p.logger.Sugar().Warnf("Pod (%v) in namespace (%v) without Owner can't be deleted",
					pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			}
		}
	}
}

func (p *PodRemediator) podHasController(pod *v1.Pod) bool {
	return len(pod.ObjectMeta.OwnerReferences) > 0
}

func (p *PodRemediator) canRecoverPod(pod *v1.Pod) bool {
	return pod.ObjectMeta.Annotations[p.filter.annotation] == "true" //TODO: check if it should be True
}

func (p *PodRemediator) getUnhealthyPods() (*v1.PodList, error) {
	p.logger.Info("getUnhealthyPods: START")
	allPods, err := p.client.GetPods(p.filter.namespace)
	if err != nil  {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return nil, err
	}
	unhealthyPods := &v1.PodList{}
	for _, pod := range allPods.Items {
		if p.isPodUnhealthy(&pod) == true {
			p.logger.Sugar().Infof("Pod (%v) in namespace (%v) is unhealthy",
				pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			unhealthyPods.Items = append(unhealthyPods.Items, pod)
		}
	}
	p.logger.Info("getUnhealthyPods: END")
	return unhealthyPods, nil
}

// This is not 100% reliable because Pod could toggle between Terminated with Error and Waiting with CrashLoopBackOff
func (p *PodRemediator) isPodUnhealthy(pod *v1.Pod) bool {
	// Check if any of Containers is in CrashLoop
	for _, containerStatus := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		if containerStatus.RestartCount > p.filter.failureThreshold {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				return true
			}
		}
		//TODO: other conditions
	}
	return false
}

func NewPodRemediator(logger *zap.Logger, client *k8s.Client) (*PodRemediator, error) {
	//TODO: read pod config
	p := &PodRemediator{
			client: client,
			logger: logger,
			frequency: 1, // Use duration
			filter: PodFilter{
				annotation: "kube_remediator/restart_unhealthy",
				failureThreshold: 5,
				namespace: "",
			},
	}
	return p, nil
}
