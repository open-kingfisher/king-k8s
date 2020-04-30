package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-k8s/util"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/kit"
	"k8s.io/api/apps/v1"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"
)

type ControllerResource struct {
	Params          *handle.Resources
	DeploymentData  *v1.Deployment
	DaemonSetData   *v1.DaemonSet
	StatefulSetData *v1.StatefulSet
	TemplateData    *common.TemplateDB
}

func (r *ControllerResource) Get() (interface{}, error) {
	switch r.Params.Controller {
	case "deployment":
		d, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
		if err == nil {
			d.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "Deployment", Version: "apps/v1"})
		}
		return d, err
	case "daemonset":
		d, err := r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
		if err == nil {
			d.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "DaemonSet", Version: "apps/v1"})
		}
		return d, err
	case "statefulset":
		d, err := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
		if err == nil {
			d.GetObjectKind().SetGroupVersionKind(schema.GroupVersionKind{Kind: "StatefulSet", Version: "apps/v1"})
		}
		return d, err
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) List() (interface{}, error) {
	switch r.Params.Controller {
	case "deployment":
		if res, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).List(metav1.ListOptions{}); err != nil {
			return nil, err
		} else {
			return res.Items, nil
		}
	case "daemonset":
		if res, err := r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).List(metav1.ListOptions{}); err != nil {
			return nil, err
		} else {
			return res.Items, nil
		}
	case "statefulset":
		if res, err := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).List(metav1.ListOptions{}); err != nil {
			return nil, err
		} else {
			return res.Items, nil
		}
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) ListPodByController() (*corev1.PodList, error) {
	switch r.Params.Controller {
	case "deployment":
		if res, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return nil, err
		} else {
			labelSelector := util.GenerateLabelSelector(res.Spec.Selector.MatchLabels)
			if pods, err := util.GetPodBySelectorLabel(labelSelector, r.Params.Namespace, r.Params.ClientSet); err != nil {
				log.Errorf("get pod by deployment error:%s", err)
				return nil, err
			} else {
				return pods, nil
			}
		}
	case "daemonset":
		if res, err := r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return nil, err
		} else {
			labelSelector := util.GenerateLabelSelector(res.Spec.Selector.MatchLabels)
			if pods, err := util.GetPodBySelectorLabel(labelSelector, r.Params.Namespace, r.Params.ClientSet); err != nil {
				log.Errorf("get pod by daemonset error:%s", err)
				return nil, err
			} else {
				return pods, nil
			}
		}
	case "statefulset":
		if res, err := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return nil, err
		} else {
			labelSelector := util.GenerateLabelSelector(res.Spec.Selector.MatchLabels)
			if pods, err := util.GetPodBySelectorLabel(labelSelector, r.Params.Namespace, r.Params.ClientSet); err != nil {
				log.Errorf("get pod by statefulset error:%s", err)
				return nil, err
			} else {
				return pods, nil
			}
		}
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) GetChart() (interface{}, error) {
	nodeDic := map[string]string{}
	chartData := chart{}
	podNumber := 0
	chartData.Name = r.Params.Name
	chartData.Rank = "deployment"
	if podList, err := r.ListPodByController(); err != nil {
		return nil, err
	} else {
		for _, pp := range podList.Items {
			// 计算Pod数量
			podNumber++
			nodeDic[pp.Spec.NodeName] = ""
		}
		// 字典排序
		var keys []string
		for k, _ := range nodeDic {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, v := range keys {
			node := chart{}
			node.Name = v
			node.Rank = "node"
			for _, pp := range podList.Items {
				pod := chart{}
				pod.Name = pp.Name
				pod.Rank = "pod"
				pod.Value = map[string]string{"ip": pp.Status.PodIP, "node": pp.Spec.NodeName, "status": string(pp.Status.Phase)}
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
				if pp.Spec.NodeName == v {
					node.Children = append(node.Children, pod)
				}
			}
			chartData.Children = append(chartData.Children, node)
		}

	}
	chartData.PodNumber = podNumber
	return &chartData, nil
}

func (r *ControllerResource) Delete() (err error) {
	switch r.Params.Controller {
	case "deployment":
		if item, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return err
		} else {
			r.Params.Uid = string(item.UID)
		}
		if err = r.DelReplicaSetForController(); err != nil {
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Delete,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "daemonset":
		if err = r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.Delete,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":
		if err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Delete,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) Patch() (res interface{}, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	switch r.Params.Controller {
	case "deployment":
		if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("Deployment patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "daemonset":
		if res, err = r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("DaemonSet patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":
		if res, err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("StatefulSet patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) PatchSync() (res interface{}, err error) {
	var data []byte
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	//defer cancel()
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	switch r.Params.Controller {
	case "deployment":
		if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("Deployment patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return r.WatchPodIP()
	case "daemonset":
		if res, err = r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("DaemonSet patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":
		if res, err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
			log.Errorf("StatefulSet patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Patch,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) Update() (res interface{}, err error) {
	switch r.Params.Controller {
	case "deployment":
		if r.Params.PostType == "form" {
			d, _ := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.DeploymentData.Name, metav1.GetOptions{})
			// Containers
			for index, newContainer := range r.DeploymentData.Spec.Template.Spec.Containers {
				for _, oldContainer := range d.Spec.Template.Spec.Containers {
					if newContainer.Name == oldContainer.Name {
						r.DeploymentData.Spec.Template.Spec.Containers[index] = oldContainer // 先等于旧的在替换
						r.DeploymentData.Spec.Template.Spec.Containers[index].Name = newContainer.Name
						r.DeploymentData.Spec.Template.Spec.Containers[index].Image = newContainer.Image
						r.DeploymentData.Spec.Template.Spec.Containers[index].ImagePullPolicy = newContainer.ImagePullPolicy
						r.DeploymentData.Spec.Template.Spec.Containers[index].Ports = newContainer.Ports
						r.DeploymentData.Spec.Template.Spec.Containers[index].VolumeMounts = newContainer.VolumeMounts
						r.DeploymentData.Spec.Template.Spec.Containers[index].Resources = newContainer.Resources
						r.DeploymentData.Spec.Template.Spec.Containers[index].Command = newContainer.Command
						r.DeploymentData.Spec.Template.Spec.Containers[index].Args = newContainer.Args
						r.DeploymentData.Spec.Template.Spec.Containers[index].Env = newContainer.Env
						r.DeploymentData.Spec.Template.Spec.Containers[index].LivenessProbe = newContainer.LivenessProbe
						r.DeploymentData.Spec.Template.Spec.Containers[index].ReadinessProbe = newContainer.ReadinessProbe
						if newContainer.SecurityContext != nil {
							if r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext == nil {
								r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext = &corev1.SecurityContext{
									Privileged: func() *bool {
										t := true
										return &t
									}(),
								}
							}
							if r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext != nil {
								r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext.Privileged = newContainer.SecurityContext.Privileged
							}
						}
						if newContainer.SecurityContext == nil {
							if r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext != nil {
								r.DeploymentData.Spec.Template.Spec.Containers[index].SecurityContext.Privileged = func() *bool {
									t := false
									return &t
								}()
							}
						}
					}
				}
			}
			// InitContainers
			for index, newContainer := range r.DeploymentData.Spec.Template.Spec.InitContainers {
				for _, oldContainer := range d.Spec.Template.Spec.InitContainers {
					if newContainer.Name == oldContainer.Name {
						r.DeploymentData.Spec.Template.Spec.InitContainers[index] = oldContainer // 先等于旧的在替换
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Name = newContainer.Name
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Image = newContainer.Image
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].ImagePullPolicy = newContainer.ImagePullPolicy
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Ports = newContainer.Ports
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].VolumeMounts = newContainer.VolumeMounts
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Resources = newContainer.Resources
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Command = newContainer.Command
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Args = newContainer.Args
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].Env = newContainer.Env
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].LivenessProbe = newContainer.LivenessProbe
						r.DeploymentData.Spec.Template.Spec.InitContainers[index].ReadinessProbe = newContainer.ReadinessProbe
						if newContainer.SecurityContext != nil {
							if r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext == nil {
								r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext = &corev1.SecurityContext{
									Privileged: func() *bool {
										t := true
										return &t
									}(),
								}
							}
							if r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext != nil {
								r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext.Privileged = newContainer.SecurityContext.Privileged
							}
						}
						if newContainer.SecurityContext == nil {
							if r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext != nil {
								r.DeploymentData.Spec.Template.Spec.InitContainers[index].SecurityContext.Privileged = func() *bool {
									t := false
									return &t
								}()
							}
						}
					}
				}
			}
		}

		if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Update(r.DeploymentData); err != nil {
			log.Errorf("Deployment update error:%s; Json:%+v; Name:%s", err, r.DeploymentData, r.DeploymentData.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Update,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.DeploymentData.Name,
			PostData:   &r.DeploymentData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "daemonset":
		if res, err = r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Update(r.DaemonSetData); err != nil {
			log.Errorf("DaemonSet update error:%s; Json:%+v; Name:%s", err, r.DaemonSetData, r.DaemonSetData.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.Update,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.DaemonSetData.Name,
			PostData:   &r.DaemonSetData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":

		if r.Params.PostType == "form" {
			d, _ := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Get(r.StatefulSetData.Name, metav1.GetOptions{})
			// Containers
			for index, newContainer := range r.StatefulSetData.Spec.Template.Spec.Containers {
				for _, oldContainer := range d.Spec.Template.Spec.Containers {
					if newContainer.Name == oldContainer.Name {
						r.StatefulSetData.Spec.Template.Spec.Containers[index] = oldContainer // 先等于旧的在替换
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Name = newContainer.Name
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Image = newContainer.Image
						r.StatefulSetData.Spec.Template.Spec.Containers[index].ImagePullPolicy = newContainer.ImagePullPolicy
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Ports = newContainer.Ports
						r.StatefulSetData.Spec.Template.Spec.Containers[index].VolumeMounts = newContainer.VolumeMounts
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Resources = newContainer.Resources
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Command = newContainer.Command
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Args = newContainer.Args
						r.StatefulSetData.Spec.Template.Spec.Containers[index].Env = newContainer.Env
						r.StatefulSetData.Spec.Template.Spec.Containers[index].LivenessProbe = newContainer.LivenessProbe
						r.StatefulSetData.Spec.Template.Spec.Containers[index].ReadinessProbe = newContainer.ReadinessProbe
						if newContainer.SecurityContext != nil {
							if r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext == nil {
								r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext = &corev1.SecurityContext{
									Privileged: func() *bool {
										t := true
										return &t
									}(),
								}
							}
							if r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext != nil {
								r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext.Privileged = newContainer.SecurityContext.Privileged
							}
						}
						if newContainer.SecurityContext == nil {
							if r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext != nil {
								r.StatefulSetData.Spec.Template.Spec.Containers[index].SecurityContext.Privileged = func() *bool {
									t := false
									return &t
								}()
							}
						}
					}
				}
			}
			// InitContainers
			for index, newContainer := range r.StatefulSetData.Spec.Template.Spec.InitContainers {
				for _, oldContainer := range d.Spec.Template.Spec.InitContainers {
					if newContainer.Name == oldContainer.Name {
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index] = oldContainer // 先等于旧的在替换
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Name = newContainer.Name
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Image = newContainer.Image
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].ImagePullPolicy = newContainer.ImagePullPolicy
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Ports = newContainer.Ports
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].VolumeMounts = newContainer.VolumeMounts
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Resources = newContainer.Resources
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Command = newContainer.Command
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Args = newContainer.Args
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].Env = newContainer.Env
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].LivenessProbe = newContainer.LivenessProbe
						r.StatefulSetData.Spec.Template.Spec.InitContainers[index].ReadinessProbe = newContainer.ReadinessProbe
						if newContainer.SecurityContext != nil {
							if r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext == nil {
								r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext = &corev1.SecurityContext{
									Privileged: func() *bool {
										t := true
										return &t
									}(),
								}
							}
							if r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext != nil {
								r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext.Privileged = newContainer.SecurityContext.Privileged
							}
						}
						if newContainer.SecurityContext == nil {
							if r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext != nil {
								r.StatefulSetData.Spec.Template.Spec.InitContainers[index].SecurityContext.Privileged = func() *bool {
									t := false
									return &t
								}()
							}
						}
					}
				}
			}
		}

		if res, err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Update(r.StatefulSetData); err != nil {
			log.Errorf("StatefulSet update error:%s; Json:%+v; Name:%s", err, r.StatefulSetData, r.StatefulSetData.Name)
			return
		}

		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Update,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.StatefulSetData.Name,
			PostData:   &r.StatefulSetData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) Create() (res interface{}, err error) {
	switch r.Params.Controller {
	case "deployment":
		if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Create(r.DeploymentData); err != nil {
			log.Errorf("Deployment create error:%s; Json:%+v; Name:%s", err, r.DeploymentData, r.DeploymentData.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Create,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.DeploymentData.Name,
			PostData:   &r.DeploymentData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "daemonset":
		if res, err = r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Create(r.DaemonSetData); err != nil {
			log.Errorf("DaemonSet create error:%s; Json:%+v; Name:%s", err, r.DaemonSetData, r.DaemonSetData.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.Create,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.DaemonSetData.Name,
			PostData:   &r.DaemonSetData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":
		if res, err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Create(r.StatefulSetData); err != nil {
			log.Errorf("StatefulSet create error:%s; Json:%+v; Name:%s", err, r.StatefulSetData, r.StatefulSetData.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Create,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.StatefulSetData.Name,
			PostData:   &r.StatefulSetData,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) SaveAsTemplate() error {
	switch r.Params.Controller {
	case "deployment":
		if res, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return err
		} else {
			deployment := &v1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:        res.Name,
					Namespace:   res.Namespace,
					Labels:      res.Labels,
					Annotations: res.Annotations,
				},
				Spec: res.Spec,
			}
			r.TemplateData.Spec = deployment
			r.TemplateData.Kind = r.Params.Controller
			if err := handle.CreateTemplate(r.TemplateData); err != nil {
				log.Errorf("Template create error:%s; Json:%+v; Name:%s", err, r.TemplateData.Spec, r.TemplateData.Name)
				return err
			}
		}

		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.SaveTemplate,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.TemplateData.Name,
			PostData:   &r.DeploymentData,
		}
		if err := auditLog.InsertAuditLog(); err != nil {
			return err
		}
	case "daemonset":
		if res, err := r.Params.ClientSet.AppsV1().DaemonSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return err
		} else {
			daemonSet := &v1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        res.Name,
					Namespace:   res.Namespace,
					Labels:      res.Labels,
					Annotations: res.Annotations,
				},
				Spec: res.Spec,
			}
			r.TemplateData.Spec = daemonSet
			r.TemplateData.Kind = r.Params.Controller
			if err := handle.CreateTemplate(r.TemplateData); err != nil {
				log.Errorf("Template create error:%s; Json:%+v; Name:%s", err, r.TemplateData.Spec, r.TemplateData.Name)
				return err
			}
		}
		auditLog := handle.AuditLog{
			Kind:       common.DaemonSet,
			ActionType: common.SaveTemplate,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.TemplateData.Name,
			PostData:   &r.DaemonSetData,
		}
		if err := auditLog.InsertAuditLog(); err != nil {
			return err
		}
	case "statefulset":
		if res, err := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); err != nil {
			return err
		} else {
			statefulSet := &v1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:        res.Name,
					Namespace:   res.Namespace,
					Labels:      res.Labels,
					Annotations: res.Annotations,
				},
				Spec: res.Spec,
			}
			r.TemplateData.Spec = statefulSet
			r.TemplateData.Kind = r.Params.Controller
			if err := handle.CreateTemplate(r.TemplateData); err != nil {
				log.Errorf("Template create error:%s; Json:%+v; Name:%s", err, r.TemplateData.Spec, r.TemplateData.Name)
				return err
			}
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.SaveTemplate,
			PostType:   common.ActionType(r.Params.PostType),
			Resources:  r.Params,
			Name:       r.TemplateData.Name,
			PostData:   &r.StatefulSetData,
		}
		if err := auditLog.InsertAuditLog(); err != nil {
			return err
		}
	default:
		return errors.New("controller kind doesn't exist")
	}
	return nil
}

func (r *ControllerResource) Scale() (res interface{}, err error) {
	scaleInt, err := strconv.Atoi(r.Params.Scale)
	if err != nil {
		return
	}
	scale := autoscalingv1.Scale{
		Spec: autoscalingv1.ScaleSpec{
			Replicas: int32(scaleInt),
		},
	}
	switch r.Params.Controller {
	case "deployment":
		if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).UpdateScale(r.Params.Name, &scale); err != nil {
			log.Errorf("Deployment scale error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.Deployment,
			ActionType: common.Scale,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	case "statefulset":
		if res, err = r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).UpdateScale(r.Params.Name, &scale); err != nil {
			log.Errorf("StatefulSet scale error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			return
		}
		auditLog := handle.AuditLog{
			Kind:       common.StatefulSet,
			ActionType: common.Scale,
			Resources:  r.Params,
			Name:       r.Params.Name,
		}
		if err = auditLog.InsertAuditLog(); err != nil {
			return
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

func (r *ControllerResource) Watch() (res watch.Interface, err error) {
	var labelSelector string
	switch r.Params.Controller {
	case "deployment":
		if deployment, errs := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); errs != nil {
			log.Errorf("Deployment get error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			return res, errs
		} else {
			// 获取指定的Deployment
			labelSelector = util.GenerateLabelSelector(deployment.Labels)
			watcher := metav1.ListOptions{
				LabelSelector: labelSelector,
			}
			if res, err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Watch(watcher); err != nil {
				log.Errorf("Deployment watch error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
				return
			}
		}
		return
	case "statefulset":
		if statefulSet, errs := r.Params.ClientSet.AppsV1().StatefulSets(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{}); errs != nil {
			log.Errorf("StatefulSet get error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			return res, errs
		} else {
			// 获取指定的StatefulSet
			labelSelector = util.GenerateLabelSelector(statefulSet.Labels)
			watcher := metav1.ListOptions{
				LabelSelector: labelSelector,
			}
			if res, err = r.Params.ClientSet.AppsV1beta2().StatefulSets(r.Params.Namespace).Watch(watcher); err != nil {
				log.Errorf("StatefulSet watch error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
				return
			}
		}
		return
	default:
		return nil, errors.New("controller kind doesn't exist")
	}
}

// 使用分组滚动更新后，剩下的手动分组进行滚动更新
func (r *ControllerResource) PatchStepResume() (interface{}, error) {
	// 检查滚动更新是否完成，已经完成直接返回
	//if ok, err := r.checkRollingUpdateIsCompleted(); err != nil {
	//	return nil, err
	//} else {
	//	if ok {
	//		return nil, errors.New("rolling update is completed")
	//	}
	//}
	var updatedReplicas int32
	var replicas int32
	if deployment, err := r.assertDeployment(); err != nil {
		return nil, err
	} else {
		updatedReplicas = deployment.Status.UpdatedReplicas
		replicas = *deployment.Spec.Replicas
		if updatedReplicas == replicas {
			podIP := make([]string, 0)
			return map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "podIP": podIP}, nil
		}
	}
	ch := make(chan int)
	// 开始监听Deployment事件
	go func() {
		if err := r.Resume(); err != nil {
			log.Errorf("watch event error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
		}
		ch <- 1
	}()
	// paused为true，开始滚动更新
	if err := r.SetResume(); err != nil {
		return nil, err
	}
	<-ch
	// 同步等待已经更新的PodIP，超时2分钟
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	//defer cancel()
	return r.WatchPodIP()
}

// 使用分组滚动更新后，剩下的自动进行滚动更新
func (r *ControllerResource) PatchAllResume() (interface{}, error) {
	if err := r.SetResume(); err != nil {
		return nil, err
	}
	// 同步等待已经更新的PodIP，超时2分钟
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	//defer cancel()
	return r.WatchPodIP()
}

// 手动暂停滚动更新
func (r *ControllerResource) PatchPause() (interface{}, error) {
	if err := r.SetPause(); err != nil {
		return nil, err
	}
	var updatedReplicas int32
	var replicas int32
	if deployment, err := r.assertDeployment(); err != nil {
		return nil, err
	} else {
		updatedReplicas = deployment.Status.UpdatedReplicas
		replicas = deployment.Status.Replicas
	}
	podIP := make([]string, 0)
	return map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "podIP": podIP}, nil
}

// 分步上线的首次上线，修改最大不可达，修改镜像地址
func (r *ControllerResource) PatchImage() (interface{}, error) {
	// 校验step参数是否正确
	if err := r.verifyStep(); err != nil {
		return nil, err
	}
	// 检查滚动更新是否完成，没有完成直接返回
	//if ok, err := r.checkRollingUpdateIsCompleted(); err != nil {
	//	return err
	//} else {
	//	if !ok {
	//		return errors.New("rolling update is not complete")
	//	}
	//}
	// 获取更新镜像的patch json
	patchImage := r.Params.PatchData.Patches
	// 不论状态是否暂停，设置设置成恢复
	if err := r.SetResume(); err != nil {
		return nil, err
	}
	go func() {
		if err := r.Resume(); err != nil {
			// 一旦发生错误，直接停止滚动更新
			if err := r.SetPause(); err != nil {
				log.Errorf("patch pause true error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			}
			log.Errorf("patch image error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
		}
	}()
	r.Params.PatchData.Patches = patchImage
	// 更新镜像的时候修改maxUnavailable值为步长
	r.SetStrategy()
	if _, err := r.Patch(); err != nil {
		return nil, err
	}
	// 同步等待已经更新的PodIP，超时2分钟
	//ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	//defer cancel()
	return r.WatchPodIP()
}

func (r *ControllerResource) SetStrategy() {
	var value interface{}
	if strings.Contains(r.Params.Step, "%") {
		value = r.Params.Step
	} else {
		value, _ = strconv.Atoi(r.Params.Step)
	}
	patch := common.PatchData{
		Op:    "replace",
		Path:  "/spec/strategy/rollingUpdate/maxUnavailable",
		Value: value,
	}
	r.Params.PatchData.Patches = append(r.Params.PatchData.Patches, patch)
}

func (r *ControllerResource) SetPause() error {
	patchEmpty := append(common.PatchJson{}.Patches, common.PatchData{})
	r.Params.PatchData.Patches = patchEmpty
	patch := common.PatchData{
		Op:    "add",
		Path:  "/spec/paused",
		Value: true,
	}
	r.Params.PatchData.Patches[0] = patch
	if _, err := r.Patch(); err != nil {
		return err
	} else {
		return nil
	}
}

func (r *ControllerResource) SetResume() error {
	patchEmpty := append(common.PatchJson{}.Patches, common.PatchData{})
	r.Params.PatchData.Patches = patchEmpty
	patch := common.PatchData{
		Op:    "add",
		Path:  "/spec/paused",
		Value: false,
	}
	//r.Params.PatchData.Patches = append(r.Params.PatchData.Patches, patch)
	r.Params.PatchData.Patches[0] = patch
	if _, err := r.Patch(); err != nil {
		return err
	} else {
		return nil
	}
}

func (r *ControllerResource) Resume() error {
	var maxUnavailable int32
	var maxSurge int32
	var lastCompletedUpdatedReplicas int32
	var replicas int32
	// 获取event
	event, err := r.Watch()
	if err != nil {
		log.Errorf("Deployment watch error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
		return err
	}
	// 获取deployment
	if deployment, err := r.assertDeployment(); err == nil {
		// 获取定义的副本数
		replicas = *deployment.Spec.Replicas
		// 获取最大不可达数 Type为0的时候为整形，为1的时候为字符串类型
		if deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.Type == 0 {
			maxUnavailable = int32(deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue())
		} else if deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.Type == 1 {
			// 百分比的字符串转换成整形
			v, err := r.percentValueToIntValue(deployment.Spec.Strategy.RollingUpdate.MaxUnavailable.String())
			if err != nil {
				log.Errorf("max unavailable percent to int error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
				return err
			} else {
				// 除不尽的处理
				roundUp := false
				if roundUp {
					// 除不尽的情况直接返回不小于结果的最小整数，例： 35 * 10 / 100 maxUnavailable = 4
					maxUnavailable = int32(math.Ceil(float64(v) * (float64(replicas)) / 100))
				} else {
					// 除不尽的情况直接返回不大于结果的最大整数，例： 35 * 10 / 100 maxUnavailable = 3
					maxUnavailable = int32(math.Floor(float64(v) * (float64(replicas)) / 100))
				}
			}
		}
		// 获取最大激增数
		if deployment.Spec.Strategy.RollingUpdate.MaxSurge.Type == 0 {
			maxSurge = int32(deployment.Spec.Strategy.RollingUpdate.MaxSurge.IntValue())
		} else if deployment.Spec.Strategy.RollingUpdate.MaxSurge.Type == 1 {
			v, err := r.percentValueToIntValue(deployment.Spec.Strategy.RollingUpdate.MaxSurge.String())
			if err != nil {
				log.Errorf("max surge percent to int error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
				return err
			} else {
				// 除不尽的处理
				roundUp := false
				if roundUp {
					// 除不尽的情况直接返回不小于结果的最小整数，例： 35 * 10 / 100 maxUnavailable = 4
					maxSurge = int32(math.Ceil(float64(v) * (float64(replicas)) / 100))
				} else {
					// 除不尽的情况直接返回不大于结果的最大整数，例： 35 * 10 / 100 maxUnavailable = 3
					maxSurge = int32(math.Floor(float64(v) * (float64(replicas)) / 100))
				}
			}
		}
		// 获取状态里面的UpdatedReplicas数值，记为上一次已经完成的数量
		lastCompletedUpdatedReplicas = deployment.Status.UpdatedReplicas
	} else {
		return err
	}
	// 副本数和上一次已经完成的数量相等说明是首次进行暂停更新
	if lastCompletedUpdatedReplicas == replicas {
		lastCompletedUpdatedReplicas = 0
	}
	// 循环event的channel以便获取Department的更新
	for v := range event.ResultChan() {
		if d, ok := v.Object.(*v1.Deployment); ok {
			log.Infof("Deployment status: %+v\n", d.Status)
			updatedReplicas := d.Status.UpdatedReplicas
			log.Infof("lastCompletedUpdatedReplicas: %d updatedReplicas: %d maxUnavailable: %d maxSurge: %d", lastCompletedUpdatedReplicas, updatedReplicas, maxUnavailable, maxSurge)
			// 传递过来的步长数等于最大不可达数，数量达到这个范围进行Pause操作
			// lastCompletedUpdatedReplicas > updatedReplicas && lastCompletedUpdatedReplicas <= maxUnavailable+updatedReplicas+maxSurge
			if updatedReplicas == lastCompletedUpdatedReplicas+maxUnavailable+maxSurge+1 {
				// 进行暂停操作
				if err := r.SetPause(); err != nil {
					log.Errorf("Deployment patch pause error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
					return err
				} else {
					event.Stop()
					return nil
				}
			}
			// 说明已经更新完成，/spec/paused 设置成false
			// maxUnavailable+updatedReplicas+maxSurge >= replicas
			if d.Status.Replicas == replicas &&
				d.Status.UpdatedReplicas == replicas &&
				d.Status.ReadyReplicas == replicas &&
				d.Status.AvailableReplicas == replicas &&
				d.Status.UnavailableReplicas == 0 &&
				lastCompletedUpdatedReplicas != 0 {
				if err := r.SetResume(); err != nil {
					log.Errorf("Deployment patch resume error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
					return err
				}
				log.Info("resume: ", true)
				// 停止次channel
				event.Stop()
				return nil
			}
		} else {
			log.Errorf("Deployment assert error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
			return err
		}
	}
	return nil
}

func (r *ControllerResource) GenerateCreateData(c *gin.Context) (err error) {
	var kindType string
	var jsonByte []byte
	switch r.Params.DataType {
	case "yaml":
		create := common.PostType{}
		if err = c.BindJSON(&create); err != nil {
			return
		}
		// YAML转JSON并获取Kind类型
		if j, kind, err := kit.YamlToJson(create.Context); err != nil {
			return err
		} else {
			kindType = kind
			jsonByte = j
		}
	case "json":
		// 为了获取kind信息
		jsons := make([]byte, c.Request.ContentLength)
		if _, err = c.Request.Body.Read(jsons); err != nil {
			if err.Error() != "EOF" {
				return
			}
		}
		if j, kind, err := kit.JsonToJson(jsons); err != nil {
			return err
		} else {
			kindType = kind
			jsonByte = j
		}
	default:
		return errors.New(common.ContentTypeError)
	}
	switch kindType {
	case "Deployment":
		// JSON反序列化
		r.Params.Controller = "deployment"
		if err = json.Unmarshal(jsonByte, &r.DeploymentData); err != nil {
			return
		}
	case "DaemonSet":
		r.Params.Controller = "daemonset"
		if err = json.Unmarshal(jsonByte, &r.DaemonSetData); err != nil {
			return
		}
	case "StatefulSet":
		r.Params.Controller = "statefulset"
		if err = json.Unmarshal(jsonByte, &r.StatefulSetData); err != nil {
			return
		}
	default:
		return errors.New("controller kind doesn't exist")
	}
	return nil
}

func (r *ControllerResource) GetReplicaSetForController() ([]string, error) {
	replicaSets := make([]string, 0)
	resource := *r.Params
	replicaSet := ReplicaSetResource{
		Params: &resource,
	}
	if replicaSetList, err := replicaSet.List(); err == nil {
		// 通过uid获取对应部署的副本集
		for _, replica := range replicaSetList.Items {
			owner := replica.ObjectMeta.OwnerReferences
			for _, controller := range owner {
				if string(controller.UID) == r.Params.Uid {
					replicaSets = append(replicaSets, replica.Name)
				}
			}
		}
		return replicaSets, nil
	} else {
		return replicaSets, err
	}
}

func (r *ControllerResource) DelReplicaSetForController() error {
	// resource 获取到r.Params真正的值，在取地址付给Params， 否则replicaSet.Params.Name = v 将修改掉r.Params.Name的值，因为Params 要求传递是指针类型
	resource := *r.Params
	replicaSet := ReplicaSetResource{
		Params: &resource,
	}
	if replicaSets, err := r.GetReplicaSetForController(); err == nil {
		// 已经获取了副本集后才能删除，如果先删除Deployments,副本集的OwnerReferences就没有了，不能找到对应的副本集了
		if err = r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
			return err
		}
		log.Info("replicaSet:", replicaSets)
		for _, v := range replicaSets {
			replicaSet.Params.Name = v
			log.Info("replicaSetName:", replicaSet.Params.Name)
			if err := replicaSet.Delete(); err != nil {
				log.Errorf("Delete replicaSet %s error:%s", v, err)
			}
		}
		return nil
	} else {
		return err
	}
}

// 滚动更新相关方法
func (r *ControllerResource) assertDeployment() (*v1.Deployment, error) {
	deploymentInterface, err := r.Get()
	if err != nil {
		log.Errorf("Deployment get error:%s; Json:%+v; Name:%s", err, "", r.Params.Name)
		return &v1.Deployment{}, err
	}
	if deployment, ok := deploymentInterface.(*v1.Deployment); ok {
		return deployment, nil
	} else {
		return &v1.Deployment{}, errors.New("deployment assert error")
	}
}

func (r *ControllerResource) checkRollingUpdateIsCompleted() (bool, error) {
	var updatedReplicas int32
	var replicas int32
	if deployment, err := r.assertDeployment(); err == nil {
		// 获取状态里面的UpdatedReplicas数值
		updatedReplicas = deployment.Status.UpdatedReplicas
		// 获取定义的副本数
		replicas = *deployment.Spec.Replicas
	} else {
		return false, err
	}
	if updatedReplicas == replicas {
		return true, nil
	} else {
		return false, nil
	}
}

func (r *ControllerResource) percentValueToIntValue(percent string) (int, error) {
	s := strings.Replace(percent, "%", "", -1)
	if v, err := strconv.Atoi(s); err != nil {
		return 0, err
	} else {
		return v, nil
	}
}

func (r *ControllerResource) verifyStep() error {
	var replicas int32
	if deployment, err := r.assertDeployment(); err == nil {
		// 获取定义的副本数
		replicas = *deployment.Spec.Replicas
	} else {
		return err
	}
	if r.Params.Step == "" {
		return errors.New("step cannot be empty")
	}
	// step 为百分比
	if strings.Contains(r.Params.Step, "%") {
		if step, err := strconv.Atoi(strings.TrimRight(r.Params.Step, "%")); err != nil {
			return errors.New("invalid step parameter: " + err.Error())
		} else {
			if step > 100 {
				return errors.New("max unavailable greater then 100%")
			}
			if step <= 0 {
				return errors.New("max unavailable less then or equal to 0")
			}
		}
	} else { // step 为整数
		if step, err := strconv.Atoi(r.Params.Step); err != nil {
			return errors.New("invalid step parameter: " + err.Error())
		} else {
			if int32(step) > replicas {
				return errors.New("max unavailable greater then replicas")
			}
			if step <= 0 {
				return errors.New("max unavailable less then or equal to 0")
			}
		}
	}
	return nil
}

func (r *ControllerResource) WatchPodIP() (map[string]interface{}, error) {
	var updatedReplicas int32
	var replicas int32
	var unavailableReplicas int32
	if deployment, err := r.assertDeployment(); err != nil {
		return nil, err
	} else {
		// 说明镜像更新完成
		podIP := make([]string, 0)
		log.Debugf("deployment Status:", deployment.Status)
		updatedReplicas = deployment.Status.UpdatedReplicas
		unavailableReplicas = deployment.Status.UnavailableReplicas
		replicas = *deployment.Spec.Replicas
		uid := string(deployment.UID)
		if replicasUID, err := r.getLatestReplica(uid); err != nil {
			log.Errorf("getLatestReplica error: %s", err)
		} else {
			if podIP, err = r.getLatestReplicaPodIp(replicasUID); err != nil {
				log.Errorf("getLatestReplicaPodIp error: %s", err)
			}
		}
		if deployment.Status.UnavailableReplicas == 0 {
			log.Debugf("return replicas and groupCompleted:", map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "unavailableReplicas": unavailableReplicas, "podIP": podIP, "groupCompleted": 1})
			return map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "unavailableReplicas": unavailableReplicas, "podIP": podIP, "groupCompleted": 1}, nil
		}
		log.Debugf("return replicas:", map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "unavailableReplicas": unavailableReplicas, "podIP": podIP})
		return map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "unavailableReplicas": unavailableReplicas, "podIP": podIP}, nil
	}
}

func (r *ControllerResource) getPodIP(ctx context.Context) (map[string]interface{}, error) {
	var updatedReplicas int32
	var replicas int32
	var count int32
	for {
		time.Sleep(3 * time.Second)
		select {
		case <-ctx.Done():
			fmt.Println("the image was not updated within 2 minutes")
			return nil, errors.New("the image was not updated within 2 minutes")
		default:
			if deployment, err := r.assertDeployment(); err != nil {
				return nil, err
			} else {
				// 说明镜像更新完成
				count++
				fmt.Printf("count:%d UnavailableReplicas:%d\n", count, deployment.Status.UnavailableReplicas)
				podIP := make([]string, 0)
				if deployment.Status.UnavailableReplicas == 0 {
					updatedReplicas = deployment.Status.UpdatedReplicas
					replicas = *deployment.Spec.Replicas
					uid := string(deployment.UID)
					if replicasUID, err := r.getLatestReplica(uid); err != nil {
						log.Errorf("getLatestReplica error: %s", err)
					} else {
						if podIP, err = r.getLatestReplicaPodIp(replicasUID); err != nil {
							log.Errorf("getLatestReplicaPodIp error: %s", err)
						}
					}
					log.Debugf("return replicas:", map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "podIP": podIP})
					return map[string]interface{}{"replicas": replicas, "updatedReplicas": updatedReplicas, "podIP": podIP}, nil
				}
			}
		}
	}
}

// 根据Deployment返回最新Replica的uid，通过比较时间获取最新Replica
func (r *ControllerResource) getLatestReplica(uid string) (string, error) {
	if uid == "" {
		return "", errors.New("deployment uid is empty")
	}
	latestVersion := 0
	replicaUid := ""
	if replicaSetList, err := r.Params.ClientSet.AppsV1().ReplicaSets(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		for _, replica := range replicaSetList.Items {
			// 在回滚的时候，最新的副本不是正确的副本，副本数需要大于0
			if *replica.Spec.Replicas > 0 {
				owner := replica.ObjectMeta.OwnerReferences
				for _, controller := range owner {
					if string(controller.UID) == uid {
						creationTimestamp := int(replica.CreationTimestamp.Unix())
						if int(replica.CreationTimestamp.Unix()) > latestVersion {
							latestVersion = creationTimestamp
							replicaUid = string(replica.UID)
						}
					}
				}
			}
		}
	} else {
		return "", err
	}
	return replicaUid, nil
}

// 通过最新的Replica获取Pod的ip
func (r *ControllerResource) getLatestReplicaPodIp(uid string) ([]string, error) {
	ip := make([]string, 0)
	if uid == "" {
		return ip, errors.New("replica uid is empty")
	}
	if pods, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		for _, p := range pods.Items {
			owner := p.ObjectMeta.OwnerReferences
			for _, controller := range owner {
				if string(controller.UID) == uid {
					ip = append(ip, p.Status.PodIP)
				}
			}
		}
	} else {
		return ip, err
	}
	return ip, nil
}

func (r *ControllerResource) GetNamespaceIsExistLabel() (bool, error) {
	namespace, err := r.Params.ClientSet.CoreV1().Namespaces().Get(r.Params.Namespace, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	switch r.Params.Name {
	case "istio":
		if v, ok := namespace.Labels["istio-injection"]; ok && v == "enabled" {
			return true, nil
		}
	default:
		return false, nil
	}
	return false, nil
}
