package http_test

import (
	"context"
	remediator_http "github.com/aksgithub/kube_remediator/pkg/http"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gotest.tools/assert"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/runtime"
	"net/http"
	"sync"
	"testing"
)

type TestHttpServerSuite struct {
	suite.Suite
	ctx    context.Context
	t      *testing.T
	logger *zap.Logger
}

func (suite *TestHttpServerSuite) SetupSuite() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	loggerConfig.EncoderConfig.MessageKey = "message"
	logger, err := loggerConfig.Build()
	runtime.Must(err)

	suite.ctx = context.Background()
	suite.logger = logger
}

func (suite *TestHttpServerSuite) httpGet(url string) (int, string) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req, err := http.NewRequest("GET", url, nil)
	response, err := client.Do(req)
	if err != nil {
		suite.logger.Sugar().Infof("Error: %s\n%+v\n", err.Error(), response)
		return 0, ""
	}
	defer response.Body.Close()
	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return 0, ""
	}
	return response.StatusCode, string(b)
}

func (suite *TestHttpServerSuite) TestServer() {
	ctx, cancel := context.WithCancel(suite.ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go remediator_http.NewServer(suite.logger).Serve(ctx, &wg)

	status, _ := suite.httpGet("http://localhost:8080/healthz")
	assert.Equal(suite.t, status, 200)
	status, _ = suite.httpGet("http://localhost:8080/metrics")
	assert.Equal(suite.t, status, 200)

	cancel()
	wg.Wait()
}

func TestHttpServer(t *testing.T) {
	suite.Run(t, &TestHttpServerSuite{t: &testing.T{}})
}
