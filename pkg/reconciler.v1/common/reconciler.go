// Copyright 2021 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"context"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	commonutil "github.com/kubeflow/common/pkg/util"
)

type KubeflowReconcilerConfig struct {
	// Enable gang scheduling by volcano
	EnableGangScheduling bool
}

func (r *KubeflowReconciler) GangSchedulingEnabled() bool {
	return r.Config.EnableGangScheduling
}

func NewDefaultKubeflowReconcilerConfig() *KubeflowReconcilerConfig {
	return &KubeflowReconcilerConfig{EnableGangScheduling: true}
}

// KubeflowReconciler reconciles a KubeflowJob object
type KubeflowReconciler struct {
	KubeflowReconcilerInterface
	client.Client
	APIReader client.Reader
	Scheme    *runtime.Scheme
	Log       logr.Logger
	recorder  record.EventRecorder
	counter   *commonutil.Counter
	Config    *KubeflowReconcilerConfig
}

func NewKubeflowReconciler(mgr manager.Manager) *KubeflowReconciler {
	return &KubeflowReconciler{
		Client:    mgr.GetClient(),
		APIReader: mgr.GetAPIReader(),
		Scheme:    mgr.GetScheme(),
		Log:       log.Log,
		recorder:  mgr.GetEventRecorderFor(ReconcilerName),
		counter:   commonutil.NewCounter(),
		Config:    NewDefaultKubeflowReconcilerConfig(),
	}
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *KubeflowReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	job, err := r.GetJob(ctx, req)
	if err != nil {
		return ctrl.Result{}, err
	}

	logger := r.GetLogger(job)

	if job.GetDeletionTimestamp() != nil {
		logger.Info(MsgReconcileCancelled, ReasonKey, ReasonJobDeleted)
		return ctrl.Result{}, nil
	}

	scheme.Scheme.Default(job)

	// Get rid of SatisfiedExpectation
	replicasSpec, err := r.ExtractReplicasSpec(job)
	if err != nil {
		return ctrl.Result{}, err
	}

	runPolicy, err := r.ExtractRunPolicy(job)
	if err != nil {
		return ctrl.Result{}, err
	}

	status, err := r.ExtractJobStatus(job)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.ReconcileJob(ctx, job, replicasSpec, status, runPolicy)
	if err != nil {
		logger.Info("Reconcile PyTorch Job error %v", err)
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeflowReconciler) SetupWithManager(mgr ctrl.Manager, obj client.Object) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(obj).
		Owns(&corev1.Pod{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
