package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-k8s/resource"
)

func GetCluster(c *gin.Context) {
	responseData := HandleCluster(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func ListCluster(c *gin.Context) {
	responseData := HandleCluster(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func DeleteCluster(c *gin.Context) {
	responseData := HandleCluster(common.Delete, c)
	c.JSON(responseData.Code, responseData)
}

func UpdateCluster(c *gin.Context) {
	responseData := HandleCluster(common.Update, c)
	c.JSON(responseData.Code, responseData)
}

func CreateCluster(c *gin.Context) {
	responseData := HandleCluster(common.Create, c)
	c.JSON(responseData.Code, responseData)
}

func HandleCluster(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, nil)
	r := resource.ClusterResource{Params: commonParams}
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.Get()
		responseData = handle.HandlerResponse(response, err)
	case common.List:
		response, err := r.List()
		responseData = handle.HandlerResponse(response, err)
	case common.Delete:
		err := r.Delete()
		responseData = handle.HandlerResponse(nil, err)
	case common.Update:
		err := r.Update(c)
		responseData = handle.HandlerResponse(nil, err)
	case common.Create:
		err := r.Create(c)
		responseData = handle.HandlerResponse(nil, err)
	}
	return
}
