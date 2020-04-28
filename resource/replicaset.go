package resource

import (
	"k8s.io/api/apps/v1"
	rbacv1beta1 "k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
)

type ReplicaSetResource struct {
	Params   *handle.Resources
	PostData *rbacv1beta1.Role
}

func (r *ReplicaSetResource) Get() (*v1.ReplicaSet, error) {
	return r.Params.ClientSet.AppsV1().ReplicaSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *ReplicaSetResource) Delete() (err error) {
	if err = r.Params.ClientSet.AppsV1().ReplicaSets(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ReplicaSet,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ReplicaSetResource) List() (*v1.ReplicaSetList, error) {
	replicaSet := &v1.ReplicaSetList{}
	if replicaSetList, err := r.Params.ClientSet.AppsV1().ReplicaSets(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		if r.Params.Uid != "" {
			for _, replica := range replicaSetList.Items {
				owner := replica.ObjectMeta.OwnerReferences
				for _, controller := range owner {
					if string(controller.UID) == r.Params.Uid {
						replicaSet.Items = append(replicaSet.Items, replica)
					}
				}
			}
			return replicaSet, nil
		} else {
			return replicaSetList, nil
		}
	} else {
		return nil, err
	}
}
