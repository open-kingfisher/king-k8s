package resource

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/common/rabbitmq"
	"github.com/open-kingfisher/king-utils/config"
	"github.com/open-kingfisher/king-utils/db"
	"github.com/open-kingfisher/king-utils/kit"
	"golang.org/x/sync/errgroup"
	"time"
)

type ClusterResource struct {
	Params *handle.Resources
	common.ClusterDB
}

func (r *ClusterResource) Get() (*common.ClusterDB, error) {
	cluster := common.ClusterDB{}
	if err := db.GetById(common.Cluster, r.Params.Name, &cluster); err != nil {
		return nil, err
	}
	return &cluster, nil
}

func (r *ClusterResource) List() ([]*common.ClusterDB, error) {
	clusters := make([]*common.ClusterDB, 0)
	if err := db.List(common.DataField, common.Cluster, &clusters, ""); err != nil {
		return nil, err
	}
	var g errgroup.Group
	var clusterList []*common.ClusterDB
	for _, cluster := range clusters {
		cluster := cluster
		g.Go(func() error {
			version := "无法获取版本，请检查集群连通性或者配置文件"
			if clientSet, err := access.Access(cluster.Id); err != nil {
				log.Errorf("%s cluster access error:%s", cluster.Id, err)
			} else {
				// 获取kubernetes集群版本
				if v, err := clientSet.Discovery().ServerVersion(); err != nil {
					log.Errorf("%s cluster get server version error:%s", cluster.Id, err)
				} else {
					version = v.GitVersion
				}
			}
			cluster.Version = version
			clusterList = append(clusterList, cluster)
			return nil
		})
	}
	g.Wait()
	return clusterList, nil
}

func (r *ClusterResource) Delete() (err error) {
	// 从用户表里删除集群
	if err = r.DeleteClusterForUser(r.Params.Name); err != nil {
		return
	}
	// 从产品表中删除集群
	if err = DeleteClusterForProduct(r.Params.Name); err != nil {
		return
	}
	// 从产品表中删除集群对应的Namespace
	if err = PatchNamespaceForProduct(r.Params.Name); err != nil {
		return
	}
	// 从namespace表中删除对应的namespace
	if err = r.DeleteNamespaceForNamespace(r.Params.Name); err != nil {
		return
	}
	// 从集群表里删除集群
	if err = db.Delete(common.Cluster, r.Params.Name); err != nil {
		return
	}
	// 删除文件
	if err = kit.DeleteConfig(common.KubeConfigPath + r.Params.Name); err != nil {
		log.Errorf("remove kubeconfig error:%s", err)
	}
	auditLog := handle.AuditLog{
		Kind:       common.Cluster,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterResource) Create(c *gin.Context) (err error) {
	r.Id = kit.UUID("c")
	r.ClusterDB.Id = kit.UUID("c")
	r.ClusterDB.Name = c.PostForm("name")
	r.ClusterDB.Describe = c.PostForm("describe")
	r.ClusterDB.Token = c.PostForm("token")
	r.ClusterDB.CAHash = c.PostForm("ca_hash")
	r.ClusterDB.Product = c.PostFormArray("product")
	// 对提交的数据进行校验
	if err = c.ShouldBindWith(&r.ClusterDB, binding.Query); err != nil {
		return
	}
	clusterList := make([]*common.ClusterDB, 0)
	if err = db.List(common.DataField, common.Cluster, &clusterList, "WHERE data-> '$.name'=?", r.ClusterDB.Name); err == nil {
		if len(clusterList) > 0 {
			return errors.New("the cluster name already exists")
		}
	}
	if file, err := c.FormFile("kub_config"); err != nil {
		return err
	} else {
		f, _ := file.Open()
		// 根据上传文件大小初始化buf
		buf := make([]byte, file.Size)
		for {
			l, _ := f.Read(buf)
			if l == 0 {
				break
			}
		}
		r.ClusterDB.KubConfig = string(buf)
		f.Close()
		// 创建 kubeconfig 配置文件
		kubeconfig := common.KubeConfigPath + r.ClusterDB.Id
		if err := kit.CreateConfig(r.ClusterDB, kubeconfig); err != nil {
			return err
		}
	}
	r.ClusterDB.CreateTime = time.Now().Unix()
	r.ClusterDB.ModifyTime = time.Now().Unix()
	if err = db.Insert(common.ClusterTable, r.ClusterDB); err != nil {
		log.Errorf("Cluster create error:%s; Json:%+v; Name:%s", err, r.ClusterDB, r.ClusterDB.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:        common.Cluster,
		ActionType:  common.Create,
		Resources:   r.Params,
		ClusterData: &r.ClusterDB,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ClusterResource) Update(c *gin.Context) (err error) {
	r.ClusterDB.Id = r.Params.Cluster
	r.ClusterDB.Name = c.PostForm("name")
	r.ClusterDB.Describe = c.PostForm("describe")
	r.ClusterDB.Token = c.PostForm("token")
	r.ClusterDB.CAHash = c.PostForm("ca_hash")
	r.ClusterDB.Product = c.PostFormArray("product")
	// 对提交的数据进行校验
	if err = c.ShouldBindWith(&r.ClusterDB, binding.Query); err != nil {
		return
	}
	clusterList := make([]*common.ClusterDB, 0)
	if err = db.List(common.DataField, common.Cluster, &clusterList, "WHERE data-> '$.name'=?", r.ClusterDB.Name); err == nil {
		if len(clusterList) > 0 {
			for _, v := range clusterList {
				if v.Id != r.Params.Cluster {
					return errors.New("the cluster name already exists")
				}
			}
		}
	}
	clusterTmp := common.ClusterDB{}
	if err = db.GetById(common.Cluster, r.Params.Cluster, &clusterTmp); err != nil {
		return err
	}
	if file, err := c.FormFile("kub_config"); err != nil {
		// 修改的时候没有更新kubeconfg文件，查询数据库获取之前的kubeconfig
		if err.Error() == "http: no such file" {
			r.ClusterDB.KubConfig = clusterTmp.KubConfig
		} else {
			return err
		}
	} else {
		f, _ := file.Open()
		// 根据上传文件大小初始化buf
		buf := make([]byte, file.Size)
		for {
			l, _ := f.Read(buf)
			if l == 0 {
				break
			}
		}
		r.ClusterDB.KubConfig = string(buf)
		f.Close()
		// 更新kubeconfig配置文件
		kubeconfig := common.KubeConfigPath + r.ClusterDB.Id
		if err = kit.CreateConfig(r.ClusterDB, kubeconfig); err != nil {
			return err
		}
	}
	r.ClusterDB.CreateTime = clusterTmp.CreateTime
	r.ClusterDB.ModifyTime = time.Now().Unix()
	if err = db.Update(common.ClusterTable, r.ClusterDB.Id, r.ClusterDB); err != nil {
		log.Errorf("Cluster update error:%s; Json:%+v; Name:%s", err, r.ClusterDB, r.ClusterDB.Name)
		return
	}
	// 发布消息给其他微服务，进行kube-config的更新
	log.Infof("%s(%s) cluster kube-config update", r.ClusterDB.Name, r.ClusterDB.Id)
	clusterByte, _ := json.Marshal(r.ClusterDB)
	if err = rabbitmq.ProducerPublish(config.RabbitMQURL, common.UpdateKubeConfig, clusterByte); err != nil {
		log.Errorf("RabbitMQ producer error: %s", err)
	}
	auditLog := handle.AuditLog{
		Kind:        common.Cluster,
		ActionType:  common.Update,
		Resources:   r.Params,
		ClusterData: &r.ClusterDB,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

// 从产品表里删除对应的cluster
func DeleteClusterForProduct(clusterId string) error {
	product := make([]*common.ProductDB, 0)
	if err := db.List(common.DataField, common.ProductTable, &product, ""); err == nil {
		for _, v := range product {
			clusterList := make([]string, 0)
			for _, n := range v.Cluster {
				if n != clusterId {
					clusterList = append(clusterList, n)
				}
			}
			if err := db.Update(common.ProductTable, v.Id, common.ProductDB{Id: v.Id, Cluster: clusterList, Namespace: v.Namespace}); err != nil {
				log.Errorf("Update product:%s error:%s", v.Id, err)
			}
		}
		return nil
	} else {
		return err
	}
}

// 从产品表里删除对应的namespace
func DeleteNamespaceForProduct(namespace string) error {
	product := make([]*common.ProductDB, 0)
	if err := db.List(common.DataField, common.ProductTable, &product, ""); err == nil {
		for _, v := range product {
			namespaceList := make([]string, 0)
			for _, n := range v.Namespace {
				if n != namespace {
					namespaceList = append(namespaceList, n)
				}
			}
			if err := db.Update(common.ProductTable, v.Id, common.ProductDB{Id: v.Id, Cluster: v.Cluster, Namespace: namespaceList}); err != nil {
				log.Errorf("Update product:%s error:%s", v.Id, err)
			}
		}
		return nil
	} else {
		return err
	}
}

// 从产品表中删除集群对应的Namespace
func PatchNamespaceForProduct(clusterId string) error {
	namespace := make([]*common.NamespaceDB, 0)
	if err := db.List(common.DataField, common.Namespace, &namespace, "WHERE data-> '$.cluster'=?", clusterId); err == nil {
		for _, v := range namespace {
			if err := DeleteNamespaceForProduct(v.Id); err != nil {
				log.Errorf("Delete namespace:%s error:%s", v.Id, err)
			}
		}
		return nil
	} else {
		return err
	}
}

// 从namespace表中删除对应的namespace
func (r *ClusterResource) DeleteNamespaceForNamespace(clusterId string) error {
	namespace := make([]*common.NamespaceDB, 0)
	if err := db.List(common.DataField, common.Namespace, &namespace, "WHERE data-> '$.cluster'=?", clusterId); err == nil {
		for _, v := range namespace {
			if err := db.Delete(common.Namespace, v.Id); err != nil {
				log.Errorf("Delete namespace:%s error:%s", v.Id, err)
			}
			// 集群删除namespace
			//if err := r.Params.ClientSet.CoreV1().Namespaces().Delete(v.Name, &metav1.DeleteOptions{}); err != nil {
			//	log.Errorf("Delete namespace:%s for kubernetes error:%s", v.Id, err)
			//}
		}
		return nil
	} else {
		return err
	}
}

// 从User表中删除对应的Cluster
func (r *ClusterResource) DeleteClusterForUser(clusterId string) error {
	namespace := make([]*common.NamespaceDB, 0)
	if err := db.List(common.DataField, common.NamespaceTable, &namespace, "WHERE data-> '$.cluster'=?", clusterId); err != nil {
		log.Errorf("User table delete namespace error:%s", err)
		return err
	}
	user := make([]*common.User, 0)
	if err := db.List(common.DataField, common.UserTable, &user, ""); err == nil {
		for _, v := range user {
			v.Cluster = kit.DeleteItemForList(clusterId, v.Cluster)
			for _, ns := range namespace {
				v.Namespace = kit.DeleteItemForList(ns.Id, v.Namespace)
			}
			if err := db.Update(common.UserTable, v.Id, v); err != nil {
				log.Errorf("User table delete cluster :%s error:%s", v.Id, err)
			}
		}
		return nil
	} else {
		return err
	}
}
