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
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	"github.com/kubeflow/common/pkg/core"
	commonutil "github.com/kubeflow/common/pkg/util"
	trainutil "github.com/kubeflow/common/pkg/util/train"
)

const DefaultContainerName = "kubeflow"

var (
	// Prometheus metrics
	createdPodsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reconciler_created_pods_total",
		Help: "The total number of created pods",
	})
	deletedPodsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reconciler_deleted_pods_total",
		Help: "The total number of deleted pods",
	})
	failedPodsCount = promauto.NewCounter(prometheus.CounterOpts{
		Name: "reconciler_failed_pods_total",
		Help: "The total number of failed pods",
	})
)

func (r *KubeflowReconciler) GenPodName(jobName string, rtype commonv1.ReplicaType, index string) string {
	return core.GenGeneralName(jobName, rtype, index)
}

func (r *KubeflowReconciler) GetDefaultContainerName() string {
	return DefaultContainerName
}

func (r *KubeflowReconciler) GetPodsForJob(ctx context.Context, job client.Object) ([]*corev1.Pod, error) {
	podList := &corev1.PodList{}
	err := r.List(ctx, podList, client.MatchingLabels(r.GenLabels(job.GetName())))
	if err != nil {
		return nil, err
	}

	var pods []*corev1.Pod = nil
	for _, pod := range podList.Items {
		pods = append(pods, &pod)
	}

	return pods, nil
	// TODO: (zw0610) adding controller reference management
	//// If any adoptions are attempted, we should first recheck for deletion
	//// with an uncached quorum read sometime after listing Pods (see #42639).
	//canAdoptFunc := RecheckDeletionTimestamp(func() (metav1.Object, error) {
	//	fresh := r.EmptyJob()
	//	err = r.APIReader.Get(ctx, types.NamespacedName{Namespace: job.GetNamespace(), Name: job.GetName()}, fresh)
	//	if err != nil {
	//		return nil, err
	//	}
	//	if fresh.GetUID() != job.GetUID() {
	//		return nil, fmt.Errorf("original Job %v/%v is gone: got uid %v, wanted %v", job.GetNamespace(), job.GetName(), fresh.GetUID(), job.GetUID())
	//	}
	//	return fresh, nil
	//})
	//cm := control.NewPodControllerRefManager(jc.PodControl, job, selector, r.GetAPIGroupVersionKind(), canAdoptFunc)
	//return cm.ClaimPods(pods)
}

func (r *KubeflowReconciler) GetPodSlices(pods []*corev1.Pod, replicas int, logger *log.Entry) [][]*corev1.Pod {
	return core.GetPodSlices(pods, replicas, logger)
}

func (r *KubeflowReconciler) FilterPodsForReplicaType(pods []*corev1.Pod, replicaType commonv1.ReplicaType) ([]*corev1.Pod, error) {
	return core.FilterPodsForReplicaType(pods, replicaType)
}

func (r *KubeflowReconciler) ReconcilePods(
	ctx context.Context,
	job client.Object,
	jobStatus *commonv1.JobStatus,
	pods []*corev1.Pod,
	rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {

	// Convert ReplicaType to lower string.
	logger := commonutil.LoggerForReplica(job, rtype)
	// Get all pods for the type rt.
	pods, err := r.FilterPodsForReplicaType(pods, rtype)
	if err != nil {
		return err
	}
	numReplicas := int(*spec.Replicas)
	var masterRole bool

	core.InitializeReplicaStatuses(jobStatus, rtype)

	// GetPodSlices will return enough information here to make decision to add/remove/update resources.
	//
	// For example, let's assume we have pods with replica-index 0, 1, 2
	// If replica is 4, return a slice with size 4. [[0],[1],[2],[]], a pod with replica-index 3 will be created.
	//
	// If replica is 1, return a slice with size 3. [[0],[1],[2]], pod with replica-index 1 and 2 are out of range and will be deleted.
	podSlices := r.GetPodSlices(pods, numReplicas, logger)
	for index, podSlice := range podSlices {
		if len(podSlice) > 1 {
			logger.Warningf("We have too many pods for %s %d", rtype, index)
		} else if len(podSlice) == 0 {
			logger.Infof("Need to create new pod: %s-%d", rtype, index)

			// check if this replica is the master role
			masterRole = r.IsMasterRole(replicas, rtype, index)
			err = r.CreateNewPod(job, rtype, strconv.Itoa(index), spec, masterRole, replicas)
			if err != nil {
				return err
			}
		} else {
			// Check the status of the current pod.
			pod := podSlice[0]

			// check if the index is in the valid range, if not, we should kill the pod
			if index < 0 || index >= numReplicas {
				err = r.DeletePod(ctx, pod.Namespace, pod.Name)
				if err != nil {
					return err
				}
			}

			// Get the exit code of the container.
			var exitCode int32 = 0xbeef // magic number
			for _, status := range pod.Status.ContainerStatuses {
				state := status.State
				if status.Name == r.GetDefaultContainerName() && state.Terminated != nil {
					exitCode = state.Terminated.ExitCode
					logger.Infof("Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
					r.recorder.Eventf(job, corev1.EventTypeNormal, "ExitedWithCode", "Pod: %v.%v exited with code %v", pod.Namespace, pod.Name, exitCode)
				}
			}
			// Check if the pod is retryable.
			if spec.RestartPolicy == commonv1.RestartPolicyExitCode {
				if pod.Status.Phase == corev1.PodFailed && trainutil.IsRetryableExitCode(exitCode) {
					failedPodsCount.Inc()
					logger.Infof("Need to restart the pod: %v.%v", pod.Namespace, pod.Name)
					if err = r.DeletePod(ctx, pod.Namespace, pod.Name); err != nil {
						return err
					}
				}
			}

			core.UpdateJobReplicaStatuses(jobStatus, rtype, pod)
		}
	}
	return nil

}

func (r *KubeflowReconciler) CreateNewPod(job client.Object, rt commonv1.ReplicaType, index string,
	spec *commonv1.ReplicaSpec, masterRole bool, replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {

	logger := commonutil.LoggerForReplica(job, rt)

	podLabels := r.GenLabels(job.GetName())
	podLabels[commonv1.ReplicaTypeLabel] = string(rt)
	podLabels[commonv1.ReplicaIndexLabel] = index
	if masterRole {
		podLabels[commonv1.JobRoleLabel] = "master"
	}

	podTemplate := spec.Template.DeepCopy()

	podTemplate.Name = r.GenPodName(job.GetName(), rt, index)
	if podTemplate.Labels == nil {
		podTemplate.Labels = make(map[string]string)
	}

	for key, value := range podLabels {
		podTemplate.Labels[key] = value
	}

	if podTemplate.Spec.RestartPolicy != corev1.RestartPolicy("") {
		errMsg := "Restart policy in pod template will be overwritten by restart policy in replica spec"
		logger.Warning(errMsg)
		r.recorder.Event(job, corev1.EventTypeWarning, "SettedPodTemplateRestartPolicy", errMsg)
	}
	if spec.RestartPolicy == commonv1.RestartPolicyExitCode {
		podTemplate.Spec.RestartPolicy = corev1.RestartPolicyNever
	} else {
		podTemplate.Spec.RestartPolicy = corev1.RestartPolicy(spec.RestartPolicy)
	}

	if r.GangSchedulingEnabled() {
		r.DecoratePodForGangScheduling(rt, podTemplate, job)
	}

	r.DecoratePod(rt, podTemplate, job)

	pod := &corev1.Pod{
		ObjectMeta: podTemplate.ObjectMeta,
		Spec:       podTemplate.Spec,
	}

	err := controllerutil.SetControllerReference(job, pod, r.GetScheme())
	if err != nil {
		return err
	}

	err = r.Create(context.Background(), pod)
	if err != nil && errors.IsTimeout(err) {
		return nil
	} else if err != nil {
		return err
	}
	createdPodsCount.Inc()
	return nil
}

func (r *KubeflowReconciler) DeletePod(ctx context.Context, ns string, name string) error {
	pod := &corev1.Pod{}
	pod.Name = name
	pod.Namespace = ns
	err := r.Delete(ctx, pod)
	if err == nil {
		deletedPodsCount.Inc()
	}
	return err
}

func (r *KubeflowReconciler) DecoratePod(rtype commonv1.ReplicaType, podTemplate *corev1.PodTemplateSpec, job client.Object) {
	// Default implementation applies nothing to podTemplate
	return
}
