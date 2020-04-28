package resource

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	"time"
)

type ClusterPluginResource struct {
	Params   *handle.Resources
	PostData *common.ClusterPluginDB
	Plugin   string
}

func (r *ClusterPluginResource) Status() (interface{}, error) {
	pods, err := r.Params.ClientSet.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err == nil {
		for _, p := range pods.Items {
			if p.Name == common.KubectlPodName {
				return map[string]int{"Status": 1}, nil
			}
		}
	}
	return map[string]int{"Status": 0}, err
}

func (r *ClusterPluginResource) List() ([]*common.ClusterPluginDB, error) {
	plugin := make([]*common.ClusterPluginDB, 0)
	if err := db.List(common.DataField, common.ClusterPluginTable, &plugin, ""); err != nil {
		return nil, err
	}
	return plugin, nil
}

func (r *ClusterPluginResource) Delete() (err error) {
	if err = db.Delete(common.ClusterPluginTable, r.Params.Name); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ClusterPlugin,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterPluginResource) Create(c *gin.Context) (err error) {
	plugin := common.ClusterPluginDB{}
	if err = c.BindJSON(&plugin); err != nil {
		return err
	}
	r.PostData = &plugin
	// 对提交的数据进行校验
	if err = c.ShouldBindWith(r.PostData, binding.Query); err != nil {
		return err
	}
	pluginList := make([]*common.ClusterPluginDB, 0)
	if err = db.List(common.DataField, common.ClusterPluginTable, &pluginList, "WHERE data-> '$.plugin'=?", r.PostData.Plugin); err == nil {
		if len(pluginList) > 0 {
			return errors.New("the plugin name already install")
		}
	} else {
		return
	}
	r.PostData.Timestamp = time.Now().Unix()
	if err = db.Insert(common.ClusterPluginTable, r.PostData); err != nil {
		log.Errorf("Cluster Plugin add error:%s; Json:%+v;", err, r.PostData)
		return err
	}
	auditLog := handle.AuditLog{
		Kind:       common.Plugin,
		ActionType: common.Create,
		Resources:  r.Params,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}
