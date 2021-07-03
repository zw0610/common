package common

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (cr *CommonReconciler) ReconcileGangSchedulingResources(ctx context.Context, job metav1.Object) error {

}

func (cr *CommonReconciler) DeleteGangSchedulingResources(ctx context.Context, job metav1.Object) error {

}
