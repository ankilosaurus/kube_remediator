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
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes/fake"
	"sync"
	"testing"
	"time"
)

type TestFailedPodReschedulerSuite struct {
	suite.Suite
	logger         *zap.Logger
	mockController *gomock.Controller
	mockClient     *mock_k8s.MockClientInterface
	pods           []corev1.Pod
	t              *testing.T
}

func TestSuiteFailedPodRescheduler(t *testing.T) {
	suite.Run(t, &TestFailedPodReschedulerSuite{t: t})
}

func (suite *TestFailedPodReschedulerSuite) SetupTest() {
	suite.logger, _ = zap.NewDevelopment()
	suite.mockController = gomock.NewController(suite.t)
	suite.mockClient = mock_k8s.NewMockClientInterface(suite.mockController)
	suite.pods = []corev1.Pod{{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "healthyPod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{
					Name: "controller",
				},
			},
			CreationTimestamp: metav1.Time{time.Now().Add(-10 * time.Minute)},
		},
		Status: corev1.PodStatus{
			Phase:  "Failed",
			Reason: "OutOfcpu",
		},
	}}
}

func (suite *TestFailedPodReschedulerSuite) TearDownTest() {
	suite.mockController.Finish()
}

func (suite *TestFailedPodReschedulerSuite) newInformerFactory() informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(fake.NewSimpleClientset(), 0, informers.WithNamespace(""))
}

func (suite *TestFailedPodReschedulerSuite) run() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel first so we can just run once and exit

	suite.mockClient.EXPECT().NewSharedInformerFactory("").Return(suite.newInformerFactory(), nil)
	r := remediator.FailedPodRescheduler{}
	err := r.Setup(suite.logger, suite.mockClient)
	assert.Equal(suite.t, err, nil)

	var wg sync.WaitGroup
	wg.Add(1)
	r.Run(ctx, &wg)
}

func (suite *TestFailedPodReschedulerSuite) TestReschedulesFailedPod() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil)
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestLoopsOverAllPods() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: append(suite.pods, suite.pods...)}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil).Times(2)
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestKeepsFailedPodWithoutOwnerReference() {
	suite.pods[0].ObjectMeta.OwnerReferences = []metav1.OwnerReference{}
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestKeepsFailedPodsWhenTheyAreCleanup() {
	suite.pods[0].ObjectMeta.OwnerReferences[0].Kind = "Job"
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestKeepsFailedPodsWithOtherReasons() {
	suite.pods[0].Status.Reason = "fake"
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestDoesNotCrashWhenDeleteFails() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(errors.New("foo"))
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestDoesNotCrashWhenListFails() {
	suite.mockClient.EXPECT().GetPods("").Return(nil, errors.New("foo"))
	suite.run()
}

func (suite *TestFailedPodReschedulerSuite) TestDoesNotDeleteWhenPodIsNew() {
	suite.pods[0].CreationTimestamp = metav1.Time{time.Now().Add(-4 * time.Minute)}
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}
