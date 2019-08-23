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
	frequency int // in minutes
	filter PodFilter
}

func (p  *PodRemediator) Run(logger *zap.Logger, stopCh <-chan struct{}) {
	ticker := time.NewTicker(time.Duration(p.frequency) * time.Minute)
	for {
		select {
		case <-ticker.C:
			p.recoverUnhealthyPods(logger)
		case s := <-stopCh:
			logger.Sugar().Infof("Pod remediator received a signal (%v) to terminate", s)
			break
		}
	}
}

func (p *PodRemediator) recoverUnhealthyPods(logger *zap.Logger) {
	unHealthyPods := p.getUnhealthyPods(logger)

	// delete pods
	for _, pod := range unHealthyPods.Items {
		logger.Sugar().Infof("Pod (%v)", pod.ObjectMeta.Name)
		if p.canRecoverPod(logger, &pod) {
			if p.podHasController(logger, &pod) == true {
				logger.Sugar().Infof("Pod (%v) in namespace (%v) is marked for deletion",
					pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
				//p.client.DeletePod(pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			} else {
				//TODO: restart the pod
				logger.Sugar().Infof("Pod (%v) in namespace (%v) is marked for restart",
					pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			}
		}
	}
}

func (p *PodRemediator) podHasController(logger *zap.Logger, pod *v1.Pod) bool {
	for _, reference := range pod.ObjectMeta.OwnerReferences {
		if reference.Controller != nil && *reference.Controller == true {
			return true
		}
	}
	return false
}

func (p *PodRemediator) canRecoverPod(logger *zap.Logger, pod *v1.Pod) bool {
	for k, v := range pod.ObjectMeta.Annotations {
		if k == p.filter.annotation && v == "true" {
			return true
		}
	}
	return false
}

func (p *PodRemediator) getUnhealthyPods(logger *zap.Logger) (*v1.PodList) {
	logger.Info("getUnhealthyPods: START")
	allPods, err := p.client.GetPods(p.filter.namespace)
	if err != nil  {
		logger.Error("Error getting pod list: ", zap.Error(err))
		return nil
	}
	unhealthyPods := &v1.PodList{}
	for _, pod := range allPods.Items {
		if p.isPodUnhealthy(logger, &pod) == true {
			logger.Sugar().Infof("Pod (%v) in namespace (%v) is unhealthy",
				pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			unhealthyPods.Items = append(unhealthyPods.Items, pod)
		}
	}
	logger.Info("getUnhealthyPods: END")
	return unhealthyPods
}

func (p *PodRemediator) isPodUnhealthy(logger * zap.Logger, pod *v1.Pod) bool {
	if pod == nil {
		logger.Warn("Pod is not valid")
		return false
	}
	if p.hasUnhealthyContainer(logger, pod) == true { return true }
	//TODO: other conditions
	return false
}

func (p *PodRemediator) hasUnhealthyContainer(logger * zap.Logger, pod *v1.Pod) bool {
	if pod == nil {
		logger.Warn("Pod is not valid")
		return false
	}

	// Check if any of Containers is in CrashLoop
	for _, containerStatus := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		if containerStatus.State.Waiting != nil &&
			containerStatus.State.Waiting.Reason == "CrashLoopBackOff" &&
			containerStatus.RestartCount > p.filter.failureThreshold {
			return true
		}
		//TODO: other conditions
	}
	return false
}

func GetNewPodRemediator(logger *zap.Logger, client *k8s.Client) (*PodRemediator, error) {
	//TODO: read pod config
	p := &PodRemediator{
			client: client,
			frequency: 1, // Use duration
			filter: PodFilter{
				annotation: "kube_remediator/restart_unhealthy",
				failureThreshold: 5,
				namespace: "",
			},
	}
	return p, nil
}
