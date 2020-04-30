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

type NodeResource struct {
	Params   *handle.Resources
	PostData *v1.Node
}

func (r *NodeResource) Get() (*v1.Node, error) {
	return r.Params.ClientSet.CoreV1().Nodes().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *NodeResource) List() (*v1.NodeList, error) {
	return r.Params.ClientSet.CoreV1().Nodes().List(metav1.ListOptions{})
}

func (r *NodeResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().Nodes().Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Node,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *NodeResource) Patch() (res *v1.Node, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().Nodes().Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("Node patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Node,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *NodeResource) Update() (res *v1.Node, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Nodes().Update(r.PostData); err != nil {
		log.Errorf("Node update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Node,
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

func (r *NodeResource) Create() (res *v1.Node, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Nodes().Create(r.PostData); err != nil {
		log.Errorf("Node create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Node,
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

func (r *NodeResource) ListPodByNode() (*v1.PodList, error) {
	podItems := &v1.PodList{}
	pod := PodResource{Params: r.Params}
	if podList, err := pod.List(); err != nil {
		return nil, err
	} else {
		for _, p := range podList.Items {
			if p.Spec.NodeName == r.Params.Name {
				podItems.Items = append(podItems.Items, p)
			}
		}
	}
	return podItems, nil
}

func (r *NodeResource) NodeMetric() (interface{}, error) {
	metric := make(map[string]map[string]int64)
	var (
		podUsedCount  int64
		podTotalCount int64
		//memoryUsedCount int64
		memoryUnused     int64
		memoryTotalCount int64
		//cpuUsedCount int64
		cpuTotalCount int64
		//storageUsedCount int64
		storageTotalCount int64
	)
	if podList, err := r.ListPodByNode(); err != nil {
		return nil, err
	} else {
		// 只要运行中的Pod
		for _, p := range podList.Items {
			if p.Status.Phase == "Running" {
				podUsedCount++
			}
		}
	}
	if node, err := r.Get(); err != nil {
		return nil, err
	} else {
		podTotalCount, _ = node.Status.Allocatable.Pods().AsInt64()
		memoryTotalCount, _ = node.Status.Capacity.Memory().AsInt64()
		memoryUnused, _ = node.Status.Allocatable.Memory().AsInt64()
		cpuTotalCount, _ = node.Status.Allocatable.Cpu().AsInt64()
		storageTotalCount, _ = node.Status.Allocatable.StorageEphemeral().AsInt64()
	}
	metric["pod"] = map[string]int64{
		"total":  podTotalCount,
		"used":   podUsedCount,
		"unused": podTotalCount - podUsedCount,
	}
	metric["memory"] = map[string]int64{
		"total":  memoryTotalCount,
		"used":   memoryTotalCount - memoryUnused,
		"unused": memoryUnused,
	}
	metric["cpu"] = map[string]int64{
		"total":  cpuTotalCount,
		"used":   podUsedCount,
		"unused": podTotalCount - podUsedCount,
	}
	metric["storage"] = map[string]int64{
		"total":  storageTotalCount,
		"used":   podUsedCount,
		"unused": storageTotalCount,
	}
	return metric, nil
}

func (r *NodeResource) GenerateCreateData(c *gin.Context) (err error) {
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
