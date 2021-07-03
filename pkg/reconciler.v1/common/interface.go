package common

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CommonReconcilerInterface interface {
	GetLoggerForKey(key string) *logrus.Entry
	GetPodsForJob(ctx context.Context, job metav1.Object) ([]*corev1.Pod, error)
	GetServicesForJob(ctx context.Context, job metav1.Object) ([]*corev1.Service, error)
	IsJobSucceeded(status *commonv1.JobStatus) bool
	IsJobFailed(status *commonv1.JobStatus) bool
	IsGangSchedulingEnabled() bool
	CleaupJob(ctx context.Context, job client.Object) error
	DeleteGangSchedulingResources(ctx context.Context, job metav1.Object) error
	ReconcileGangSchedulingResources(ctx context.Context, job metav1.Object) error
	UpdateJobStatus(ctx context.Context, job client.Object) error
	// ReconcilePods checks and updates pods for each given ReplicaSpec.
	// It will requeue the job in case of an error while creating/deleting pods.
	// Common implementation will be provided and User can still override this to implement their own reconcile logic
	ReconcilePods(ctx context.Context, job metav1.Object, pods []*v1.Pod, rtype commonv1.ReplicaType,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error

	// ReconcileServices checks and updates services for each given ReplicaSpec.
	// It will requeue the job in case of an error while creating/deleting services.
	// Common implementation will be provided and User can still override this to implement their own reconcile logic
	ReconcileServices(ctx context.Context, job metav1.Object, services []*v1.Service, rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec) error

	GetExpectedServicesForReplicaType(job metav1.Object, rtype commonv1.ReplicaType) ([]*corev1.Service, error)
}
