package reconciler

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type BaseReconciler interface {
	ReconcileJob(job interface{}) error
	GetLoggerForKey(key string)
	GetPodsForJob(job metav1.Object)
	IsJobSucceeded(status *commonv1.JobStatus) bool
	IsJobFailed(status *commonv1.JobStatus) bool
	CleanupJob(job metav1.Object) error
	GetTotalReplicas(job metav1.Object) error
	UpdateJobStatus(status *commonv1.JobStatus) error
	// ReconcilePods checks and updates pods for each given ReplicaSpec.
	// It will requeue the job in case of an error while creating/deleting pods.
	// Common implementation will be provided and User can still override this to implement their own reconcile logic
	ReconcilePods(job interface{}, jobStatus *commonv1.JobStatus, pods []*v1.Pod, rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error

	// ReconcileServices checks and updates services for each given ReplicaSpec.
	// It will requeue the job in case of an error while creating/deleting services.
	// Common implementation will be provided and User can still override this to implement their own reconcile logic
	ReconcileServices(job metav1.Object, services []*v1.Service, rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec) error
}