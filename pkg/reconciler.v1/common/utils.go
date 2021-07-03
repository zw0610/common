package common

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commonv1 "github.com/kubeflow/common/pkg/apis/common/v1"
)

const DefaultGroupName = "kubeflow.org"

func GetDefaultLabels(jobName string) map[string]string {
	labelGroupNameKey := commonv1.GroupNameLabel
	labelJobNameKey := commonv1.JobNameLabel
	groupName := DefaultGroupName
	return map[string]string{
		labelGroupNameKey: groupName,
		labelJobNameKey:   strings.Replace(jobName, "/", "-", -1),
	}
}

type generic struct {
	SourceSlice string
	Idx         int
	Val         int
}

func SeparateByOverlapping(expects []metav1.Object, exists []metav1.Object) (toCreate, toUpdate, toDelete []interface{}) {

	indexMap := map[string]generic{}

	for idx, item := range expects {
		key := item.GetNamespace() + "/" + item.GetName()
		indexMap[key] = generic{
			SourceSlice: "expects",
			Idx:         idx,
			Val:         1,
		}
	}

	for idx, item := range exists {
		key := item.GetNamespace() + "/" + item.GetName()
		g, ok := indexMap[key]
		if !ok {
			indexMap[key] = generic{
				SourceSlice: "exists",
				Idx:         idx,
				Val:         -1,
			}
		} else {
			g.Val = 0
			indexMap[key] = g
		}
	}

	for _, item := range indexMap {
		source := expects
		if item.SourceSlice == "exists" {
			source = exists
		}

		switch item.Val {
		case 1:
			toCreate = append(toCreate, source[item.Idx])
		case -1:
			toDelete = append(toDelete, source[item.Idx])
		case 0:
			toUpdate = append(toUpdate, source[item.Idx])
		}
	}

	return
}
