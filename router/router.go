package router

import (
	"github.com/gin-gonic/gin"
	"github.com/open-kingfisher/king-k8s/impl"
	"github.com/open-kingfisher/king-utils/common"
	jwtAuth "github.com/open-kingfisher/king-utils/middleware/jwt"
	"golang.org/x/net/websocket"
	"net/http"
)

func SetupRouter(r *gin.Engine) *gin.Engine {
	//重新定义404
	r.NoRoute(NoRoute)

	// web terminal
	r.GET(common.K8SPath+"terminal", func(c *gin.Context) {
		impl := websocket.Handler(impl.Terminal)
		impl.ServeHTTP(c.Writer, c.Request)
	})

	authorize := r.Group("/", jwtAuth.JWTAuth())
	{
		// nodes
		authorize.GET(common.K8SPath+"nodes", impl.ListNode)
		authorize.GET(common.K8SPath+"nodes/:name", impl.GetNode)
		authorize.DELETE(common.K8SPath+"nodes/:name", impl.DeleteNode)
		authorize.PUT(common.K8SPath+"nodes/:name", impl.UpdateNode)
		authorize.POST(common.K8SPath+"nodes", impl.CreateNode)
		authorize.PATCH(common.K8SPath+"nodes/:name", impl.PatchNode)
		authorize.GET(common.K8SPath+"listPodByNode/:name", impl.ListPodByNode)
		authorize.GET(common.K8SPath+"nodeMetric/:name", impl.NodeMetric)

		// namespace
		authorize.GET(common.K8SPath+"namespace", impl.ListNamespace)
		authorize.GET(common.K8SPath+"namespaceAll", impl.ListAllNamespace)
		authorize.GET(common.K8SPath+"namespace/:name", impl.GetNamespace)
		authorize.DELETE(common.K8SPath+"namespace/:name", impl.DeleteNamespace)
		authorize.POST(common.K8SPath+"namespace", impl.CreateNamespace)
		authorize.PATCH(common.K8SPath+"namespace/:name", impl.PatchNamespace)
		authorize.PUT(common.K8SPath+"namespace", impl.UpdateNamespace)

		// pod
		authorize.GET(common.K8SPath+"pod", impl.ListPod)
		authorize.GET(common.K8SPath+"pod/:name", impl.GetPod)
		authorize.GET(common.K8SPath+"pod/:name/log", impl.LogPod)
		authorize.GET(common.K8SPath+"pod/:name/debug", impl.DebugPod)
		// pod 救援模式
		authorize.POST(common.K8SPath+"pod/:name/rescue", impl.RescuePod)
		authorize.DELETE(common.K8SPath+"pod/:name", impl.DeletePod)
		authorize.PATCH(common.K8SPath+"pod/patch/:name", impl.PatchPod)
		authorize.PATCH(common.K8SPath+"pod/evict/:name", impl.EvictPod)
		authorize.GET(common.K8SPath+"kubectl/install", impl.KubectlPod)
		authorize.GET(common.K8SPath+"kubectl/uninstall", impl.UnKubectlPod)
		authorize.GET(common.K8SPath+"debug/podIP/:name", impl.GetDebugPodIPByPod)

		// service
		authorize.GET(common.K8SPath+"service", impl.ListService)
		authorize.GET(common.K8SPath+"service/:name", impl.GetService)
		authorize.GET(common.K8SPath+"listPodByService/:name", impl.ListPodByService)
		authorize.DELETE(common.K8SPath+"service/:name", impl.DeleteService)
		authorize.PATCH(common.K8SPath+"service/patch/:name", impl.PatchService)
		authorize.POST(common.K8SPath+"service", impl.CreateService)
		authorize.PUT(common.K8SPath+"service", impl.UpdateService)

		// limit range
		authorize.GET(common.K8SPath+"limitrange", impl.ListLimitRange)
		authorize.GET(common.K8SPath+"limitrange/:name", impl.GetLimitRange)
		authorize.DELETE(common.K8SPath+"limitrange/:name", impl.DeleteLimitRange)
		authorize.PATCH(common.K8SPath+"limitrange/patch/:name", impl.PatchLimitRange)
		authorize.POST(common.K8SPath+"limitrange", impl.CreateLimitRange)
		authorize.PUT(common.K8SPath+"limitrange", impl.UpdateLimitRange)

		// resource quotas
		authorize.GET(common.K8SPath+"resourcequota", impl.ListResourceQuota)
		authorize.GET(common.K8SPath+"resourcequota/:name", impl.GetResourceQuota)
		authorize.DELETE(common.K8SPath+"resourcequota/:name", impl.DeleteResourceQuota)
		authorize.PATCH(common.K8SPath+"resourcequota/patch/:name", impl.PatchResourceQuota)
		authorize.POST(common.K8SPath+"resourcequota", impl.CreateResourceQuota)
		authorize.PUT(common.K8SPath+"resourcequota", impl.UpdateResourceQuota)

		// controller (Deployment DaemonSet StatefulSet)
		authorize.GET(common.K8SPath+"controller/:controller", impl.ListController)
		authorize.GET(common.K8SPath+"controller/:controller/:name", impl.GetController)
		authorize.GET(common.K8SPath+"listPodByController/:controller/:name", impl.ListPodByController)
		authorize.DELETE(common.K8SPath+"controller/:controller/:name", impl.DeleteController)
		authorize.PATCH(common.K8SPath+"controller/:controller/patch/:name", impl.PatchController)
		authorize.PATCH(common.K8SPath+"controller/:controller/restart/:name", impl.RestartController)
		// 使用分步滚动更新的时候使用此接口
		authorize.PATCH(common.K8SPath+"controller/:controller/patchSync/:name", impl.UpdatePatchSyncImageController)
		authorize.PATCH(common.K8SPath+"controller/:controller/rolling/:name", impl.UpdatePatchImageController)
		authorize.PATCH(common.K8SPath+"controller/:controller/step/resume/:name", impl.UpdatePatchStepResumeController)
		authorize.PATCH(common.K8SPath+"controller/:controller/all/resume/:name", impl.UpdatePatchAllResumeController)
		authorize.PATCH(common.K8SPath+"controller/:controller/pause/:name", impl.UpdatePatchPauseController)
		authorize.PATCH(common.K8SPath+"controller/:controller/watch/:name", impl.WatchPodIPController)
		// 接口调用报错
		//authorize.PATCH(common.K8SPath+"controller/:controller/scale/:name", impl.ScaleController)
		authorize.POST(common.K8SPath+"controller", impl.CreateController)
		authorize.PUT(common.K8SPath+"controller/:controller", impl.UpdateController)

		authorize.GET(common.K8SPath+"controllerChart/:controller/:name", impl.GetControllerChart)
		authorize.PUT(common.K8SPath+"template/:controller/:name", impl.SaveAsTemplate)
		authorize.GET(common.K8SPath+"namespaceLabel/:name", impl.GetNamespaceIsExistLabel)

		// replica set
		authorize.GET(common.K8SPath+"replicaset", impl.ListReplicaSet)
		authorize.GET(common.K8SPath+"replicaset/:name", impl.GetReplicaSet)
		authorize.DELETE(common.K8SPath+"replicaset/:name", impl.DeleteReplicaSet)

		// secret
		authorize.GET(common.K8SPath+"secret", impl.ListSecret)
		authorize.GET(common.K8SPath+"secret/:name", impl.GetSecret)
		authorize.DELETE(common.K8SPath+"secret/:name", impl.DeleteSecret)
		authorize.PATCH(common.K8SPath+"secret/patch/:name", impl.PatchSecret)
		authorize.POST(common.K8SPath+"secret", impl.CreateSecret)
		authorize.PUT(common.K8SPath+"secret", impl.UpdateSecret)

		// config map
		authorize.GET(common.K8SPath+"configmap", impl.ListConfigMap)
		authorize.GET(common.K8SPath+"configmap/:name", impl.GetConfigMap)
		authorize.DELETE(common.K8SPath+"configmap/:name", impl.DeleteConfigMap)
		authorize.PATCH(common.K8SPath+"configmap/patch/:name", impl.PatchConfigMap)
		authorize.POST(common.K8SPath+"configmap", impl.CreateConfigMap)
		authorize.PUT(common.K8SPath+"configmap", impl.UpdateConfigMap)

		// event
		authorize.GET(common.K8SPath+"event", impl.ListEvent)

		// ingress
		authorize.GET(common.K8SPath+"ingress", impl.ListIngress)
		authorize.GET(common.K8SPath+"ingress/:name", impl.GetIngress)
		authorize.GET(common.K8SPath+"ingressChart/:name", impl.GetIngressChart)
		authorize.DELETE(common.K8SPath+"ingress/:name", impl.DeleteIngress)
		authorize.PATCH(common.K8SPath+"ingress/patch/:name", impl.PatchIngress)
		authorize.POST(common.K8SPath+"ingress", impl.CreateIngress)
		authorize.PUT(common.K8SPath+"ingress", impl.UpdateIngress)

		// role
		authorize.GET(common.K8SPath+"role", impl.ListRole)
		authorize.GET(common.K8SPath+"role/:name", impl.GetRole)
		authorize.DELETE(common.K8SPath+"role/:name", impl.DeleteRole)
		authorize.PATCH(common.K8SPath+"role/patch/:name", impl.PatchRole)
		authorize.POST(common.K8SPath+"role", impl.CreateRole)
		authorize.PUT(common.K8SPath+"role", impl.UpdateRole)

		// cluster role
		authorize.GET(common.K8SPath+"clusterrole", impl.ListClusterRole)
		authorize.GET(common.K8SPath+"clusterrole/:name", impl.GetClusterRole)
		authorize.DELETE(common.K8SPath+"clusterrole/:name", impl.DeleteClusterRole)
		authorize.PATCH(common.K8SPath+"clusterrole/patch/:name", impl.PatchClusterRole)
		authorize.POST(common.K8SPath+"clusterrole", impl.CreateClusterRole)
		authorize.PUT(common.K8SPath+"clusterrole", impl.UpdateClusterRole)

		// Horizontal Pod Auto scaler
		authorize.GET(common.K8SPath+"hpa", impl.ListHPA)
		authorize.GET(common.K8SPath+"hpa/:name", impl.GetHPA)
		authorize.DELETE(common.K8SPath+"hpa/:name", impl.DeleteHPA)
		authorize.PATCH(common.K8SPath+"hpa/patch/:name", impl.PatchHPA)
		authorize.POST(common.K8SPath+"hpa", impl.CreateHPA)
		authorize.PUT(common.K8SPath+"hpa", impl.UpdateHPA)

		// role binding
		authorize.GET(common.K8SPath+"bind/role", impl.ListRoleBinding)
		authorize.GET(common.K8SPath+"bind/role/:name", impl.GetRoleBinding)
		authorize.DELETE(common.K8SPath+"bind/role/:name", impl.DeleteRoleBinding)
		authorize.PATCH(common.K8SPath+"bind/role/patch/:name", impl.PatchRoleBinding)
		authorize.POST(common.K8SPath+"bind/role", impl.CreateRoleBinding)
		authorize.PUT(common.K8SPath+"bind/role", impl.UpdateRoleBinding)

		// cluster role binding
		authorize.GET(common.K8SPath+"bind/clusterrole", impl.ListClusterRoleBinding)
		authorize.GET(common.K8SPath+"bind/clusterrole/:name", impl.GetClusterRoleBinding)
		authorize.DELETE(common.K8SPath+"bind/clusterrole/:name", impl.DeleteClusterRoleBinding)
		authorize.PATCH(common.K8SPath+"bind/clusterrole/patch/:name", impl.PatchClusterRoleBinding)
		authorize.POST(common.K8SPath+"bind/clusterrole", impl.CreateClusterRoleBinding)
		authorize.PUT(common.K8SPath+"bind/clusterrole", impl.UpdateClusterRoleBinding)

		// service account
		authorize.GET(common.K8SPath+"serviceaccount", impl.ListServiceAccount)
		authorize.GET(common.K8SPath+"serviceaccount/:name", impl.GetServiceAccount)
		authorize.DELETE(common.K8SPath+"serviceaccount/:name", impl.DeleteServiceAccount)
		authorize.PATCH(common.K8SPath+"serviceaccount/patch/:name", impl.PatchServiceAccount)
		authorize.POST(common.K8SPath+"serviceaccount", impl.CreateServiceAccount)
		authorize.PUT(common.K8SPath+"serviceaccount", impl.UpdateServiceAccount)

		// persistent volume claims
		authorize.GET(common.K8SPath+"pvc", impl.ListPVC)
		authorize.GET(common.K8SPath+"pvc/:name", impl.GetPVC)
		authorize.DELETE(common.K8SPath+"pvc/:name", impl.DeletePVC)
		authorize.PATCH(common.K8SPath+"pvc/patch/:name", impl.PatchPVC)
		authorize.POST(common.K8SPath+"pvc", impl.CreatePVC)
		authorize.PUT(common.K8SPath+"pvc", impl.UpdatePVC)

		// persistent volume
		authorize.GET(common.K8SPath+"pv", impl.ListPV)
		authorize.GET(common.K8SPath+"pv/:name", impl.GetPV)
		authorize.DELETE(common.K8SPath+"pv/:name", impl.DeletePV)
		authorize.PATCH(common.K8SPath+"pv/patch/:name", impl.PatchPV)
		authorize.POST(common.K8SPath+"pv", impl.CreatePV)
		authorize.PUT(common.K8SPath+"pv", impl.UpdatePV)

		// storage classes
		authorize.GET(common.K8SPath+"sc", impl.ListStorageClasses)
		authorize.GET(common.K8SPath+"sc/:name", impl.GetStorageClasses)
		authorize.DELETE(common.K8SPath+"sc/:name", impl.DeleteStorageClasses)
		authorize.PATCH(common.K8SPath+"sc/patch/:name", impl.PatchStorageClasses)
		authorize.POST(common.K8SPath+"sc", impl.CreateStorageClasses)
		authorize.PUT(common.K8SPath+"sc", impl.UpdateStorageClasses)

		// metrics
		authorize.GET(common.K8SPath+"metrics/node", impl.ListNodeMetrics)
		authorize.GET(common.K8SPath+"metrics/node/:name", impl.GetNodeMetrics)
		authorize.GET(common.K8SPath+"metrics/pod", impl.ListPodMetrics)
		authorize.GET(common.K8SPath+"metrics/pod/:name", impl.GetPodMetrics)

		// custom metrics
		authorize.GET(common.K8SPath+"metrics/custom/:metricsKind/:name/:metricsName", impl.GetCustomMetrics)
		authorize.GET(common.K8SPath+"metrics/customs/:metricsKind/:name/:metricsName", impl.ListCustomMetrics)

		// dashboard
		authorize.GET(common.K8SPath+"dashboard/infocard", impl.ListInfoCard)
		authorize.GET(common.K8SPath+"dashboard/application", impl.ListApplication)
		authorize.GET(common.K8SPath+"dashboard/history", impl.ListHistory)
		authorize.GET(common.K8SPath+"dashboard/podStatus", impl.ListPodStatus)

		// Ping test
		authorize.GET("/ping", impl.Ping)

		// harbor
		authorize.GET(common.K8SPath+"harbor/projects", impl.ListProjects)
		authorize.GET(common.K8SPath+"harbor/projects/:id", impl.ListImages)
		authorize.GET(common.K8SPath+"harbor/tags/:project/:image", impl.ListTags)

		// cluster
		authorize.GET(common.K8SPath+"cluster", impl.ListCluster)
		authorize.GET(common.K8SPath+"cluster/:name", impl.GetCluster)
		authorize.DELETE(common.K8SPath+"cluster/:name", impl.DeleteCluster)
		authorize.POST(common.K8SPath+"cluster", impl.CreateCluster)
		authorize.PUT(common.K8SPath+"cluster", impl.UpdateCluster)

		// plugin
		authorize.GET(common.K8SPath+"clusterplugin", impl.ListClusterPlugin)
		authorize.GET(common.K8SPath+"clusterplugin/status", impl.StatusClusterPlugin)
		authorize.POST(common.K8SPath+"clusterplugin", impl.CreateClusterPlugin)

		// prometheus
		authorize.GET(common.K8SPath+"prometheus/node/:name", impl.PNodeMetrics)

		// search pod ip
		authorize.GET(common.K8SPath+"search/podip/:name", impl.GetSearch)

		// endpoint
		authorize.GET(common.K8SPath+"endpoint", impl.ListEndpoint)
		authorize.GET(common.K8SPath+"endpoint/:name", impl.GetEndpoint)
		authorize.DELETE(common.K8SPath+"endpoint/:name", impl.DeleteEndpoint)
		authorize.PATCH(common.K8SPath+"endpoint/patch/:name", impl.PatchEndpoint)
		authorize.POST(common.K8SPath+"endpoint", impl.CreateEndpoint)
		authorize.PUT(common.K8SPath+"endpoint", impl.UpdateEndpoint)
	}
	return r
}

// 重新定义404错误
func NoRoute(c *gin.Context) {
	responseData := common.ResponseData{Code: http.StatusNotFound, Msg: "404 Not Found"}
	c.JSON(http.StatusNotFound, responseData)
}
