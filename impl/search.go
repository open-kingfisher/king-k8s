package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-k8s/resource"
)

func GetSearch(c *gin.Context) {
	responseData := HandleSearch(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func HandleSearch(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取HTTP的参数，存到handle.Resources结构体中
	//commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.SearchResource{
		Params: handle.GenerateCommonParams(c, nil),
	}
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.Get()
		responseData = handle.HandlerResponse(response, err)
	}
	return
}
