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

type RoleBindingResource struct {
	Params   *handle.Resources
	PostData *v1beta1.RoleBinding
}

func (r *RoleBindingResource) Get() (*v1beta1.RoleBinding, error) {
	return r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *RoleBindingResource) List() (*v1beta1.RoleBindingList, error) {
	return r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *RoleBindingResource) Delete() (err error) {
	if err = r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.RoleBinding,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *RoleBindingResource) Patch() (res *v1beta1.RoleBinding, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("RoleBinding patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.RoleBinding,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *RoleBindingResource) Update() (res *v1beta1.RoleBinding, err error) {
	if res, err = r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).Update(r.PostData); err != nil {
		log.Errorf("RoleBinding update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.RoleBinding,
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

func (r *RoleBindingResource) Create() (res *v1beta1.RoleBinding, err error) {
	if res, err = r.Params.ClientSet.RbacV1beta1().RoleBindings(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("RoleBinding create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.RoleBinding,
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

func (r *RoleBindingResource) GenerateCreateData(c *gin.Context) (err error) {
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
