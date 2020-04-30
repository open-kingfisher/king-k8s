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

func GetController(c *gin.Context) {
	responseData := HandleController(common.Get, c)
	c.JSON(http.StatusOK, responseData)
}

func ListController(c *gin.Context) {
	responseData := HandleController(common.List, c)
	c.JSON(http.StatusOK, responseData)
}

func DeleteController(c *gin.Context) {
	responseData := HandleController(common.Delete, c)
	c.JSON(http.StatusOK, responseData)
}

func PatchController(c *gin.Context) {
	responseData := HandleController(common.Patch, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdateController(c *gin.Context) {
	responseData := HandleController(common.Update, c)
	c.JSON(http.StatusOK, responseData)
}

func CreateController(c *gin.Context) {
	responseData := HandleController(common.Create, c)
	c.JSON(http.StatusOK, responseData)
}

func ScaleController(c *gin.Context) {
	responseData := HandleController(common.Scale, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdatePatchStepResumeController(c *gin.Context) {
	responseData := HandleController(common.PatchStepResume, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdatePatchAllResumeController(c *gin.Context) {
	responseData := HandleController(common.PatchAllResume, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdatePatchPauseController(c *gin.Context) {
	responseData := HandleController(common.PatchPause, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdatePatchImageController(c *gin.Context) {
	responseData := HandleController(common.PatchImage, c)
	c.JSON(http.StatusOK, responseData)
}

func UpdatePatchSyncImageController(c *gin.Context) {
	responseData := HandleController(common.PatchSyncImage, c)
	c.JSON(http.StatusOK, responseData)
}

func WatchPodIPController(c *gin.Context) {
	responseData := HandleController(common.WatchPodIP, c)
	c.JSON(http.StatusOK, responseData)
}

func ListPodByController(c *gin.Context) {
	responseData := HandleController(common.ListPodByController, c)
	c.JSON(http.StatusOK, responseData)
}

func SaveAsTemplate(c *gin.Context) {
	responseData := HandleController(common.SaveAsTemplate, c)
	c.JSON(http.StatusOK, responseData)
}

func GetControllerChart(c *gin.Context) {
	responseData := HandleController(common.GetChart, c)
	c.JSON(http.StatusOK, responseData)
}

func GetNamespaceIsExistLabel(c *gin.Context) {
	responseData := HandleController(common.GetNamespaceIsExistLabel, c)
	c.JSON(http.StatusOK, responseData)
}

func HandleController(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(c.Query("cluster"))
	responseData = handle.HandlerResponse(nil, err)
	if responseData.Code != http.StatusOK {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return
	}
	// 获取HTTP的参数，存到handle.Resources结构体中
	commonParams := handle.GenerateCommonParams(c, clientSet)
	r := resource.ControllerResource{Params: commonParams}
	// 调用结构体方法
	switch action {
	case common.Get:
		response, err := r.Get()
		responseData = handle.HandlerResponse(response, err)
	case common.List:
		response, err := r.List()
		responseData = handle.HandlerResponse(response, err)
	case common.ListPodByController:
		response, err := r.ListPodByController()
		responseData = handle.HandlerResponse(response, err)
	case common.GetChart:
		response, err := r.GetChart()
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
	case common.WatchPodIP:
		response, err := r.WatchPodIP()
		responseData = handle.HandlerResponse(response, err)
	case common.PatchSyncImage:
		if err := c.BindJSON(&r.Params.PatchData); err == nil {
			response, err := r.PatchSync()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.Scale:
		if err := c.BindJSON(&r.Params.PatchData); err == nil {
			response, err := r.Scale()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.PatchImage:
		if err := c.BindJSON(&r.Params.PatchData); err == nil {
			response, err := r.PatchImage()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.PatchStepResume:
		r.Params.PatchData = &common.PatchJson{}
		response, err := r.PatchStepResume()
		responseData = handle.HandlerResponse(response, err)
	case common.PatchAllResume:
		r.Params.PatchData = &common.PatchJson{}
		response, err := r.PatchAllResume()
		responseData = handle.HandlerResponse(response, err)
	case common.PatchPause:
		r.Params.PatchData = &common.PatchJson{}
		response, err := r.PatchPause()
		responseData = handle.HandlerResponse(response, err)
	case common.Update:
		switch r.Params.Controller {
		case "deployment":
			if err := c.BindJSON(&r.DeploymentData); err == nil {
				response, err := r.Update()
				responseData = handle.HandlerResponse(response, err)
			} else {
				responseData = handle.HandlerResponse(nil, err)
			}
		case "daemonset":
			if err := c.BindJSON(&r.DaemonSetData); err == nil {
				response, err := r.Update()
				responseData = handle.HandlerResponse(response, err)
			} else {
				responseData = handle.HandlerResponse(nil, err)
			}
		case "statefulset":
			if err := c.BindJSON(&r.StatefulSetData); err == nil {
				response, err := r.Update()
				responseData = handle.HandlerResponse(response, err)
			} else {
				responseData = handle.HandlerResponse(nil, err)
			}
		default:
			responseData = handle.HandlerResponse(nil, errors.New("controller kind doesn't exist"))
		}
	case common.SaveAsTemplate:
		if err := c.BindJSON(&r.TemplateData); err == nil {
			err := r.SaveAsTemplate()
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.Create:
		if err := r.GenerateCreateData(c); err == nil {
			response, err := r.Create()
			responseData = handle.HandlerResponse(response, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.GetNamespaceIsExistLabel:
		response, err := r.GetNamespaceIsExistLabel()
		responseData = handle.HandlerResponse(response, err)
	}
	return
}
