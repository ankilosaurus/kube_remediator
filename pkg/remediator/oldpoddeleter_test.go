package remediator_test

import (
	"context"
	"errors"
	"github.com/aksgithub/kube_remediator/pkg/k8s/mock"
	"github.com/aksgithub/kube_remediator/pkg/remediator"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sync"
	"testing"
	"time"
)

type TestOldPodDeleterSuite struct {
	suite.Suite
	logger         *zap.Logger
	mockController *gomock.Controller
	mockClient     *mock_k8s.MockClientInterface
	pods           []corev1.Pod
	t              *testing.T
}

func TestSuiteOldPodDeleter(t *testing.T) {
	suite.Run(t, &TestOldPodDeleterSuite{t: t})
}

func (suite *TestOldPodDeleterSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	suite.mockController = gomock.NewController(suite.t)
	suite.mockClient = mock_k8s.NewMockClientInterface(suite.mockController)
	suite.pods = []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "foo",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-25 * time.Hour)),
			Labels: map[string]string{
				"kube-remediator/OldPodDeleter": "true",
			},
		},
	}}
}

func (suite *TestOldPodDeleterSuite) TeardownTest() {
	suite.mockController.Finish()
}

func (suite *TestOldPodDeleterSuite) run() {
	oldPodDeleter, err := remediator.NewOldPodDeleter(suite.logger, suite.mockClient)
	assert.Equal(suite.t, err, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel first so we can just run once and exit

	var wg sync.WaitGroup
	wg.Add(1)

	oldPodDeleter.Run(ctx, &wg)
}

func (suite *TestOldPodDeleterSuite) TestDeletesOldPods() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil)
	suite.run()
}

func (suite *TestOldPodDeleterSuite) TestKeepsNewPods() {
	suite.pods[0].ObjectMeta.CreationTimestamp = metav1.NewTime(time.Now().Add(-23 * time.Hour))
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestOldPodDeleterSuite) TestDoesNotCrashWhenListFails() {
	suite.mockClient.EXPECT().GetPods("").Return(nil, errors.New("Foo"))
	suite.run()
}

func (suite *TestOldPodDeleterSuite) TestDoesNotCrashWhenDeleteFails() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(errors.New("Foo"))
	suite.run()
}
