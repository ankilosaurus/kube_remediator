package k8s

import (
	"go.uber.org/zap"
	"k8s.io/client-go/informers"
)

func NewSharedInformerFactory(logger *zap.Logger, ns string) (informers.SharedInformerFactory, error) {
	clientSet, err := newClientSet()
	if err != nil {
		logger.Warn("Error initializing informer:", zap.Error(err))
		return nil, err
	}

	factory := informers.NewSharedInformerFactoryWithOptions(clientSet, 0, informers.WithNamespace(ns))
	return factory, nil
}
