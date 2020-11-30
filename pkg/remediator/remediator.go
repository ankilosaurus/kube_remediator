
package remediator

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sync"
	"time"
)

// will later be used to make arrays or remediators / testing
type BaseIntf interface {
	Setup(*zap.Logger, k8s.ClientInterface) error
	Run(context.Context, *sync.WaitGroup)
}

type Base struct {
	BaseIntf
	client k8s.ClientInterface
	logger *zap.Logger
}

func (p *Base) Setup(logger *zap.Logger, client k8s.ClientInterface) error {
	p.client = client
	p.logger = logger
	return nil
}

func (p *Base) logStartAndStop(fn func()) {
	defer p.logger.Info("Stopping", zap.String("reason", "Signal"))
	p.logger.Info("Starting")
	fn()
}

func (p *Base) reconcileEvery(ctx context.Context, fn func(), interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	p.logStartAndStop(func(){
		// Run on start
		fn()

		for {
			select {
			case <-ticker.C:
				fn() // untested section
			case <-ctx.Done():
				return
			}
		}
	})

}

func (p *Base) deletePod(pod v1.Pod) {
	podInfo := []zap.Field{
		zap.String("name", pod.ObjectMeta.Name),
		zap.String("namespace", pod.ObjectMeta.Namespace),
	}
	p.tryWithLogging("Deleting Pod", podInfo, func() error {
		return p.client.DeletePod(&pod)
	})
}

func (p *Base) tryWithLogging(message string, logInfo []zap.Field, fn func() error) {
	p.logger.Info(message, logInfo...)
	if err := fn(); err != nil {
		p.logger.Warn("Error "+message, append(logInfo, zap.Error(err))...)
	}
}
