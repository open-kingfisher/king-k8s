package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	"k8s.io/api/rbac/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterRoleBindingResource struct {
	Params   *handle.Resources
	PostData *v1beta1.ClusterRoleBinding
}

func (r *ClusterRoleBindingResource) Get() (*v1beta1.ClusterRoleBinding, error) {
	return r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *ClusterRoleBindingResource) List() (*v1beta1.ClusterRoleBindingList, error) {
	return r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().List(metav1.ListOptions{})
}

func (r *ClusterRoleBindingResource) Delete() (err error) {
	if err = r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ClusterRoleBinding,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterRoleBindingResource) Patch() (res *v1beta1.ClusterRoleBinding, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("ClusterRoleBinding patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ClusterRoleBinding,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterRoleBindingResource) Update() (res *v1beta1.ClusterRoleBinding, err error) {
	if res, err = r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Update(r.PostData); err != nil {
		log.Errorf("ClusterRoleBinding update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ClusterRoleBinding,
		ActionType: common.Update,
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterRoleBindingResource) Create() (res *v1beta1.ClusterRoleBinding, err error) {
	if res, err = r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Create(r.PostData); err != nil {
		log.Errorf("ClusterRoleBinding create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ClusterRoleBinding,
		ActionType: common.Create,
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterRoleBindingResource) GenerateCreateData(c *gin.Context) (err error) {
	switch r.Params.DataType {
	case "yaml":
		var j []byte
		create := common.PostType{}
		if err = c.BindJSON(&create); err != nil {
			return
		}
		if j, _, err = kit.YamlToJson(create.Context); err != nil {
			return
		}
		if err = json.Unmarshal(j, &r.PostData); err != nil {
			return
		}
	case "json":
		if err = c.BindJSON(&r.PostData); err != nil {
			return
		}
	default:
		return errors.New(common.ContentTypeError)
	}
	return nil
}

// kingfisher service account 创建描述
func getKingfisherClusterRoleBindingSpec() *v1beta1.ClusterRoleBinding {
	clusterRoleBinding := &v1beta1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: common.ClusterRoleBindingName,
		},
		RoleRef: v1beta1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     "cluster-admin", // kingfisher 账户与 cluster-admin 进行绑定，cluster-admin 有最大权限
		},
		Subjects: []v1beta1.Subject{
			{Kind: "ServiceAccount",
				Name:      common.ServiceAccountName, //　已经创建的kingfisher Service Account
				Namespace: common.KubectlNamespace,
			},
		},
	}
	return clusterRoleBinding
}
