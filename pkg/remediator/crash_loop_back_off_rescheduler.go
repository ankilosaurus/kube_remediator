package remediator

import (
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"time"
)

type PodFilter struct {
	annotation       string
	failureThreshold int32
	namespace        string
}

type CrashLoopBackOffRescheduler struct {
	client    *k8s.Client
	logger    *zap.SugaredLogger
	frequency time.Duration
	filter    PodFilter
}

// Entrypoint
func (p *CrashLoopBackOffRescheduler) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(p.frequency)
	for {
		select {
		case <-ticker.C:
			p.reschedulePods()
		case <-stopCh:
			p.logger.Info("Received signal to stop")
			return
		}
	}
}

func (p *CrashLoopBackOffRescheduler) reschedulePods() {
	for _, pod := range *p.getCrashLoopBackOffPods() {
		if p.canRecoverPod(&pod) { // TODO: stop early
			podInfo := []zap.Field{
				zap.String("name", pod.ObjectMeta.Name),
				zap.String("namespace", pod.ObjectMeta.Namespace),
			}
			if p.podHasController(&pod) {
				p.tryWithLogging("Deleting Pod", &podInfo, func() error {
					return p.client.DeletePod(&pod)
				})
			} else {
				p.tryWithLogging("Recreating Pod", &podInfo, func() error {
					return p.client.RecreatePod(&pod)
				})
			}
		}
	}
}

func (p *CrashLoopBackOffRescheduler) tryWithLogging(message string, logInfo *[]zap.Field, fn func() error) {
	p.logger.Info(message, logInfo)
	if err := fn(); err != nil {
		p.logger.Error("Error "+message, append(*logInfo, zap.Error(err)))
	}
}

// Assuming Pod has owner reference of kind Controller
func (p *CrashLoopBackOffRescheduler) podHasController(pod *v1.Pod) bool {
	return len(pod.ObjectMeta.OwnerReferences) > 0
}

func (p *CrashLoopBackOffRescheduler) canRecoverPod(pod *v1.Pod) bool {
	return pod.ObjectMeta.Annotations[p.filter.annotation] == "true" //TODO: check if it should be True
}

func (p *CrashLoopBackOffRescheduler) getCrashLoopBackOffPods() *[]v1.Pod {
	p.logger.Info("getCrashLoopBackOffPods: START")
	allPods, err := p.client.GetPods(p.filter.namespace)
	if err != nil {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return &[]v1.Pod{}
	}
	unhealthyPods := &v1.PodList{}
	for _, pod := range allPods.Items {
		if p.isPodUnhealthy(&pod) == true {
			p.logger.Infof("Pod (%v) in namespace (%v) is unhealthy",
				pod.ObjectMeta.Name, pod.ObjectMeta.Namespace)
			unhealthyPods.Items = append(unhealthyPods.Items, pod)
		}
	}
	p.logger.Info("getCrashLoopBackOffPods: END")
	return &unhealthyPods.Items
}

// This is not 100% reliable because Pod could toggle between Terminated with Error and Waiting with CrashLoopBackOff
func (p *CrashLoopBackOffRescheduler) isPodUnhealthy(pod *v1.Pod) bool {
	// Check if any of Containers is in CrashLoop
	for _, containerStatus := range append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...) {
		if containerStatus.RestartCount > p.filter.failureThreshold {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				return true
			}
		}
	}
	//TODO: other conditions
	return false
}

func NewPodRemediator(logger *zap.SugaredLogger, client *k8s.Client) (*CrashLoopBackOffRescheduler, error) {
	viper.SetConfigFile("config/pod_remediator.json")
	viper.SetConfigType("json")
	logger.Infof("Reading config from %v", viper.ConfigFileUsed())
	filter := PodFilter{
		annotation:       "kube_remediator/restart_unhealthy",
		failureThreshold: 5,
		namespace:        "",
	}
	if err := viper.ReadInConfig(); err != nil {
		logger.Error("Failed to read config file", zap.Error(err))
	} else {
		logger.Infof("Config: %v", viper.AllSettings())
		filter = PodFilter{
			annotation:       viper.GetString("annotation"),
			failureThreshold: viper.GetInt32("failureThreshold"),
			namespace:        viper.GetString("namespace"),
		}
	}

	p := &CrashLoopBackOffRescheduler{
		client:    client,
		logger:    logger,
		frequency: viper.GetDuration("frequency"),
		filter:    filter,
	}
	return p, nil
}
