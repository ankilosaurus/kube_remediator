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
// TODO: is there no better way of doing this ?
func signalHandler(cancelFn func(), wg *sync.WaitGroup, logger *zap.Logger) {
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
	logger.Sugar().Warnf("Signal %v Received, Shutting Down", signal) // TODO: prefer structured logging
	cancelFn()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup

	// build a logger without timestamps because docker already logs with timestamps / use "message"
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	loggerConfig.EncoderConfig.MessageKey = "message"
	logger, err := loggerConfig.Build()
	runtime.Must(err)

	// init client
	k8sClient, err := k8s.NewClient(logger)
	runtime.Must(err)

	wg.Add(1)
	go signalHandler(cancel, &wg, logger)

	// init remediators
	remediator, err := remediator.NewCrashLoopBackOffRescheduler(logger, k8sClient)
	if err != nil {
		logger.Panic("Error initializing CrashLoopBackOffRescheduler", zap.Error(err))
	}

	wg.Add(1)
	go remediator.Run(ctx, &wg)

	wg.Add(1)
	go HealthCheck(ctx, &wg, logger)

	wg.Wait()
}

// allow checking from the outside if the app is still running
// eventually this should show if the remediators are working
// maybe later also for /metrics
// TODO: own file with own class and same Run interface
func HealthCheck(ctx context.Context, wg *sync.WaitGroup, logger *zap.Logger) {
	defer wg.Done()

	logger.Info("Starting")

	//register handler
	mux := http.NewServeMux()
	healthz.RegisterHandler(mux)
	srv := &http.Server{Addr: "localhost:8080", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			logger.Error("Error listening", zap.Error(err))
		}
	}()
	<-ctx.Done()
	logger.Info("Stopping", zap.String("reason", "Signal"))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
