package impl

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-k8s/resource"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"net/http"
)

func PNodeMetrics(c *gin.Context) {
	responseData := HandlePrometheus(common.NodeMetric, c)
	c.JSON(responseData.Code, responseData)
}

func HandlePrometheus(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.PrometheusClient()
	if err != nil && err.Error() == common.ClusterNotExistError {
		err = errors.New("prometheus does not exist")
	}
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.PrometheusClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, nil)
	r := resource.PrometheusResource{
		Params:    commonParams,
		ClientSet: clientSet,
	}
	// 调用结构体方法
	switch action {
	case common.NodeMetric:
		response, err := r.NodeMetric()
		responseData = handle.HandlerResponse(response, err)
	}
	return
}
