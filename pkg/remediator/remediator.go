package remediator

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sync"
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
