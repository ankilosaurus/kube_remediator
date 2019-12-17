package main

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/http"
	"github.com/aksgithub/kube_remediator/pkg/k8s"
	"github.com/aksgithub/kube_remediator/pkg/remediator"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/runtime"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"sync"
	"syscall"
)

// catch interrupts to gracefully exit since otherwise goroutines get killed without running defer
// TODO: is there no better way of doing this ?
func signalHandler(cancelFn func(), wg *sync.WaitGroup, logger *zap.Logger) {
	defer cancelFn()
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
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	// build a logger:
	// - without timestamps because docker already logs with timestamps
	// - use "message" instead of "msg" for consistency with other services / datadog parsing
	// - remove caller since it points to shared methods most of the time anyway
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	loggerConfig.EncoderConfig.MessageKey = "message"
	loggerConfig.DisableCaller = true

	// general logger
	logger, err := loggerConfig.Build()
	runtime.Must(err)

	wg.Add(1)
	go signalHandler(cancel, &wg, logger)

	remediators := []remediator.BaseIntf{
		&remediator.OldPodDeleter{},
		&remediator.CrashLoopBackOffRescheduler{},
	}

	for _, r := range remediators {
		// remediator.OldPodDeleter -> OldPodDeleter
		name := strings.Split(reflect.TypeOf(r).String(), ".")[1]

		// make each logged line show what remediator it came from
		loggerConfig.InitialFields = map[string]interface{}{"remediator": name}

		logger, err := loggerConfig.Build()
		runtime.Must(err)

		k8sClient, err := k8s.NewClient(logger)
		runtime.Must(err)

		err = r.Setup(logger, k8sClient)
		if err != nil {
			logger.Panic("Error initializing", zap.Error(err))
		}

		wg.Add(1)
		go r.Run(ctx, &wg)
	}

	wg.Add(1)
	go http.NewServer(logger).Serve(ctx, &wg)

	<-ctx.Done()
	wg.Wait()
}
