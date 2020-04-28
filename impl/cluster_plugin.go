package impl

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-k8s/resource"
	"net/http"
)

func ListClusterPlugin(c *gin.Context) {
	responseData := HandleClusterPlugin(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func DeleteClusterPlugin(c *gin.Context) {
	responseData := HandleClusterPlugin(common.Delete, c)
	c.JSON(responseData.Code, responseData)
}

func UpdateClusterPlugin(c *gin.Context) {
	responseData := HandleClusterPlugin(common.Update, c)
	c.JSON(responseData.Code, responseData)
}

func CreateClusterPlugin(c *gin.Context) {
	responseData := HandleClusterPlugin(common.Create, c)
	c.JSON(responseData.Code, responseData)
}
func StatusClusterPlugin(c *gin.Context) {
	responseData := HandleClusterPlugin(common.Status, c)
	c.JSON(responseData.Code, responseData)
}

func HandleClusterPlugin(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(c.Query("cluster"))
	if err != nil && err.Error() == common.ClusterNotExistError {
		err = errors.New("cluster does not exist")
	}
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.ClusterPluginResource{Params: commonParams}
	// 调用结构体方法
	switch action {
	case common.List:
		response, err := r.List()
		responseData = handle.HandlerResponse(response, err)
	case common.Delete:
		err := r.Delete()
		responseData = handle.HandlerResponse(nil, err)
	//case common.Update:
	//	err := r.Update(c)
	//	responseData = handle.HandlerResponse(nil, err)
	case common.Status:
		r.Plugin = c.Query("plugin")
		response, err := r.Status()
		responseData = handle.HandlerResponse(response, err)
	case common.Create:
		err := r.Create(c)
		responseData = handle.HandlerResponse(nil, err)
	}
	return
}
