package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	"github.com/open-kingfisher/king-utils/kit"
	"time"
)

type CustomParam struct {
	Exist string
}

type NamespaceResource struct {
	Params       *handle.Resources
	PostData     *v1.Namespace
	CustomParams *CustomParam
}

type NamespaceResponse struct {
	common.Info
	Cluster     common.Info       `json:"cluster"`
	CreateTime  int64             `json:"createTime"`
	ModifyTime  int64             `json:"modifyTime"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func (r *NamespaceResource) Get() (*v1.Namespace, error) {
	return r.Params.ClientSet.CoreV1().Namespaces().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *NamespaceResource) ListAll() (*v1.NamespaceList, error) {
	return r.Params.ClientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
}

func (r *NamespaceResource) List() ([]*NamespaceResponse, error) {
	clusterDB := common.ClusterDB{}
	if err := db.GetById(common.ClusterTable, r.Params.Cluster, &clusterDB); err != nil {
		return nil, err
	}
	namespaceDB := make([]common.NamespaceDB, 0)
	if err := db.List(common.DataField, common.NamespaceTable, &namespaceDB, "where data -> '$.cluster'=?", r.Params.Cluster); err != nil {
		return nil, err
	}
	data := make([]*NamespaceResponse, 0)
	if namespaces, err := r.Params.ClientSet.CoreV1().Namespaces().List(metav1.ListOptions{}); err == nil {
		for _, v := range namespaceDB {
			for _, n := range namespaces.Items {
				if v.Name == n.Name {
					data = append(data, &NamespaceResponse{
						Info: common.Info{
							Id:   v.Id,
							Name: v.Name,
						},
						Cluster: common.Info{
							Id:   v.Cluster,
							Name: clusterDB.Name,
						},
						Labels:      n.Labels,
						Annotations: n.Annotations,
						CreateTime:  v.CreateTime,
						ModifyTime:  v.ModifyTime,
					})
				}
			}
		}
		return data, nil
	} else {
		return nil, err
	}
}

func (r *NamespaceResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().Namespaces().Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	// 获取Namespace ID
	namespace := common.NamespaceDB{}
	if err = db.Get(common.NamespaceTable, map[string]interface{}{"$.name": r.Params.Name, "$.cluster": r.Params.Cluster}, &namespace); err != nil {
		return
	}
	// 从Namespace表中删除对应的namespace
	if err = db.Delete(common.NamespaceTable, namespace.Id); err != nil {
		return
	}
	// 从产品表里删除对应的namespace
	if err = DeleteNamespaceForProduct(r.Params.Name); err != nil {
		return
	}
	// 从user表中删除对应的namespace
	user := make([]*common.User, 0)
	if err := db.List(common.DataField, common.UserTable, &user, ""); err == nil {
		for _, v := range user {
			v.Namespace = kit.DeleteItemForList(namespace.Id, v.Namespace)
			if err := db.Update(common.UserTable, v.Id, v); err != nil {
				log.Errorf("User table delete name :%s error:%s", v.Id, err)
			}
		}
	}
	auditLog := handle.AuditLog{
		Kind:       common.Namespace,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *NamespaceResource) Patch() (res *v1.Namespace, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().Namespaces().Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("Namespace patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Namespace,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *NamespaceResource) Update() (res *v1.Namespace, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Namespaces().Update(r.PostData); err != nil {
		log.Errorf("Namespace update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Namespace,
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

func (r *NamespaceResource) Create() (res *v1.Namespace, err error) {
	namespace := common.NamespaceDB{
		Id:         kit.UUID("n"),
		Name:       r.PostData.Name,
		Cluster:    r.Params.Cluster,
		Product:    r.Params.Product,
		CreateTime: time.Now().Unix(),
		ModifyTime: time.Now().Unix(),
	}
	namespaceList := make([]*common.NamespaceDB, 0)
	clause := "WHERE data-> '$.name'=? and data-> '$.cluster'=?"
	if err = db.List(common.DataField, common.NamespaceTable, &namespaceList, clause, r.PostData.Name, r.Params.Cluster); err == nil && len(namespaceList) > 0 {
		return nil, errors.New("the namespace name already exists")
	}
	switch r.CustomParams.Exist {
	case "0":
		if res, err = r.Params.ClientSet.CoreV1().Namespaces().Create(r.PostData); err != nil {
			log.Errorf("Namespace create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
			return
		}
		if err = db.Insert(common.NamespaceTable, namespace); err != nil {
			return
		}
	case "1":
		// 提交参数说已经存在，进行认证，若果不存在，返回错误
		r.Params.Name = r.PostData.Name
		if _, err := r.Get(); err != nil {
			return nil, err
		}
		if err = db.Insert(common.NamespaceTable, namespace); err != nil {
			return
		}
	default:
		return nil, errors.New("exist params not exist, please provide query parameters: exist 0 or 1")
	}
	auditLog := handle.AuditLog{
		Kind:       common.Namespace,
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

func (r *NamespaceResource) GenerateCreateData(c *gin.Context) (err error) {
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
