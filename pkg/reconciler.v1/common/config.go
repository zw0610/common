package common

import "time"

const (
	DefaultRequeueTime          = 1 * time.Minute
	DefaultGangSchedulingChoice = false
	DefaultGCWithOwnerRef       = true
)

type JobReconcilerConfiguration struct {
	// Requeue After time.Duration
	RequeueTime time.Duration

	// Enable gang scheduling by volcano
	EnableGangScheduling bool

	// Enable setting owner reference
	EnableGarbageCollectionWithOwnerReference bool
}

func MakeDefaultJobReconcilerConfiguration() *JobReconcilerConfiguration {
	return &JobReconcilerConfiguration{
		RequeueTime:          DefaultRequeueTime,
		EnableGangScheduling: DefaultGangSchedulingChoice,
		EnableGarbageCollectionWithOwnerReference: DefaultGCWithOwnerRef,
	}
}
