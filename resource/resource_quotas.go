package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
)

type ResourceQuotasResource struct {
	Params   *handle.Resources
	PostData *v1.ResourceQuota
}

func (r *ResourceQuotasResource) Get() (*v1.ResourceQuota, error) {
	return r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *ResourceQuotasResource) List() (*v1.ResourceQuotaList, error) {
	return r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *ResourceQuotasResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		msg := fmt.Sprintf("resourcequotas \"%s\" not found", r.Params.Name)
		if err.Error() != msg {
			return
		}
	}
	auditLog := handle.AuditLog{
		Kind:       common.ResourceQuota,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ResourceQuotasResource) Patch() (res *v1.ResourceQuota, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("ResourceQuota patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ResourceQuota,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ResourceQuotasResource) Update() (res *v1.ResourceQuota, err error) {
	r.Params.Name = r.PostData.Name
	// quota 不存在创建
	if _, err = r.Get(); err != nil {
		if _, err = r.Create(); err != nil {
			return
		}
	} else {
		if res, err = r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).Update(r.PostData); err != nil {
			log.Errorf("LimitRange update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
			return
		}
	}
	auditLog := handle.AuditLog{
		Kind:       common.ResourceQuota,
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

func (r *ResourceQuotasResource) Create() (res *v1.ResourceQuota, err error) {
	if res, err = r.Params.ClientSet.CoreV1().ResourceQuotas(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("ResourceQuota create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ResourceQuota,
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

func (r *ResourceQuotasResource) GenerateCreateData(c *gin.Context) (err error) {
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
