package remediator

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
)

type OldPodDeleter struct {
	Base
}

func (p *OldPodDeleter) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	p.reconcileEvery(ctx, p.deleteOldPods, 1*time.Hour)
}

func (p *OldPodDeleter) deleteOldPods() {
	p.logger.Info("Running")

	// get all pods that opted in to deletion
	pods, err := p.client.GetPods("", metav1.ListOptions{
		LabelSelector: "kube-remediator/OldPodDeleter=true",
	})
	if err != nil {
		p.logger.Error("Error getting pod list", zap.Error(err))
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
