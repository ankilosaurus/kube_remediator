package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"github.com/aksgithub/kube_remediator/pkg/remediator"
	"k8s.io/apimachinery/pkg/util/runtime"
)

// catch interrupts to gracefully exit since otherwise goroutines get killed without running defer
func signalHandler(logger *zap.Logger) <-chan struct{} {
	stop := make(chan struct{})
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGSEGV,
			syscall.SIGABRT,
			syscall.SIGILL,
			syscall.SIGFPE)
		signal := <-c
		logger.Sugar().Warnf("Signal (%v) Received, Shutting Down", signal)
		close(stop)
	}()
	return stop
}

func main() {
	var wg sync.WaitGroup

	// init log
	logger, err := zap.NewProduction()
	runtime.Must(err)

	// init client
	k8sClient, err := k8s.NewClient(logger)
	if err != nil {
		logger.Panic("Error initializing k8s client: ", zap.Error(err))
	}

	// init remediators
	podRemediator, err := remediator.NewPodRemediator(logger, k8sClient)
	if err != nil {
		logger.Panic("Error initializing Pod remediator: ", zap.Error(err))
	}

	stopCh := signalHandler(logger)

	logger.Info("Starting Pod remediator")
	wg.Add(1)
	go func() {
		defer wg.Done()
		podRemediator.Run(stopCh)
	}()
	wg.Wait()

	logger.Fatal("Exiting main()")
}
