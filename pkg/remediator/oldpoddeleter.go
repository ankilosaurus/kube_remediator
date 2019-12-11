package remediator

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
)

type OldPodDeleter struct {
	Base
}

// Entry point
func (p *OldPodDeleter) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	p.logger.Info("Starting")

	// Run on start
	p.deleteOldPods()

	for {
		select {
		case <-ticker.C:
			p.deleteOldPods() // untested section
		case <-ctx.Done():
			p.logger.Info("Stopping", zap.String("reason", "Signal"))
			return
		}
	}
}

func (p *OldPodDeleter) deleteOldPods() {
	p.logger.Info("Running")

	// get all pods that opted in to deletion
	pods, err := p.client.GetPods("", metav1.ListOptions{
		LabelSelector: "kube-remediator/OldPodDeleter=true",
	})
	if err != nil {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return
	}

	// deleteOldPods those that are too old
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, pod := range pods.Items {
		if pod.ObjectMeta.CreationTimestamp.Time.After(cutoff) {
			continue
		}
		p.deletePod(pod)
	}
}

func NewOldPodDeleter(logger *zap.Logger, client k8s.ClientInterface) (*OldPodDeleter, error) {
	p := &OldPodDeleter{}
	p.client = client
	p.logger = logger
	return p, nil
}
