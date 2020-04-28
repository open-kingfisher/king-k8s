package impl

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-k8s/resource"
	"net/http"
)

type Hand interface {
	Get(c *gin.Context)
}

func Handle(c *gin.Context) {
	responseData := HandleNamespace(common.ActionType(c.Request.Method), c)
	c.JSON(responseData.Code, responseData)
}

func GetNamespace(c *gin.Context) {
	fmt.Println(c.Request.Method)
	responseData := HandleNamespace(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func ListNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.List, c)
	c.JSON(responseData.Code, responseData)
}
func ListAllNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.ListAll, c)
	c.JSON(responseData.Code, responseData)
}

func ListNamespaceAndCluster(c *gin.Context) {
	fmt.Println(c.Request.Method)
	responseData := HandleNamespace(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func DeleteNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.Delete, c)
	c.JSON(responseData.Code, responseData)
}

func PatchNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.Patch, c)
	c.JSON(responseData.Code, responseData)
}

func UpdateNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.Update, c)
	c.JSON(responseData.Code, responseData)
}

func CreateNamespace(c *gin.Context) {
	responseData := HandleNamespace(common.Create, c)
	c.JSON(responseData.Code, responseData)
}

func HandleNamespace(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.NamespaceResource{
		Params: commonParams,
		CustomParams: &resource.CustomParam{
			Exist: c.Query("exist"),
		},
	}
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.Get()
		responseData = handle.HandlerResponse(response, err)
	case common.List:
		response, err := r.List()
		responseData = handle.HandlerResponse(response, err)
	case common.ListAll:
		response, err := r.ListAll()
		responseData = handle.HandlerResponse(response, err)
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
	}
	return
}
