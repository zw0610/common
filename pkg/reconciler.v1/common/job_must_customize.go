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

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *KubeflowJobReconciler) GetJob(ctx context.Context, req ctrl.Request) (client.Object, error) {
	panic("implement KubeflowReconciler.GetJob!")
}

func (r *KubeflowJobReconciler) ExtractReplicasSpec(job client.Object) (map[commonv1.ReplicaType]*commonv1.ReplicaSpec, error) {
	panic("implement KubeflowReconciler.ExtractReplicasSpec!")
}

func (r *KubeflowJobReconciler) ExtractRunPolicy(job client.Object) (*commonv1.RunPolicy, error) {
	panic("implement KubeflowReconciler.ExtractRunPolicy")
}

func (r *KubeflowJobReconciler) ExtractJobStatus(job client.Object) (*commonv1.JobStatus, error) {
	panic("implement KubeflowReconciler.ExtractJobStatus")
}

func (r *KubeflowJobReconciler) IsMasterRole(replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec, rtype commonv1.ReplicaType, index int) bool {
	panic("implement KubeflowReconciler.IsMasterRole")
}
