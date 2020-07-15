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
	"strconv"
)

const (
	DefaultDebugImage = "registry.kingfisher.com/kingfisher/netshoot"
	KingDebug         = "king-debug"
	KingKubectl       = "king-kubectl"
)

func GetPod(c *gin.Context) {
	responseData := HandlePod(common.Get, c)
	c.JSON(responseData.Code, responseData)
}

func ListPod(c *gin.Context) {
	responseData := HandlePod(common.List, c)
	c.JSON(responseData.Code, responseData)
}

func DeletePod(c *gin.Context) {
	responseData := HandlePod(common.Delete, c)
	c.JSON(responseData.Code, responseData)
}

func PatchPod(c *gin.Context) {
	responseData := HandlePod(common.Patch, c)
	c.JSON(responseData.Code, responseData)
}

func LogPod(c *gin.Context) {
	responseData := HandlePod(common.Log, c)
	c.JSON(responseData.Code, responseData)
}

func EvictPod(c *gin.Context) {
	responseData := HandlePod(common.Evict, c)
	c.JSON(responseData.Code, responseData)
}

func DebugPod(c *gin.Context) {
	responseData := HandlePod(common.Debug, c)
	c.JSON(responseData.Code, responseData)
}

func RescuePod(c *gin.Context) {
	responseData := HandlePod(common.Rescue, c)
	c.JSON(responseData.Code, responseData)
}

func KubectlPod(c *gin.Context) {
	responseData := HandlePod(common.Kubectl, c)
	c.JSON(responseData.Code, responseData)
}

func UnKubectlPod(c *gin.Context) {
	responseData := HandlePod(common.UnKubectl, c)
	c.JSON(responseData.Code, responseData)
}

func GetDebugPodIPByPod(c *gin.Context) {
	responseData := HandlePod(common.DebugPodIPByPod, c)
	c.JSON(responseData.Code, responseData)
}

func OfflinePod(c *gin.Context) {
	responseData := HandlePod(common.Offline, c)
	c.JSON(responseData.Code, responseData)
}

func OnlinePod(c *gin.Context) {
	responseData := HandlePod(common.Online, c)
	c.JSON(responseData.Code, responseData)
}

func HandlePod(action common.ActionType, c *gin.Context) (responseData *common.ResponseData) {
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
	sinceSeconds, err := strconv.Atoi(c.Query("sinceSeconds"))
	if err != nil {
		// 1分钟
		sinceSeconds = 1 * 60
		//log.Errorf("%s%s", "the time format is not correct", err)
	}
	sinceSecondsInt64 := int64(sinceSeconds)

	r := resource.PodResource{Params: commonParams,
		SinceSeconds: &sinceSecondsInt64,
		Container:    c.Query("container"),
	}
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
	case common.Log:
		response, err := r.Log()
		responseData = handle.HandlerResponse(response, err)
	case common.Evict:
		err := r.Evict()
		responseData = handle.HandlerResponse(nil, err)
	case common.Offline:
		response, err := r.Offline()
		responseData = handle.HandlerResponse(response, err)
	case common.Online:
		response, err := r.Online()
		responseData = handle.HandlerResponse(response, err)
	case common.Debug:
		var image, debugImage string
		if c.Query("image") == "" {
			image = common.REGISTRYURL + KingDebug
		} else {
			image = c.Query("image")
		}
		if c.Query("debugImage") == "" {
			debugImage = DefaultDebugImage
		} else {
			debugImage = c.Query("debugImage")
		}
		r.Container = c.Query("container")
		r.Image = image
		r.DebugImage = debugImage
		r.EntryPoint = c.Query("entryPoint")
		response, err := r.Debug()
		responseData = handle.HandlerResponse(response, err)
	case common.Rescue:
		r.Container = c.Query("container")
		//r.EntryPoint = c.Query("entryPoint")
		condition := r.RescueCondition
		if err = c.BindJSON(&condition); err == nil {
			r.RescueCondition = condition
			err := r.Rescue()
			responseData = handle.HandlerResponse(nil, err)
		} else {
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.Kubectl:
		r.Image = common.REGISTRYURL + KingKubectl
		//version := c.Query("kubectlVersion")
		//if !strings.HasPrefix(version, "v") {
		//	version = "v" + version
		//}
		if version, err := clientSet.Discovery().ServerVersion(); err != nil {
			responseData = handle.HandlerResponse(nil, err)
		} else {
			r.KubectlVersion = version.GitVersion
			r.Plugin = c.Query("plugin")
			err := r.Kubectl()
			responseData = handle.HandlerResponse(nil, err)
		}
	case common.UnKubectl:
		r.Plugin = c.Query("plugin")
		err := r.UnKubectl()
		responseData = handle.HandlerResponse(nil, err)
	case common.DebugPodIPByPod:
		response, err := r.GetDebugPodIPByPod()
		responseData = handle.HandlerResponse(response, err)
	}
	return
}
