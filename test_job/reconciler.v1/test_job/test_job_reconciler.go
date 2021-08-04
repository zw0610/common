package test_job

import (
	"context"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"

	common_reconciler "github.com/kubeflow/common/pkg/reconciler.v1/common"
	v1 "github.com/kubeflow/common/test_job/apis/test_job/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ common_reconciler.KubeflowReconcilerInterface = &TestReconciler{}

type TestReconciler struct {
	common_reconciler.KubeflowReconcilerInterface
	Job *v1.TestJob
}

func (r *TestReconciler) GetJob(ctx context.Context, req ctrl.Request) (client.Object, error) {
	return r.Job, nil
}

func (r *TestReconciler) ExtractReplicasSpec(job client.Object) (map[commonv1.ReplicaType]*commonv1.ReplicaSpec, error) {
	rs := map[commonv1.ReplicaType]*commonv1.ReplicaSpec{}
	for k, v := range r.Job.Spec.TestReplicaSpecs {
		rs[commonv1.ReplicaType(k)] = v
	}

	return rs, nil
}

func (r *TestReconciler) ExtractRunPolicy(job client.Object) (*commonv1.RunPolicy, error) {
	return r.Job.Spec.RunPolicy, nil
}

func (r *TestReconciler) ExtractJobStatus(job client.Object) (*commonv1.JobStatus, error) {
	return &r.Job.Status, nil
}

func (r *TestReconciler) IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rtype commonv1.ReplicaType, index int) bool {
	return string(rtype) == string(v1.TestReplicaTypeMaster)
}
