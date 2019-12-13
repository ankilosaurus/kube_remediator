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
)

type TestCrashLoopBackOffReschedulerSuite struct {
	suite.Suite
	logger         *zap.Logger
	mockController *gomock.Controller
	mockClient     *mock_k8s.MockClientInterface
	pods           []corev1.Pod
	t              *testing.T
}

func TestSuiteCrashLoopBackOffRescheduler(t *testing.T) {
	suite.Run(t, &TestCrashLoopBackOffReschedulerSuite{t: t})
}

func (suite *TestCrashLoopBackOffReschedulerSuite) SetupTest() {
	remediator.CONFIG_FILE = "../../config/crash_loop_back_off_rescheduler.json" // TODO: set on the instance or global folder
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
			Annotations: map[string]string{
				"kube-remediator/CrashLoopBackOffRemediator": "true",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				{
					RestartCount: 0,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				{
					RestartCount: 6,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				},
			},
		},
	}}
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TearDownTest() {
	suite.mockController.Finish()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) newInformerFactory() informers.SharedInformerFactory {
	return informers.NewSharedInformerFactoryWithOptions(fake.NewSimpleClientset(), 0, informers.WithNamespace(""))
}

func (suite *TestCrashLoopBackOffReschedulerSuite) run() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel first so we can just run once and exit

	suite.mockClient.EXPECT().NewSharedInformerFactory("").Return(suite.newInformerFactory(), nil)
	crashloop, err := remediator.NewCrashLoopBackOffRescheduler(suite.logger, suite.mockClient)
	assert.Equal(suite.t, err, nil)
	assert.Assert(suite.t, crashloop != nil)

	var wg sync.WaitGroup
	wg.Add(1)

	crashloop.Run(ctx, &wg)
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestReschedulesUnhealthyPod() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestLoopsOverAllPods() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: append(suite.pods, suite.pods...)}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil).Times(2)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestKeepsUnhealthyPodWithoutOwnerReference() {
	suite.pods[0].ObjectMeta.OwnerReferences = []metav1.OwnerReference{}
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestKeepsPodBelowThreshold() {
	suite.pods[0].Status.ContainerStatuses[0].RestartCount = 5
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestReschedulesBasedOnInitContainers() {
	suite.pods[0].Status.ContainerStatuses[0].RestartCount = 0 // make healthy
	suite.pods[0].Status.InitContainerStatuses[0].RestartCount = 6
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(nil)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestKeepsWithOtherReason() {
	suite.pods[0].Status.ContainerStatuses[0].State.Waiting.Reason = "X"
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestDoesNotCrashWhenDeleteFails() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[0]).Return(errors.New("Foo"))
	suite.run()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TestDoesNotCrashWhenListFails() {
	suite.mockClient.EXPECT().GetPods("").Return(nil, errors.New("Foo"))
	suite.run()
}
