package remediator

import (
	"go.uber.org/zap"
)

// will later be used to make arrays or remediators / testing
type Remediator interface {
	Run(logger *zap.Logger, stopCH <-chan struct{})
}
