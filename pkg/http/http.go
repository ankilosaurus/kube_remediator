package http

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/healthz"
	"github.com/aksgithub/kube_remediator/pkg/metrics"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type Server struct {
	logger *zap.Logger
}

func NewServer(logger *zap.Logger) *Server {
	return &Server{logger: logger}
}

// allow checking from the outside if the app is still running
// eventually this should show if the remediators are working
// maybe later also for /metrics
func (s *Server) Serve(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	s.logger.Info("Starting")

	//register sandler
	mux := http.NewServeMux()
	healthz.RegisterHandler(mux)
	metrics.RegisterHandler(mux)
	srv := &http.Server{Addr: ":8080", Handler: mux}

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			s.logger.Error("Error listening", zap.Error(err))
		}
	}()
	<-ctx.Done()
	s.logger.Info("Stopping", zap.String("reason", "Signal"))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
