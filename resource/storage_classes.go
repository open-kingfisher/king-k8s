package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	"k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StorageClassesResource struct {
	Params   *handle.Resources
	PostData *v1.StorageClass
}

func (r *StorageClassesResource) Get() (*v1.StorageClass, error) {
	return r.Params.ClientSet.StorageV1().StorageClasses().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *StorageClassesResource) List() (*v1.StorageClassList, error) {
	return r.Params.ClientSet.StorageV1().StorageClasses().List(metav1.ListOptions{})
}

func (r *StorageClassesResource) Delete() (err error) {
	if err = r.Params.ClientSet.StorageV1().StorageClasses().Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.StorageClasses,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *StorageClassesResource) Patch() (res *v1.StorageClass, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.StorageV1().StorageClasses().Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("StorageClasses patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.StorageClasses,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *StorageClassesResource) Update() (res *v1.StorageClass, err error) {
	if res, err = r.Params.ClientSet.StorageV1().StorageClasses().Update(r.PostData); err != nil {
		log.Errorf("StorageClasses update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.StorageClasses,
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

func (r *StorageClassesResource) Create() (res *v1.StorageClass, err error) {
	if res, err = r.Params.ClientSet.StorageV1().StorageClasses().Create(r.PostData); err != nil {
		log.Errorf("StorageClasses create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.StorageClasses,
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

func (r *StorageClassesResource) GenerateCreateData(c *gin.Context) (err error) {
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
