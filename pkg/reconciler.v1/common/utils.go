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
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (r *KubeflowReconciler) GenLabels(jobName string) map[string]string {
	labelGroupName := commonv1.GroupNameLabel
	labelJobName := commonv1.JobNameLabel
	groupName := r.GetGroupNameLabelValue()
	return map[string]string{
		labelGroupName: groupName,
		labelJobName:   strings.Replace(jobName, "/", "-", -1),
	}
}

func (r *KubeflowReconciler) GetRecorder() record.EventRecorder {
	return r.recorder
}

func (r *KubeflowReconciler) GetLogger(job client.Object) logr.Logger {
	return r.Log.WithValues(
		job.GetObjectKind().GroupVersionKind().Kind,
		types.NamespacedName{Name: job.GetName(), Namespace: job.GetNamespace()}.String())
}

func (r *KubeflowReconciler) GetScheme() *runtime.Scheme {
	return r.Scheme
}
