package resource

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-k8s/util"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type IngressResource struct {
	Params   *handle.Resources
	PostData *v1beta1.Ingress
}

func (r *IngressResource) Get() (*v1beta1.Ingress, error) {
	i, err := r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
	if err == nil {
		i.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "Ingress", Version: "extensions/v1beta1"})
	}
	return i, err
}

func (r *IngressResource) List() (*v1beta1.IngressList, error) {
	return r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *IngressResource) Delete() (err error) {
	if err = r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Ingress,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *IngressResource) Patch() (res *v1beta1.Ingress, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("Ingress patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Ingress,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *IngressResource) Update() (res *v1beta1.Ingress, err error) {
	if res, err = r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).Update(r.PostData); err != nil {
		log.Errorf("Ingress update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Ingress,
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

func (r *IngressResource) Create() (res *v1beta1.Ingress, err error) {
	if res, err = r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("Ingress create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Ingress,
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

type chart struct {
	Name      string                 `json:"name"`
	Value     map[string]string      `json:"value,omitempty"`
	Children  []chart                `json:"children,omitempty"`
	LineStyle map[string]string      `json:"lineStyle,omitempty"`
	ItemStyle map[string]interface{} `json:"itemStyle,omitempty"`
	Rank      string                 `json:"rank,omitempty"`
	PodNumber int                    `json:"podNumber,omitempty"`
}

func (r *IngressResource) GetChart() (interface{}, error) {
	serviceCache := make(map[string]string)
	podCache := make(map[string]*v1.PodList)
	chartData := chart{}
	if ingress, err := r.Get(); err != nil {
		return nil, err
	} else {
		podNumber := 0
		chartData.Name = r.Params.Name
		chartData.Rank = "ingress"
		for _, v := range ingress.Spec.Rules {
			host := chart{}
			host.Name = v.Host
			host.Rank = "domain"
			for _, p := range v.HTTP.Paths {
				path := chart{}
				path.Name = p.Path
				path.Rank = "patch"
				backend := chart{}
				backend.Rank = "service"
				backend.Name = fmt.Sprintf("%s:%d", p.Backend.ServiceName, p.Backend.ServicePort.IntVal)
				if labelSelector, ok := serviceCache[p.Backend.ServiceName]; !ok {
					if s, err := r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Get(p.Backend.ServiceName, metav1.GetOptions{}); err != nil {
						log.Errorf("get service by ingress error:%s", err)
						return nil, err
					} else {
						labelSelector = util.GenerateLabelSelector(s.Spec.Selector)
						serviceCache[p.Backend.ServiceName] = labelSelector
					}
				}
				labelSelector := serviceCache[p.Backend.ServiceName]
				if _, ok := podCache[labelSelector]; !ok {
					if pods, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).List(metav1.ListOptions{LabelSelector: labelSelector}); err != nil {
						log.Errorf("get pod by service error:%s", err)
						return nil, err
					} else {
						podCache[labelSelector] = pods
					}
				}
				for _, pp := range podCache[labelSelector].Items {
					pod := chart{}
					pod.Name = pp.Name
					pod.Rank = "pod"
					pod.Value = map[string]string{"ip": pp.Status.PodIP, "node": pp.Spec.NodeName, "status": string(pp.Status.Phase)}
					// 计算Pod数量
					podNumber++
					if pp.Status.Phase != "Running" {
						//pod.LineStyle = map[string]string{
						//	"color": "#ed4014",
						//	"width": "3",
						//	"type": "dotted",
						//}
						pod.ItemStyle = map[string]interface{}{
							"color":       "#ed4014",
							"borderColor": "#ed4014",
							"borderWidth": "3",
						}
					}
					backend.Children = append(backend.Children, pod)
				}
				path.Children = append(path.Children, backend)
				host.Children = append(host.Children, path)
			}
			chartData.Children = append(chartData.Children, host)
		}
		chartData.PodNumber = podNumber
	}
	return &chartData, nil
}

func (r *IngressResource) GetIngressByDeployment() (string, error) {
	i, err := r.Params.ClientSet.ExtensionsV1beta1().Ingresses(r.Params.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	ingressName := ""
	for _, ingress := range i.Items {
		for _, rules := range ingress.Spec.Rules {
			if rules.HTTP != nil && len(rules.HTTP.Paths) > 0 {
				for _, host := range rules.HTTP.Paths {
					serviceName, _ := r.GetServiceNameByDeploymentName()
					if host.Backend.ServiceName == serviceName {
						ingressName = ingress.Name
						goto End
					}
				}
			}
		}
	}
End:
	return ingressName, err
}

func (r *IngressResource) GetService(serviceName string) (map[string]string, error) {
	s, err := r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).Get(serviceName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.Spec.Selector, nil
}

func (r *IngressResource) GetServiceNameByDeploymentName() (string, error) {
	d, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	s, err := r.Params.ClientSet.CoreV1().Services(r.Params.Namespace).List(metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	serviceName := ""
	for _, service := range s.Items {
		tmp := true
		if service.Spec.Selector != nil {
			for key, value := range service.Spec.Selector {
				if d.Spec.Selector.MatchLabels[key] != value {
					tmp = false
					break
				}
			}
		} else {
			tmp = false
			continue
		}
		if tmp {
			serviceName = service.Name
			break
		}
	}
	return serviceName, nil
}

func (r *IngressResource) GenerateCreateData(c *gin.Context) (err error) {
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
