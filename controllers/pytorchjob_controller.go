/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"

	commonapiv1 "github.com/kubeflow/common/pkg/apis/common/v1"

	pytorchv1 "github.com/kubeflow/common/apis/pytorch/v1"
	"github.com/kubeflow/common/pkg/reconciler.v1/common"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// PyTorchJobReconciler reconciles a PyTorchJob object
type PyTorchJobReconciler struct {
	client.Client

	common.CommonReconciler

	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubeflow.org,resources=pytorchjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubeflow.org,resources=pytorchjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubeflow.org,resources=pytorchjobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the PyTorchJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *PyTorchJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	logger := r.GetLoggerForKey(req.NamespacedName.String())

	var job pytorchv1.PyTorchJob
	if err := r.Get(ctx, req.NamespacedName, &job); err != nil {
		logger.Error(err, "unable to fetch PyTorchJob")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	oldStatus := job.Status.DeepCopy()
	pods, err := r.GetPodsForJob(ctx, &job)
	if err != nil {
		logger.Warnf("getPodsForPyTorchJob error %v", err)
		return ctrl.Result{}, err
	}

	services, err := r.GetServicesForJob(ctx, &job)
	if err != nil {
		logger.Warnf("getServicesForPyTorchJob error %v", err)
		return ctrl.Result{}, err
	}

	// If the PyTorchJob is terminated, delete all pods and services.
	if r.IsJobSucceeded(&job.Status) || r.IsJobFailed(&job.Status) {
		if err = r.CleaupJob(ctx, &job); err != nil {
			return ctrl.Result{}, err
		}

		if r.IsGangSchedulingEnabled() {
			if err = r.DeleteGangSchedulingResources(ctx, &job); err != nil {
				return ctrl.Result{}, err
			}
		}

		// At this point the pods may have been deleted, so if the job succeeded, we need to manually set the replica status.
		// If any replicas are still Active, set their status to succeeded.
		if r.IsJobSucceeded(&job.Status) {
			for rtype := range job.Status.ReplicaStatuses {
				job.Status.ReplicaStatuses[rtype].Succeeded += job.Status.ReplicaStatuses[rtype].Active
				job.Status.ReplicaStatuses[rtype].Active = 0
			}
		}
		if !apiequality.Semantic.DeepEqual(*oldStatus, job.Status) {
			err = r.UpdateJobStatus(ctx, &job)
			if err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	// Reconcile PodGroup
	if r.IsGangSchedulingEnabled() {
		err = r.ReconcileGangSchedulingResources(ctx, &job)
		if err != nil {
			logger.Warnf("Sync PodGroup %v: %v", job.Name, err)
		}
	}

	// Save the current state of the replicas
	replicasStatus := make(map[commonapiv1.ReplicaType]*commonapiv1.ReplicaSpec)
	for rtype, spec := range job.Spec.PyTorchReplicaSpecs {
		replicasStatus[commonapiv1.ReplicaType(rtype)] = spec
	}

	// Reconcile Pods
	for rtype, rspec := range replicasStatus {
		err = r.ReconcilePods(ctx, &job, pods, rtype, replicasStatus)
		if err != nil {
			logger.Warnf("reconcile Pods error %v", err)
			return ctrl.Result{}, err
		}

		// Service is in need only for Master
		if string(rtype) == string(pytorchv1.PyTorchReplicaTypeMaster) {
			err = r.ReconcileServices(ctx, &job, services, rtype, rspec)
			if err != nil {
				logger.Warnf("reconcileServices error %v", err)
				return ctrl.Result{}, err
			}
		}
	}

	if !apiequality.Semantic.DeepEqual(*oldStatus, job.Status) {
		err = r.UpdateJobStatus(ctx, &job)
		if err != nil {

		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PyTorchJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pytorchv1.PyTorchJob{}).
		Complete(r)
}
