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

type TestCompletedPodDeleterSuite struct {
	suite.Suite
	logger         *zap.Logger
	mockController *gomock.Controller
	mockClient     *mock_k8s.MockClientInterface
	pods           []corev1.Pod
	t              *testing.T
}

func TestSuitecompletedPodDeleter(t *testing.T) {
	suite.Run(t, &TestCompletedPodDeleterSuite{t: t})
}

func (suite *TestCompletedPodDeleterSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	suite.mockController = gomock.NewController(suite.t)
	suite.mockClient = mock_k8s.NewMockClientInterface(suite.mockController)
	suite.pods = []corev1.Pod{{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "foo",
			Namespace:         "default",
			CreationTimestamp: metav1.NewTime(time.Now().Add(-25 * time.Hour)),
		},
		Status: corev1.PodStatus{
			Phase:  "Completed",
		},
	}}
}

func (suite *TestCompletedPodDeleterSuite) TeardownTest() {
	suite.mockController.Finish()
}

func (suite *TestCompletedPodDeleterSuite) run() {
	completedPodDeleter := remediator.CompletedPodDeleter{}
	err := completedPodDeleter.Setup(suite.logger, suite.mockClient)
	assert.Equal(suite.t, err, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel first so we can just run once and exit

	var wg sync.WaitGroup
	wg.Add(1)

	completedPodDeleter.Run(ctx, &wg)
}

func (suite *TestCompletedPodDeleterSuite) TestDeleteCompletedPods() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil)
	suite.run()
}

func (suite *TestCompletedPodDeleterSuite) TestKeepsNewPods() {
	suite.pods[0].ObjectMeta.CreationTimestamp = metav1.NewTime(time.Now().Add(-23 * time.Hour))
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestCompletedPodDeleterSuite) TestDoesNotCrashWhenListFails() {
	suite.mockClient.EXPECT().GetPods("").Return(nil, errors.New("Foo"))
	suite.run()
}

func (suite *TestCompletedPodDeleterSuite) TestDoesNotCrashWhenDeleteFails() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(errors.New("Foo"))
	suite.run()
}
