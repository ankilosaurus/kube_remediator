package remediator

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"github.com/aksgithub/kube_remediator/pkg/metrics"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"sync"
)

// TODO: this cannot be global since we have multiple remediator in this package ... folder can be set though but should
// still not be here
var CONFIG_FILE = "config/crash_loop_back_off_rescheduler.json"

type PodFilter struct {
	annotation       string
	failureThreshold int32
	namespace        string
}

type CrashLoopBackOffRescheduler struct {
	Base
	filter          PodFilter
	informerFactory informers.SharedInformerFactory
	metrics         *metrics.CrashLoopBackOff_Metrics
}

func (p *CrashLoopBackOffRescheduler) Setup(logger *zap.Logger, client k8s.ClientInterface) error {
	logger.Info("Reading config", zap.String("file", CONFIG_FILE))
	viper.SetConfigFile(CONFIG_FILE)
	viper.SetConfigType("json")
	viper.SetDefault("annotation", "kube-remediator/CrashLoopBackOffRemediator")
	viper.SetDefault("failureThreshold", 5)
	viper.SetDefault("namespace", "")

	if err := viper.ReadInConfig(); err != nil {
		return err // untested section
	}

	logger.Sugar().Infof("Config %v", viper.AllSettings()) // TODO: prefer using zap.Map or something like that
	filter := PodFilter{
		annotation:       viper.GetString("annotation"),
		failureThreshold: viper.GetInt32("failureThreshold"),
		namespace:        viper.GetString("namespace"),
	}

	metrics := metrics.NewCrashLoopBackOffMetrics(logger)
	metrics.Register()

	informerFactory, err := client.NewSharedInformerFactory(filter.namespace)
	if err != nil {
		return err // untested section
	}
	p.informerFactory = informerFactory
	p.filter = filter
	p.metrics = metrics
	p.logger = logger
	p.client = client
	return nil
}

func (p *CrashLoopBackOffRescheduler) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	p.logger.Info("Starting")
	// Check for any CrashLoopBackOff Pods first
	p.reschedulePods()

	informer := p.informerFactory.Core().V1().Pods().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: p.rescheduleIfNecessary,
	})
	informer.Run(ctx.Done())

	<-ctx.Done()
	p.metrics.UnRegister()
	p.logger.Info("Stopping", zap.String("reason", "Signal"))
}

func (p *CrashLoopBackOffRescheduler) reschedulePods() {
	p.logger.Info("Running")
	for _, pod := range *p.getCrashLoopBackOffPods() {
		p.rescheduleIfNecessary(nil, &pod)
	}
}

func (p *CrashLoopBackOffRescheduler) rescheduleIfNecessary(oldObj, newObj interface{}) {
	pod := newObj.(*v1.Pod)
	if p.shouldReschedule(pod) {
		p.deletePod(*pod)
	}
}

func (p *CrashLoopBackOffRescheduler) getCrashLoopBackOffPods() *[]v1.Pod {
	pods, err := p.client.GetPods(p.filter.namespace, metav1.ListOptions{})
	if err != nil {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return &[]v1.Pod{}
	}
	var unhealthyPods []v1.Pod
	for _, pod := range pods.Items {
		if p.shouldReschedule(&pod) {
			unhealthyPods = append(unhealthyPods, pod)
		}
	}
	return &unhealthyPods
}

func (p *CrashLoopBackOffRescheduler) shouldReschedule(pod *v1.Pod) bool {
	return (p.filter.annotation == "" || pod.ObjectMeta.Annotations[p.filter.annotation] != "false") && // not opted-out
		len(pod.ObjectMeta.OwnerReferences) > 0 && // Assuming Pod has owner reference of kind Controller
		p.isPodUnhealthy(pod)
}

// This is not 100% reliable because Pod could toggle between Terminated with Error and Waiting with CrashLoopBackOff
func (p *CrashLoopBackOffRescheduler) isPodUnhealthy(pod *v1.Pod) bool {
	// Check if any of Containers is in CrashLoop
	statuses := append(pod.Status.InitContainerStatuses, pod.Status.ContainerStatuses...)
	for _, containerStatus := range statuses {
		if containerStatus.RestartCount >= p.filter.failureThreshold {
			if containerStatus.State.Waiting != nil && containerStatus.State.Waiting.Reason == "CrashLoopBackOff" {
				return true
			}
		}
	}
	return false
}
