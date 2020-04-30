package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-k8s/util"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type ServiceResource struct {
	Params   *handle.Resources
	PostData *v1.Service
}

func (r *ServiceResource) Get() (*v1.Service, error) {
	s, err := r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
	if err == nil {
		s.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "Service", Version: "v1"})
	}
	return s, err
}

func (r *ServiceResource) List() (*v1.ServiceList, error) {
	return r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *ServiceResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Service,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ServiceResource) Patch() (res *v1.Service, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("Service patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Service,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ServiceResource) Update() (res *v1.Service, err error) {
	r.Params.Name = r.PostData.Name
	old, _ := r.Get()
	// 使用提交的结构体替换原来的，避免部分原有字段给删掉
	old.Labels = r.PostData.Labels
	old.Spec.Type = r.PostData.Spec.Type
	if r.PostData.Spec.Type == "ExternalName" {
		old.Spec.ExternalName = r.PostData.Spec.ExternalName
	}
	old.Spec.Ports = r.PostData.Spec.Ports
	old.Spec.Selector = r.PostData.Spec.Selector
	old.Spec.SessionAffinity = r.PostData.Spec.SessionAffinity
	if res, err = r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Update(old); err != nil {
		log.Errorf("Service update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Service,
		ActionType: common.Update,
		PostType:   common.ActionType(r.Params.PostType),
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ServiceResource) Create() (res *v1.Service, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("Service create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Service,
		ActionType: common.Create,
		PostType:   common.ActionType(r.Params.PostType),
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ServiceResource) ListPodByService() (*v1.PodList, error) {
	if service, err := r.Get(); err != nil {
		return nil, err
	} else {
		labelSelector := util.GenerateLabelSelector(service.Spec.Selector)
		if pods, err := util.GetPodBySelectorLabel(labelSelector, r.Params.Namespace, r.Params.ClientSet); err != nil {
			log.Errorf("get pod by service error:%s", err)
			return nil, err
		} else {
			return pods, nil
		}
	}
}

func (r *ServiceResource) GenerateCreateData(c *gin.Context) (err error) {
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
