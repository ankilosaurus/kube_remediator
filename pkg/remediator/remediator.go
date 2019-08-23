package remediator

import (
	"go.uber.org/zap"
)

type Remediator interface {
	Run(logger *zap.Logger, stopCH <-chan struct{})
}
