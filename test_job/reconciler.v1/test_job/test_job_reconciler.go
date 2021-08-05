package test_job

import (
	"context"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	common_reconciler "github.com/kubeflow/common/pkg/reconciler.v1/common"
	v1 "github.com/kubeflow/common/test_job/apis/test_job/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ common_reconciler.KubeflowReconcilerInterface = &TestReconciler{}

type TestReconciler struct {
	common_reconciler.KubeflowReconciler
	Job      *v1.TestJob
	Pods     []*corev1.Pod
	Services []*corev1.Service
	PodGroup client.Object
	Cache    []client.Object
}

func (r *TestReconciler) GetJob(ctx context.Context, req ctrl.Request) (client.Object, error) {
	return r.Job, nil
}

func (r *TestReconciler) GetDefaultContainerName() string {
	return v1.DefaultContainerName
}

func (r *TestReconciler) GetPodGroupForJob(ctx context.Context, job client.Object) (client.Object, error) {
	return r.PodGroup, nil
}

func (r *TestReconciler) GetPodsForJob(ctx context.Context, job client.Object) ([]*corev1.Pod, error) {
	return r.Pods, nil
}

func (r *TestReconciler) GetServicesForJob(ctx context.Context, job client.Object) ([]*corev1.Service, error) {
	return r.Services, nil
}

func (r *TestReconciler) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) error {
	r.Cache = append(r.Cache, obj)
	return nil
}

func (r *TestReconciler) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	for idx, o := range r.Cache {
		if o.GetName() == obj.GetName() && o.GetNamespace() == obj.GetNamespace() && o.GetObjectKind() == obj.GetObjectKind() {
			r.Cache = append(r.Cache[:idx], r.Cache[idx+1:]...)
			return nil
		}
	}
	return errors.NewNotFound(schema.GroupResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Resource: obj.GetSelfLink(),
	}, obj.GetName())
}

func (r *TestReconciler) Update(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
	for idx, o := range r.Cache {
		if o.GetName() == obj.GetName() && o.GetNamespace() == obj.GetNamespace() && o.GetObjectKind() == obj.GetObjectKind() {
			r.Cache[idx] = obj
			return nil
		}
	}
	return errors.NewNotFound(schema.GroupResource{
		Group:    obj.GetObjectKind().GroupVersionKind().Group,
		Resource: obj.GetSelfLink(),
	}, obj.GetName())
}

func (r *TestReconciler) ExtractReplicasSpec(job client.Object) (map[commonv1.ReplicaType]*commonv1.ReplicaSpec, error) {
	tj := job.(*v1.TestJob)

	rs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
	for k, v := range tj.Spec.TestReplicaSpecs {
		rs[commonv1.ReplicaType(k)] = v
	}

	return rs, nil
}

func (r *TestReconciler) ExtractRunPolicy(job client.Object) (*commonv1.RunPolicy, error) {
	tj := job.(*v1.TestJob)

	return tj.Spec.RunPolicy, nil
}

func (r *TestReconciler) ExtractJobStatus(job client.Object) (*commonv1.JobStatus, error) {
	tj := job.(*v1.TestJob)

	return &tj.Status, nil
}

func (r *TestReconciler) IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rtype commonv1.ReplicaType, index int) bool {
	return string(rtype) == string(v1.TestReplicaTypeMaster)
}
