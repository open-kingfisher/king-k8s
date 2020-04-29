package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	hpav2beta2 "k8s.io/api/autoscaling/v2beta2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"strings"
)

type HPAResource struct {
	Params   *handle.Resources
	PostData *hpav2beta2.HorizontalPodAutoscaler
}

func (r *HPAResource) Get() (*hpav2beta2.HorizontalPodAutoscaler, error) {
	return r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *HPAResource) List() (*hpav2beta2.HorizontalPodAutoscalerList, error) {
	hpa := &hpav2beta2.HorizontalPodAutoscalerList{}
	if hpaList, err := r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		if r.Params.Kind != "" && r.Params.KindName != "" {
			for _, v := range hpaList.Items {
				if strings.ToLower(v.Spec.ScaleTargetRef.Kind) == r.Params.Kind && v.Spec.ScaleTargetRef.Name == r.Params.KindName {
					hpa.Items = append(hpa.Items, v)
				}
			}
			return hpa, nil
		} else {
			return hpaList, nil
		}
	} else {
		return nil, err
	}
}

func (r *HPAResource) Delete() (err error) {
	if err = r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.HPA,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *HPAResource) Patch() (res *hpav2beta2.HorizontalPodAutoscaler, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("HPA patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.HPA,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *HPAResource) Update() (res *hpav2beta2.HorizontalPodAutoscaler, err error) {
	if res, err = r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).Update(r.PostData); err != nil {
		log.Errorf("HPA update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.HPA,
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

func (r *HPAResource) Create() (res *hpav2beta2.HorizontalPodAutoscaler, err error) {
	if res, err = r.Params.ClientSet.AutoscalingV2beta2().HorizontalPodAutoscalers(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("HPA create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.HPA,
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

func (r *HPAResource) GenerateCreateData(c *gin.Context) (err error) {
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
