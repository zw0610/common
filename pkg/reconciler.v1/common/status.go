package common

import (
	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
	corev1 "k8s.io/api/core/v1"
)

func (cr *CommonReconciler) IsJobSucceeded(status *commonv1.JobStatus) bool {
	return hasCondition(status, commonv1.JobSucceeded)
}
func (cr *CommonReconciler) IsJobFailed(status *commonv1.JobStatus) bool {
	return hasCondition(status, commonv1.JobFailed)
}

func hasCondition(status *commonv1.JobStatus, condType commonv1.JobConditionType) bool {
	for _, condition := range status.Conditions {
		if condition.Type == condType && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
