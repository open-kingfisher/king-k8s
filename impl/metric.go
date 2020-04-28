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

func GetNodeMetrics(c *gin.Context) {
	responseData := HandleMetrics(common.Get, common.Node, c)
	c.JSON(responseData.Code, responseData)
}

func ListNodeMetrics(c *gin.Context) {
	responseData := HandleMetrics(common.List, common.Node, c)
	c.JSON(responseData.Code, responseData)
}

func GetPodMetrics(c *gin.Context) {
	responseData := HandleMetrics(common.Get, common.Pod, c)
	c.JSON(responseData.Code, responseData)
}

func ListPodMetrics(c *gin.Context) {
	responseData := HandleMetrics(common.List, common.Pod, c)
	c.JSON(responseData.Code, responseData)
}

func GetCustomMetrics(c *gin.Context) {
	responseData := HandleCustomMetrics(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func ListCustomMetrics(c *gin.Context) {
	responseData := HandleCustomMetrics(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func HandleMetrics(action common.ActionType, metrics string, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	metricsClient, err := access.MetricsClient(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, nil)
	r := resource.MetricResource{Params: commonParams}
	r.MetricsClient = metricsClient
	// 调用结构体方法
	switch action {
	case common.Get:
		if metrics == common.Pod {
			response, err := r.GetPodMetrics()
			responseData = handle.HandlerResponse(response, err)
		} else if metrics == common.Node {
			response, err := r.GetNodeMetrics()
			responseData = handle.HandlerResponse(response, err)
		}
	case common.List:
		if metrics == common.Pod {
			response, err := r.ListPodMetrics()
			responseData = handle.HandlerResponse(response.Items, err)
		} else if metrics == common.Node {
			response, err := r.ListNodeMetrics()
			responseData = handle.HandlerResponse(response.Items, err)
		}
	}
	return
}

func HandleCustomMetrics(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	CustomMMetricsClient, err := access.CustomMetricsClient(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, nil)
	r := resource.MetricResource{Params: commonParams}
	r.CustomMetricsClient = CustomMMetricsClient
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.GetCustomMetrics()
		responseData = handle.HandlerResponse(response, err)
	case common.List:
		response, err := r.ListCustomMetrics()
		if err != nil {
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(response.Items, err)
		}
	}
	return
}
