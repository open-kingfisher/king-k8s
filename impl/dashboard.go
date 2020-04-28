package impl

import (
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-k8s/resource"
	"net/http"
)

func ListInfoCard(c *gin.Context) {
	responseData := HandleDashboard("InfoCard", c)
	c.JSON(responseData.Code, responseData)
}

func ListApplication(c *gin.Context) {
	responseData := HandleDashboard("Application", c)
	c.JSON(responseData.Code, responseData)
}

func ListHistory(c *gin.Context) {
	responseData := HandleDashboard("History", c)
	c.JSON(responseData.Code, responseData)
}

func HandleDashboard(action string, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.DashboardResource{Params: commonParams}
	// 调用结构体方法
	switch action {
	case "InfoCard":
		response, err := r.ListInfoCard()
		responseData = handle.HandlerResponse(response, err)
	case "Application":
		response, err := r.ListApplication()
		responseData = handle.HandlerResponse(response, err)
	case "History":
		response, err := r.ListHistory()
		responseData = handle.HandlerResponse(response, err)
	}
	return
}
