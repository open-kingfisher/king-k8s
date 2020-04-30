package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type PVResource struct {
	Params   *handle.Resources
	PostData *v1.PersistentVolume
}

func (r *PVResource) Get() (*v1.PersistentVolume, error) {
	return r.Params.ClientSet.CoreV1().PersistentVolumes().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *PVResource) List() (*v1.PersistentVolumeList, error) {
	return r.Params.ClientSet.CoreV1().PersistentVolumes().List(metav1.ListOptions{})
}

func (r *PVResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().PersistentVolumes().Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.PV,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PVResource) Patch() (res *v1.PersistentVolume, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().PersistentVolumes().Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("PV patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.PV,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PVResource) Update() (res *v1.PersistentVolume, err error) {
	if res, err = r.Params.ClientSet.CoreV1().PersistentVolumes().Update(r.PostData); err != nil {
		log.Errorf("PV update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.PV,
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

func (r *PVResource) Create() (res *v1.PersistentVolume, err error) {
	if res, err = r.Params.ClientSet.CoreV1().PersistentVolumes().Create(r.PostData); err != nil {
		log.Errorf("PV create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.PV,
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

func (r *PVResource) GenerateCreateData(c *gin.Context) (err error) {
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
