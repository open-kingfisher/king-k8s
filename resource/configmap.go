package resource

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type ConfigMapResource struct {
	Params   *handle.Resources
	PostData *v1.ConfigMap
}

func (r *ConfigMapResource) Get() (*v1.ConfigMap, error) {
	return r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *ConfigMapResource) List() (*v1.ConfigMapList, error) {
	return r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *ConfigMapResource) Delete() (err error) {
	if err = r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ConfigMap,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ConfigMapResource) Patch() (res *v1.ConfigMap, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("ConfigMap patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ConfigMap,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *ConfigMapResource) Update() (res *v1.ConfigMap, err error) {
	if res, err = r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).Update(r.PostData); err != nil {
		log.Errorf("ConfigMap update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ConfigMap,
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

func (r *ConfigMapResource) Create() (res *v1.ConfigMap, err error) {
	if res, err = r.Params.ClientSet.CoreV1().ConfigMaps(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("ConfigMap create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.ConfigMap,
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

func (r *ConfigMapResource) GenerateCreateData(c *gin.Context) (err error) {
	data := make(map[string]string)
	// 获取KV值
	kv := c.PostForm("kv")
	if err = json.Unmarshal([]byte(kv), &data); err != nil {
		return
	}
	// 获取文件名
	fileName := c.PostForm("fileName")
	fileNameList := make([]string, 0)
	if err = json.Unmarshal([]byte(fileName), &fileNameList); err != nil {
		return
	}
	// 通过文件名获取文件内容，存储到data中
	for _, fileName := range fileNameList {
		if file, err := c.FormFile(fileName); err != nil {
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
			data[fileName] = string(buf)
		}
	}
	r.PostData = &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: c.Query("name"),
		},
		Data: data,
	}
	//switch r.Params.DataType {
	//case "yaml":
	//	var j []byte
	//	create := common.PostType{}
	//	if err = c.BindJSON(&create); err != nil {
	//		return
	//	}
	//	if j, _, err = kit.YamlToJson(create.Context); err != nil {
	//		return
	//	}
	//	if err = json.Unmarshal(j, &r.PostData); err != nil {
	//		return
	//	}
	//case "json":
	//	if err = c.BindJSON(&r.PostData); err != nil {
	//		return
	//	}
	//default:
	//	return errors.New(common.ContentTypeError)
	//}
	return nil
}
