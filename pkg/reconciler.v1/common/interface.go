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
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
)

type ReconcilerUtilInterface interface {
	GetRecorder() record.EventRecorder
	GetLogger() logr.Logger
	GetScheme() *runtime.Scheme
}

type GangSchedulingInterface interface {
	ReconcilerUtilInterface

	GetGangSchedulerName() string
	GetPodGroupName(job client.Object) string
	GetPodGroupForJob(ctx context.Context, job client.Object) (client.Object, error)
	DeletePodGroup(ctx context.Context, job client.Object) error
	ReconcilePodGroup(ctx context.Context, job client.Object, runPolicy *commonv1.RunPolicy,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error
	DecoratePodForGangScheduling(rtype commonv1.ReplicaType, podTemplate *corev1.PodTemplateSpec, job client.Object)
}

type PodInterface interface {
	ReconcilerUtilInterface

	GenPodName(jobName string, rtype commonv1.ReplicaType, index string) string
	GetDefaultContainerName() string
	GetPodsForJob(ctx context.Context, job client.Object) ([]*corev1.Pod, error)
	FilterPodsForReplicaType(pods []*corev1.Pod, replicaType commonv1.ReplicaType) ([]*corev1.Pod, error)
	GetPodSlices(pods []*corev1.Pod, replicas int, logger *logrus.Entry) [][]*corev1.Pod
	ReconcilePods(
		ctx context.Context,
		job client.Object,
		jobStatus *commonv1.JobStatus,
		pods []*corev1.Pod,
		rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error
	CreateNewPod(job client.Object, rt commonv1.ReplicaType, index string,
		spec *commonv1.ReplicaSpec, masterRole bool, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error
	DeletePod(ctx context.Context, ns string, name string) error
	DecoratePod(rtype commonv1.ReplicaType, podTemplate *corev1.PodTemplateSpec, job client.Object)
}

type ServiceInterface interface {
	ReconcilerUtilInterface

	GetPortsFromJob(spec *commonv1.ReplicaSpec) (map[string]int32, error)
	GetServicesForJob(ctx context.Context, job client.Object) ([]*corev1.Service, error)
	FilterServicesForReplicaType(services []*corev1.Service,
		replicaType commonv1.ReplicaType) ([]*corev1.Service, error)
	GetServiceSlices(services []*corev1.Service, replicas int, logger *logrus.Entry) [][]*corev1.Service
	ReconcileServices(
		job client.Object,
		services []*corev1.Service,
		rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec) error
	CreateNewService(job metav1.Object, rtype commonv1.ReplicaType,
		spec *commonv1.ReplicaSpec, index string) error
	DeleteService(ns string, name string, job client.Object) error
	DecorateService(rtype commonv1.ReplicaType, podTemplate *corev1.PodTemplateSpec, job client.Object)
}

type JobInterface interface {
	ReconcilerUtilInterface

	PodInterface
	ServiceInterface
	GangSchedulingInterface

	GetReconcilerName() string
	GetGroupNameLabelValue() string

	// GetJob MUST be overridden to get jobs with specified kind
	GetJob(ctx context.Context, req ctrl.Request) (client.Object, error)

	// ExtractReplicasSpec MUST be overridden to extract ReplicasSpec from a job
	ExtractReplicasSpec(job client.Object) (map[commonv1.ReplicaType]*commonv1.ReplicaSpec, error)

	// ExtractRunPolicy MUST be overridden to extract the pointer of RunPolicy from a job
	ExtractRunPolicy(job client.Object) (*commonv1.RunPolicy, error)

	// ExtractJobStatus MUST be overridden to extract the pointer of JobStatus from a job
	ExtractJobStatus(job client.Object) (*commonv1.JobStatus, error)

	// IsMasterRole MUST be overridden to determine whether this ReplicaType with index specified is a master role.
	// MasterRole pod will have "job-role=master" set in its label
	IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rtype commonv1.ReplicaType, index int) bool

	ReconcileJob(
		ctx context.Context,
		job client.Object,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
		status *commonv1.JobStatus,
		runPolicy *commonv1.RunPolicy) error
	DeleteJob(job client.Object) error
	UpdateJobStatus(
		job client.Object,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec,
		jobStatus *commonv1.JobStatus) error
	UpdateJobStatusInAPIServer(ctx context.Context, job client.Object) error
	CleanupResources(runPolicy *commonv1.RunPolicy, status commonv1.JobStatus, job client.Object) error
	CleanupJob(runPolicy *commonv1.RunPolicy, status commonv1.JobStatus, job client.Object) error

	RecordAbnormalPods(activePods []*corev1.Pod, object client.Object)
	SetStatusForSuccessJob(status *commonv1.JobStatus)

	IsFlagReplicaTypeForJobStatus(rtype commonv1.ReplicaType) bool
	IsJobSucceeded(status commonv1.JobStatus) bool
	IsJobFailed(status commonv1.JobStatus) bool
	ShouldCleanUp(status commonv1.JobStatus) bool
	PastBackoffLimit(jobName string, runPolicy *commonv1.RunPolicy,
		replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, pods []*corev1.Pod) (bool, error)
	PastActiveDeadline(runPolicy *commonv1.RunPolicy, jobStatus *commonv1.JobStatus) bool
}

type KubeflowReconcilerInterface interface {
	ReconcilerUtilInterface

	JobInterface

	Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error)
	SetupWithManager(mgr ctrl.Manager, obj client.Object) error
}
