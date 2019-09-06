package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/aksgithub/kube_remediator/pkg/healthz"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"github.com/aksgithub/kube_remediator/pkg/remediator"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
)

// catch interrupts to gracefully exit since otherwise goroutines get killed without running defer
func signalHandler(cancelFn func(), wg *sync.WaitGroup, logger *zap.SugaredLogger) {
	defer wg.Done()
	c := make(chan os.Signal, 1)
	signal.Notify(c,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGSEGV,
		syscall.SIGABRT,
		syscall.SIGILL,
		syscall.SIGFPE)
	signal := <-c
	logger.Warnf("Signal (%v) Received, Shutting Down", signal)
	cancelFn()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	// init log
	plainLogger, err := zap.NewProduction()
	runtime.Must(err)
	logger := plainLogger.Sugar()

	// init client
	k8sClient, err := k8s.NewClient(logger)
	runtime.Must(err)

	wg.Add(1)
	go signalHandler(cancel, &wg, logger)

	// init remediators
	podRemediator, err := remediator.NewPodRemediator(logger, k8sClient)
	if err != nil {
		logger.Panic("Error initializing Pod remediator: ", zap.Error(err))
	}

	wg.Add(1)
	logger.Info("Starting CrashLoopBackOffRescheduler")
	go podRemediator.Run(ctx, &wg)

	wg.Add(1)
	logger.Info("Starting Http listener")
	go Server(ctx, &wg, logger)

	wg.Wait()
}

func Server(ctx context.Context, wg *sync.WaitGroup, logger *zap.SugaredLogger) {
	defer wg.Done()
	//register handler
	mux := http.NewServeMux()
	healthz.RegisterHandler(mux)
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Error listening: ", zap.Error(err))
		}
	}()
	<-ctx.Done()
	logger.Info("Received signal to stop")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
