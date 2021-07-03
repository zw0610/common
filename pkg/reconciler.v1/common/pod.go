package common

import (
	"context"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cr *CommonReconciler) ReconcilePods(ctx context.Context, job metav1.Object, pods []*v1.Pod, rtype commonv1.ReplicaType,
	replicas map[commonv1.ReplicaType]*commonv1.ReplicaSpec) error {
}

func (cr *CommonReconciler) GetPodsForJob(ctx context.Context, job metav1.Object) ([]*corev1.Pod, error) {
	var result corev1.PodList
	opts := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(GetDefaultLabels(job.GetName())),
		Namespace:     job.GetNamespace(),
	}
	err := cr.List(ctx, &result, opts)
	if err != nil {
		return nil, err
	}

	var plist []*corev1.Pod = nil
	for _, p := range result.Items {
		plist = append(plist, &p)
	}
	return plist, nil
}
