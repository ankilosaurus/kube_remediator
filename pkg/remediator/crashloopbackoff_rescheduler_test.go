package remediator_test

import (
	"context"
	"github.com/aksgithub/kube_remediator/pkg/k8s/mock"
	"github.com/aksgithub/kube_remediator/pkg/metrics"
	"github.com/aksgithub/kube_remediator/pkg/remediator"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

type TestCrashLoopBackOffReschedulerSuite struct {
	suite.Suite
	ctx            context.Context
<<<<<<< HEAD
=======
	t              *testing.T
>>>>>>> 054422e... unit-test crashloopbackoff_rescheduler
	logger         *zap.Logger
	mockController *gomock.Controller
	mockClient     *mock_k8s.MockClientInterface
	pods           []corev1.Pod
<<<<<<< HEAD
	remediator     *remediator.CrashLoopBackOffRescheduler
	t              *testing.T
=======
>>>>>>> 054422e... unit-test crashloopbackoff_rescheduler
}

func (suite *TestCrashLoopBackOffReschedulerSuite) SetupSuite() {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.EncoderConfig.TimeKey = ""
	loggerConfig.EncoderConfig.MessageKey = "message"
	logger, err := loggerConfig.Build()
<<<<<<< HEAD
	assert.Equal(suite.t, err, nil)
=======
	runtime.Must(err)
>>>>>>> 054422e... unit-test crashloopbackoff_rescheduler

	suite.ctx = context.Background()
	suite.logger = logger
	suite.mockController = gomock.NewController(suite.t)
	suite.mockClient = mock_k8s.NewMockClientInterface(suite.mockController)
<<<<<<< HEAD

	suite.remediator, err = suite.testGetRemediator()
	assert.Equal(suite.t, err, nil)
	assert.Assert(suite.t, suite.remediator != nil)
}

func (suite *TestCrashLoopBackOffReschedulerSuite) testGetRemediator() (*remediator.CrashLoopBackOffRescheduler, error) {
	remediator.CONFIG_FILE = "../../config/crash_loop_back_off_rescheduler.json"
	cm := metrics.NewCrashLoopBackOffMetrics(suite.logger)
	cm.RegisterMetrics()
	remediator, err := remediator.NewCrashLoopBackOffRescheduler(suite.logger, suite.mockClient, cm)
	return remediator, err
=======
>>>>>>> 054422e... unit-test crashloopbackoff_rescheduler
}

func (suite *TestCrashLoopBackOffReschedulerSuite) SetupTest() {
	var healthyPod = corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "healthyPod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					Name: "controller",
				},
			},
			Annotations: map[string]string{
				"kube-remediator/CrashLoopBackOffRemediator": "true",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					RestartCount: 0,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					RestartCount: 0,
					State: corev1.ContainerState{
						Running: &corev1.ContainerStateRunning{},
					},
				},
			},
		},
	}
	var unHealthyPod = corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "unHealthyPod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				metav1.OwnerReference{
					Name: "controller",
				},
			},
			Annotations: map[string]string{
				"kube-remediator/CrashLoopBackOffRemediator": "true",
			},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					RestartCount: 2,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				},
			},
			ContainerStatuses: []corev1.ContainerStatus{
				corev1.ContainerStatus{
					RestartCount: 3,
					State: corev1.ContainerState{
						Waiting: &corev1.ContainerStateWaiting{Reason: "CrashLoopBackOff"},
					},
				},
			},
		},
	}
	suite.pods = []corev1.Pod{healthyPod, unHealthyPod}
}

func (suite *TestCrashLoopBackOffReschedulerSuite) testRemediator() {
<<<<<<< HEAD
	ctx, cancel := context.WithCancel(suite.ctx)
	cancel() // cancel first so we can just run once and exit
	suite.remediator.Run(ctx, nil)
=======
	remediator, err := suite.testGetRemediator()
	assert.Equal(suite.t, err, nil)
	assert.Assert(suite.t, remediator != nil)

	ctx, cancel := context.WithCancel(suite.ctx)
	cancel() // cancel first so we can just run once and exit
	remediator.Run(ctx, nil)
}

func (suite *TestCrashLoopBackOffReschedulerSuite) testGetRemediator() (*remediator.CrashLoopBackOffRescheduler, error) {
	remediator.CONFIG_FILE = "../../config/crash_loop_back_off_rescheduler.json"
	remediator, err := remediator.NewCrashLoopBackOffRescheduler(suite.logger, suite.mockClient)
	return remediator, err
>>>>>>> 054422e... unit-test crashloopbackoff_rescheduler
}

// Restart only Unhealthy Pod
func (suite *TestCrashLoopBackOffReschedulerSuite) TestReschedulePods() {
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil).Times(1)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[1]).Return(nil).Times(1)

	suite.testRemediator()
}

// Unhealthy Pod without OwnerReference should not be deleted
func (suite *TestCrashLoopBackOffReschedulerSuite) TestUnHealthyPodWithoutOwnerReference() {
	suite.pods[1].ObjectMeta.OwnerReferences = []metav1.OwnerReference{}
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil).Times(1)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[1]).Return(nil).Times(0)

	suite.testRemediator()
}

// Unhealthy Pod's RestartCount does not meet threshold
func (suite *TestCrashLoopBackOffReschedulerSuite) TestUnHealthyPodRestartCount() {
	suite.pods[1].Status.ContainerStatuses[0].RestartCount = 2
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil).Times(1)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[1]).Return(nil).Times(0)

	suite.testRemediator()
}

// Unhealthy Pod has Init container in CrashLoopBackOff
func (suite *TestCrashLoopBackOffReschedulerSuite) TestUnHealthyPodInitContainerRestartCount() {
	suite.pods[1].Status.ContainerStatuses[0].RestartCount = 2 // reduce
	suite.pods[1].Status.InitContainerStatuses[0].RestartCount = 3
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil).Times(1)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[1]).Return(nil).Times(1)

	suite.testRemediator()
}

// Unhealthy Pod has restartCount with reason != CrashLoopBackOff
func (suite *TestCrashLoopBackOffReschedulerSuite) TestUnHealthyPodWithUnknownReason() {
	suite.pods[1].Status.ContainerStatuses[0].State.Waiting.Reason = "X"
	suite.mockClient.EXPECT().GetPods("").Return(&corev1.PodList{Items: suite.pods}, nil).Times(1)
	suite.mockClient.EXPECT().DeletePod(&suite.pods[1]).Return(nil).Times(0)

	suite.testRemediator()
}

func (suite *TestCrashLoopBackOffReschedulerSuite) TearDownSuite() {
	defer suite.mockController.Finish()
}

func TestSuiteCrashLoopBackOffRescheduler(t *testing.T) {
	suite.Run(t, &TestCrashLoopBackOffReschedulerSuite{t: &testing.T{}})
}
