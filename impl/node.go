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

func GetNode(c *gin.Context) {
	responseData := HandleNode(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func ListNode(c *gin.Context) {
	responseData := HandleNode(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func DeleteNode(c *gin.Context) {
	responseData := HandleNode(common.Delete, c)
	c.JSON(responseData.Code, responseData)
}

func PatchNode(c *gin.Context) {
	responseData := HandleNode(common.Patch, c)
	c.JSON(responseData.Code, responseData)
}

func UpdateNode(c *gin.Context) {
	responseData := HandleNode(common.Update, c)
	c.JSON(responseData.Code, responseData)
}

func CreateNode(c *gin.Context) {
	responseData := HandleNode(common.Create, c)
	c.JSON(responseData.Code, responseData)
}

func ListPodByNode(c *gin.Context) {
	responseData := HandleNode(common.ListPodByNode, c)
	c.JSON(responseData.Code, responseData)
}

func NodeMetric(c *gin.Context) {
	responseData := HandleNode(common.NodeMetric, c)
	c.JSON(responseData.Code, responseData)
}

func HandleNode(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.NodeResource{Params: commonParams}
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.Get()
		responseData = handle.HandlerResponse(response, err)
	case common.List:
		response, err := r.List()
		if err != nil {
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(response.Items, err)
		}
	case common.Delete:
		err := r.Delete()
		responseData = handle.HandlerResponse(nil, err)
	case common.Patch:
		if err := c.BindJSON(&r.Params.PatchData); err == nil {
			response, err := r.Patch()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.Update:
		if err := c.BindJSON(&r.PostData); err == nil {
			response, err := r.Update()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.Create:
		if err := r.GenerateCreateData(c); err == nil {
			if r.PostData != nil {
				response, err := r.Create()
				responseData = handle.HandlerResponse(response, err)
			} else {
				responseData = handle.HandlerResponse(nil, errors.New("the post data does not match the type"))
			}
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.ListPodByNode:
		response, err := r.ListPodByNode()
		if err != nil {
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(response.Items, err)
		}
	case common.NodeMetric:
		response, err := r.NodeMetric()
		if err != nil {
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(response, err)
		}
	}
	return
}
