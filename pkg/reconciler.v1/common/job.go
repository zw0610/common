package common

import (
	"context"

	"github.com/kubeflow/common/pkg/util"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CommonReconciler struct {
	CommonReconcilerInterface
	client.Client
	Config *JobReconcilerConfiguration
}

func MakeCommonReconciler(config *JobReconcilerConfiguration) CommonReconciler {
	if config == nil {
		config = MakeDefaultJobReconcilerConfiguration()
	}

	return CommonReconciler{
		Config: config,
	}
}

func (cr *CommonReconciler) IsGangSchedulingEnabled() bool {
	return cr.Config.EnableGangScheduling
}

func (cr *CommonReconciler) GetLoggerForKey(key string) *logrus.Entry {
	return util.LoggerForKey(key)
}

func (cr *CommonReconciler) UpdateJobStatus(ctx context.Context, job client.Object) error {
	return cr.Status().Update(ctx, job)
}

func (cr *CommonReconciler) CleaupJob(ctx context.Context, job client.Object) error {
	return cr.Delete(ctx, job)
}
