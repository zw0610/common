package common

import (
	"context"
	"strings"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cr *CommonReconciler) ReconcileServices(ctx context.Context, job metav1.Object, services []*v1.Service, rtype commonv1.ReplicaType,
	spec *commonv1.ReplicaSpec) error {

	rt := strings.ToLower(string(rtype))
	exists, err := FilterServicesForReplicaType(services, rt)
	if err != nil {
		return err
	}

	expects, err := cr.GetExpectedServicesForReplicaType(job, rtype)
	if err != nil {
		return err
	}

	var genericExists, genericExpects []metav1.Object
	for _, item := range exists {
		genericExists = append(genericExists, metav1.Object(item))
	}
	for _, item := range expects {
		genericExpects = append(genericExpects, metav1.Object(item))
	}

	toCreate, toUpdate, toDelete := SeparateByOverlapping(genericExpects, genericExists)

	for _, item := range toCreate {
		err = cr.Create(ctx, item.(*corev1.Service))
		if err != nil {
			return err
		}
	}

	for _, item := range toDelete {
		err = cr.Delete(ctx, item.(*corev1.Service))
		if err != nil {
			return err
		}
	}

	for _, item := range toUpdate {

	}

	return nil
}

func (cr *CommonReconciler) GetServicesForJob(ctx context.Context, job metav1.Object) ([]*corev1.Service, error) {
	var result corev1.ServiceList
	opts := &client.ListOptions{
		LabelSelector: labels.SelectorFromSet(GetDefaultLabels(job.GetName())),
		Namespace:     job.GetNamespace(),
	}
	err := cr.List(ctx, &result, opts)
	if err != nil {
		return nil, err
	}

	var slist []*corev1.Service = nil
	for _, s := range result.Items {
		slist = append(slist, &s)
	}
	return slist, nil
}

func FilterServicesForReplicaType(services []*corev1.Service, replicaType string) ([]*corev1.Service, error) {
	var result []*v1.Service

	replicaSelector := &metav1.LabelSelector{
		MatchLabels: make(map[string]string),
	}

	replicaSelector.MatchLabels[commonv1.ReplicaTypeLabel] = replicaType
	selector, err := metav1.LabelSelectorAsSelector(replicaSelector)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		if selector.Matches(labels.Set(service.Labels)) {
			result = append(result, service)
		}
	}
	return result, nil
}
