package remediator

import (
	"context"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"time"
)

type CompletedPodDeleter struct {
	Base
}

func (p *CompletedPodDeleter) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	p.reconcileEvery(ctx, p.deleteCompletedPods, 1 * time.Hour)
}

func (p *CompletedPodDeleter) deleteCompletedPods() {
	p.logger.Info("Running")

	// get completed pods
	pods, err := p.client.GetPods("", metav1.ListOptions{FieldSelector: "status.phase=Completed"})
	if err != nil {
		p.logger.Error("Error getting pod list: ", zap.Error(err))
		return
	}

	// delete those that are too old (could delete pods that ran a long time early, but good enough for now)
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, pod := range pods.Items {
		if pod.ObjectMeta.CreationTimestamp.Time.After(cutoff) {
			continue
		}
		p.deletePod(pod)
	}
}



