package resource

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	appv1 "k8s.io/api/apps/v1"
	"k8s.io/api/core/v1"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	watchType "k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/watch"
	"k8s.io/kubernetes/pkg/client/conditions"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	"github.com/open-kingfisher/king-utils/kit"
	"strings"
	"time"
)

const (
	ServiceAccountExistsError     = "serviceaccounts \"" + common.ServiceAccountName + "\" already exists"
	ClusterRoleBindingExistsError = "clusterrolebindings.rbac.authorization.k8s.io \"" + common.ClusterRoleBindingName + "\" already exists"
	RescueConditionVolume         = "volume"
	RescueConditionEnv            = "env"
	RescueConditionInitContainers = "initContainers"
	RescueConditionIstioInject    = "istioInject"
	RescueConditionPrivileged     = "privileged"
	RescueConditionAffinity       = "affinity"
	RescueConditionToleration     = "toleration"
	DebugSockPort                 = 9091
	DebugPodPrefix                = "debug-pod-"
)

var (
	DisableIsitoInject = map[string]string{"sidecar.istio.io/inject": "false"}
)

type PodResource struct {
	Params          *handle.Resources
	PostData        *v1.Pod
	SinceSeconds    *int64   `json:"sinceSeconds"`
	Container       string   `json:"container"`
	Image           string   `json:"image"`
	DebugImage      string   `json:"debugImage"`
	EntryPoint      string   `json:"entryPoint"`
	KubectlVersion  string   `json:"kubectlVersion "`
	Plugin          string   `json:"plugin "`
	RescueCondition []string `json:"rescueCondition"`
}

func (r *PodResource) Get() (*v1.Pod, error) {
	go func() {
		watcher, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Watch(metav1.ListOptions{})
		if err != nil {
			log.Errorf("Pod watch error: %s", err)
		}
		for item := range watcher.ResultChan() {
			fmt.Println("##################################################")
			fmt.Println(item)
			fmt.Println("##################################################")
		}
	}()
	return r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *PodResource) List() (*v1.PodList, error) {
	podItems := &v1.PodList{}
	if pods, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		if r.Params.Uid != "" {
			for _, p := range pods.Items {
				owner := p.ObjectMeta.OwnerReferences
				for _, controller := range owner {
					if string(controller.UID) == r.Params.Uid {
						podItems.Items = append(podItems.Items, p)
					}
				}
			}
		} else {
			podItems = pods
		}
	} else {
		return podItems, err
	}
	return podItems, nil
}

func (r *PodResource) Delete() (err error) {
	pod, _ := r.Get()
	if err = r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Delete(r.Params.Name, &metav1.DeleteOptions{}); err != nil {
		return
	}
	// 判断pod已经删除
	watcher, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Watch(metav1.SingleObject(pod.ObjectMeta))
	for v := range watcher.ResultChan() {
		if v.Type == watchType.Deleted {
			watcher.Stop()
		}
	}
	if err != nil {
		return err
	}
	auditLog := handle.AuditLog{
		Kind:       common.Pod,
		ActionType: common.Delete,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PodResource) Patch() (res *v1.Pod, err error) {
	var data []byte
	if data, err = json.Marshal(r.Params.PatchData.Patches); err != nil {
		return
	}
	if res, err = r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Patch(r.Params.Name, types.JSONPatchType, data); err != nil {
		log.Errorf("Pod patch error:%s; Json:%+v; Name:%s", err, string(data), r.Params.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Pod,
		ActionType: common.Patch,
		Resources:  r.Params,
		Name:       r.Params.Name,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PodResource) Update() (res *v1.Pod, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Update(r.PostData); err != nil {
		log.Errorf("Pod update error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Pod,
		ActionType: common.Update,
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PodResource) Create() (res *v1.Pod, err error) {
	if res, err = r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Create(r.PostData); err != nil {
		log.Errorf("Pod create error:%s; Json:%+v; Name:%s", err, r.PostData, r.PostData.Name)
		return
	}
	auditLog := handle.AuditLog{
		Kind:       common.Pod,
		ActionType: common.Create,
		Resources:  r.Params,
		Name:       r.PostData.Name,
		PostData:   r.PostData,
	}
	if err = auditLog.InsertAuditLog(); err != nil {
		return
	}
	return
}

func (r *PodResource) Log() (*string, error) {
	logRequest := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).GetLogs(r.Params.Name, &v1.PodLogOptions{SinceSeconds: r.SinceSeconds, Container: r.Container})
	logResponse := logRequest.Do()
	if podLog, err := logResponse.Raw(); err == nil {
		logs := string(podLog)
		return &logs, nil
	} else {
		log.Errorf("Pod log error:%s; Json:%+v; Name:%s", err, r, r.Params.Name)
		return nil, err
	}
}

func (r *PodResource) Evict() (err error) {
	// https://kubernetes.io/docs/tasks/administer-cluster/safely-drain-node/#the-eviction-api
	return r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Evict(&policy.Eviction{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Eviction",
			APIVersion: "policy/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Params.Name,
			Namespace: r.Params.Namespace,
		},
	})
}

// 用于把Pod绑定到指定的主机上面
func (r *PodResource) Bind() (err error) {
	return r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Bind(&v1.Binding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Params.Name,
			Namespace: r.Params.Namespace,
		},
		Target: v1.ObjectReference{
			Kind: "Node",
			Name: "node name",
		},
	})
}

func (r *PodResource) Event() error {
	go func() {
		watcher, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Watch(metav1.ListOptions{})
		if err != nil {
			log.Errorf("Pod watch error: %s", err)
		}
		for item := range watcher.ResultChan() {
			fmt.Println(item)
		}
	}()
	return nil
}

// 创建Debug Pod
func (r *PodResource) Debug() (interface{}, error) {
	podInfo, err := r.getContainerIDAndNode()
	if err != nil {
		log.Errorf("getContainerIDAndNode error: %s", err)
		return nil, err
	}
	podSpec, err := r.getDebugPodSpec(podInfo)
	if err != nil {
		log.Errorf("getDebugPodSpec error: %s", err)
		return nil, err
	}
	pod, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Create(podSpec)
	if err != nil {
		log.Errorf("Pod create error:%s; Json:%+v; Name:%s", err, podSpec, r.Params.Name)
		return nil, err
	}
	watcher, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Watch(metav1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// 读取watch items，直到达到满足的条件
	_, err = watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		return nil, err
	}
	return r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Get(pod.Name, metav1.GetOptions{})
}

func (r *PodResource) getContainerIDAndNode() (map[string]string, error) {
	if pod, err := r.Get(); err != nil {
		return nil, err
	} else {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name != r.Container {
				continue
			}
			if !containerStatus.Ready {
				return nil, fmt.Errorf("container [%s] not ready", r.Container)
			}
			return map[string]string{
				"ContainerID": containerStatus.ContainerID,
				"NodeName":    pod.Spec.NodeName,
				"UID":         string(pod.ObjectMeta.UID),
				"Name":        pod.ObjectMeta.Name,
			}, nil
		}

		for _, initContainerStatus := range pod.Status.InitContainerStatuses {
			if initContainerStatus.Name != r.Container {
				continue
			}
			if initContainerStatus.State.Running == nil {
				return nil, fmt.Errorf("init container [%s] is not running", r.Container)
			}
			return map[string]string{
				"ContainerID": initContainerStatus.ContainerID,
				"NodeName":    pod.Spec.NodeName,
				"UID":         string(pod.ObjectMeta.UID),
				"Name":        pod.ObjectMeta.Name,
			}, nil
		}

	}
	return nil, fmt.Errorf("cannot find specified container [%s]", r.Container)
}

// Debug Pod创建的描述
func (r *PodResource) getDebugPodSpec(podInfo map[string]string) (*v1.Pod, error) {
	containerID := strings.Split(podInfo["ContainerID"], "//")
	if len(containerID) != 2 {
		return nil, errors.New("get container id failed")
	}
	t := false
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        DebugPodPrefix + r.Params.Name, // debug-pod加上要debug的容器名称
			Namespace:   r.Params.Namespace,
			Labels:      map[string]string{"debug-pod": "enable"},
			Annotations: DisableIsitoInject, // 不开启istio注入
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion:         "v1",
					Kind:               "Pod",
					Name:               podInfo["Name"],
					UID:                types.UID(podInfo["UID"]),
					Controller:         &t, // 如果是个控制器
					BlockOwnerDeletion: &t, // 不阻塞Owner的删除
				},
			},
		},
		Spec: v1.PodSpec{
			NodeName: podInfo["NodeName"], // 必须是要debug的Pod所在的节点
			Containers: []v1.Container{
				{
					Name:            "debug-pod",
					Image:           r.Image, // 此镜像由king-debug项目生成
					ImagePullPolicy: v1.PullAlways,
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      "docker",
							MountPath: "/var/run/docker.sock", // 挂载的卷，此Pod的主要工作就是通过节点的docker.sock创建容器
						},
					},
					Env: []v1.EnvVar{
						{Name: "DEBUG_IMAGE", Value: r.DebugImage},
						{Name: "CONTAINER_ID", Value: containerID[1]},
						{Name: "ENTRY_POINT", Value: r.EntryPoint},
					},
					Lifecycle: &v1.Lifecycle{ // Pod的生命周期管理
						PostStart: &v1.Handler{ // 在容器创建之后，容器的Entrypoint执行之前，执行debug拉取镜像创建容器
							Exec: &v1.ExecAction{
								Command: []string{"/usr/local/bin/debug"},
							},
						},
						PreStop: &v1.Handler{ // Pod的生命周期管理，在Pod结束的时候，删除debug容器
							Exec: &v1.ExecAction{
								Command: []string{"/usr/local/bin/attach", "destroy"},
							},
						},
					},
					ReadinessProbe: &v1.Probe{ // 添加ReadinessProbe，避免服务没有启动导致终端连接失败
						Handler: v1.Handler{
							TCPSocket: &v1.TCPSocketAction{
								Port: intstr.FromInt(DebugSockPort), // sock端口
							},
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "docker",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/var/run/docker.sock",
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	log.Infof("Debug Pod info: name:%s, namespace:%s, image:%s\n", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace, pod.Spec.Containers[0].Image)
	return pod, nil
}

// 创建Kubectl
func (r *PodResource) Kubectl() error {
	// 创建Service Account
	serviceAccountSpec := getKingfisherServiceAccountSpec()
	if _, err := r.Params.ClientSet.CoreV1().ServiceAccounts(common.KubectlNamespace).Create(serviceAccountSpec); err != nil {
		if err.Error() != ServiceAccountExistsError {
			log.Errorf("ServiceAccount create error:%s; Json:%+v;", err, serviceAccountSpec)
			return err
		}
	}
	// 创建集群角色绑定
	clusterRoleBindingSpec := getKingfisherClusterRoleBindingSpec()
	if _, err := r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Create(clusterRoleBindingSpec); err != nil {
		if err.Error() != ClusterRoleBindingExistsError {
			log.Errorf("ClusterRoleBinding create error:%s; Json:%+v;", err, clusterRoleBindingSpec)
			return err
		}
	}
	//// 创建Pod
	//podSpec := r.getKubectlPodSpec()
	//pod, err := r.Params.ClientSet.CoreV1().Pods(common.KubectlNamespace).Create(r.getKubectlPodSpec())
	//if err != nil {
	//	log.Errorf("Pod create error:%s; Json:%+v; Name:%s", err, podSpec, r.Params.Name)
	//	return err
	//}
	// 创建Deployment
	deploymentSpec := r.getKubectlDeploymentSpec()
	deployment, err := r.Params.ClientSet.AppsV1().Deployments(common.KubectlNamespace).Create(deploymentSpec)
	if err != nil {
		log.Errorf("Deployment create error:%s; Json:%+v; Name:%s", err, deploymentSpec, r.Params.Name)
		return err
	}
	watcher, err := r.Params.ClientSet.AppsV1().Deployments(common.KubectlNamespace).Watch(metav1.SingleObject(deployment.ObjectMeta))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// 读取watch items，直到达到满足的条件
	_, err = watch.UntilWithoutRetry(ctx, watcher, deploymentRunning)
	if err != nil {
		return err
	}
	//watcher, err := r.Params.ClientSet.CoreV1().Pods(common.KubectlNamespace).Watch(metav1.SingleObject(pod.ObjectMeta))
	//if err != nil {
	//	return err
	//}
	//ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	//defer cancel()
	//// 读取watch items，直到达到满足的条件
	//_, err = watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	//if err != nil {
	//	return err
	//}
	// 写入数据库
	pluginList := common.ClusterPluginDB{}
	query := map[string]interface{}{
		"$.plugin":  r.Plugin,
		"$.cluster": r.Params.Cluster,
	}
	if err = db.Get(common.ClusterPluginTable, query, &pluginList); err == nil {
		pluginList.Status = 1
		pluginList.Timestamp = time.Now().Unix()
		if err = db.Update(common.ClusterPluginTable, pluginList.Id, pluginList); err != nil {
			return err
		}
	} else {
		clusterPlugin := common.ClusterPluginDB{
			Id:        kit.UUID("p"),
			Plugin:    r.Plugin,
			Cluster:   r.Params.Cluster,
			Status:    1,
			Timestamp: time.Now().Unix(),
		}
		if err = db.Insert(common.ClusterPluginTable, clusterPlugin); err != nil {
			log.Errorf("Cluster Plugin add error:%s; Json:%+v;", err, r.PostData)
			return err
		}
	}
	return nil
}

// 删除Kubectl
func (r *PodResource) UnKubectl() error {
	// 删除Service Account
	if err := r.Params.ClientSet.CoreV1().ServiceAccounts(common.KubectlNamespace).Delete(common.ServiceAccountName, &metav1.DeleteOptions{}); err != nil {
		log.Errorf("ServiceAccount Delete error:%s;", err)
	}
	// 删除集群角色绑定
	if err := r.Params.ClientSet.RbacV1beta1().ClusterRoleBindings().Delete(common.ClusterRoleBindingName, &metav1.DeleteOptions{}); err != nil {
		log.Errorf("ClusterRoleBinding delete error:%s；", err)
	}
	//// 删除Pod
	//err := r.Params.ClientSet.CoreV1().Pods(common.KubectlNamespace).Delete(common.KubectlPodName, &metav1.DeleteOptions{})
	//if err != nil {
	//	log.Errorf("Pod delete error:%s;", err)
	//	return err
	//}
	// 删除Deployment
	item, err := r.Params.ClientSet.AppsV1().Deployments(common.KubectlNamespace).Get(common.KubectlDeploymentName, metav1.GetOptions{})
	if err != nil {
		errInfo := fmt.Sprintf("deployments.apps \"%s\" not found", common.KubectlDeploymentName)
		if err.Error() == errInfo {
			log.Errorf("deployment delete error:%s；", err)
		} else {
			log.Errorf("deployment delete error:%s；", err)
			return err
		}
	} else {
		r.Params.Uid = string(item.UID)
		d := ControllerResource{
			Params: r.Params,
		}
		d.Params.Name = common.KubectlDeploymentName
		d.Params.Namespace = common.KubectlNamespace
		err := d.DelReplicaSetForController()
		if err != nil {
			log.Errorf("ReplicaSet delete error:%s;", err)
			return err
		}
	}
	pluginList := common.ClusterPluginDB{}
	query := map[string]interface{}{
		"$.plugin":  r.Plugin,
		"$.cluster": r.Params.Cluster,
	}
	if err = db.Get(common.ClusterPluginTable, query, &pluginList); err == nil {
		pluginList.Status = 0
		pluginList.Timestamp = time.Now().Unix()
		if err = db.Update(common.ClusterPluginTable, pluginList.Id, pluginList); err != nil {
			return err
		}
	} else {
		log.Errorf("Cluster plugin modify status error:%s;", err)
		return err
	}
	return nil
}

// kubectl deployment创建的描述
func (r *PodResource) getKubectlDeploymentSpec() *appv1.Deployment {
	var replicas *int32
	var tmp int32 = 1
	replicas = &tmp

	deployment := &appv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      common.KubectlDeploymentName,
			Namespace: common.KubectlNamespace,
		},
		Spec: appv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": common.KubectlDeploymentName},
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      map[string]string{"app": common.KubectlDeploymentName},
					Annotations: DisableIsitoInject, // 不开启istio注入
				},
				Spec: v1.PodSpec{
					ServiceAccountName: "kingfisher",
					Containers: []v1.Container{
						{
							Name:            common.KubectlDeploymentName,
							Image:           r.Image,       // 此镜像由king-kubectl项目生成
							ImagePullPolicy: v1.PullAlways, // 避免因为latest导致镜像不更新
							Env: []v1.EnvVar{
								{Name: "KUBECTL_VERSION", Value: r.KubectlVersion},
							},
							Lifecycle: &v1.Lifecycle{ // Pod的生命周期管理
								PostStart: &v1.Handler{ // 在容器创建之后，容器的Entrypoint执行之前，执行king-kubectl.sh从网络中下载对应版本的kubectl
									Exec: &v1.ExecAction{
										Command: []string{"/bin/sh", "/opt/king-kubectl.sh"},
									},
								},
							},
						},
					},
					RestartPolicy: v1.RestartPolicyAlways,
				},
			},
		},
	}
	return deployment
}

// 已经就绪的副本数等于设置的副本数认为成功运行
func deploymentRunning(event watchType.Event) (bool, error) {
	switch event.Type {
	case watchType.Deleted:
		return false, nil
	}
	switch t := event.Object.(type) {
	case *appv1.Deployment:
		if t.Status.ReadyReplicas == *t.Spec.Replicas {
			return true, nil
		}
	}
	return false, nil
}

// kubectl Pod创建的描述
func (r *PodResource) getKubectlPodSpec() *v1.Pod {
	pod := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "king-kubectl-pod",
			Namespace: common.KubectlNamespace,
		},
		Spec: v1.PodSpec{
			ServiceAccountName: "kingfisher",
			Containers: []v1.Container{
				{
					Name:            "king-kubectl",
					Image:           r.Image, // 此镜像由king-debug项目生成
					ImagePullPolicy: v1.PullAlways,
					Env: []v1.EnvVar{
						{Name: "KUBECTL_VERSION", Value: r.KubectlVersion},
					},
					Lifecycle: &v1.Lifecycle{ // Pod的生命周期管理
						PostStart: &v1.Handler{ // 在容器创建之后，容器的Entrypoint执行之前，执行debug拉取镜像创建容器
							Exec: &v1.ExecAction{
								Command: []string{"/bin/sh", "/opt/king-kubectl.sh"},
							},
						},
					},
				},
			},
			RestartPolicy: v1.RestartPolicyNever,
		},
	}
	log.Infof("Kubectl Pod info: [Name:%s, Namespace:%s, Image:%s]\n", pod.ObjectMeta.Name, pod.ObjectMeta.Namespace, pod.Spec.Containers[0].Image)
	return pod
}

func (r *PodResource) GenerateCreateData(c *gin.Context) (err error) {
	switch r.Params.DataType {
	case "yaml":
		var j []byte
		create := common.PostType{}
		if err = c.BindJSON(&create); err != nil {
			return
		}
		if j, _, err = kit.YamlToJson(create.Context); err != nil {
			return
		}
		if err = json.Unmarshal(j, &r.PostData); err != nil {
			return
		}
	case "json":
		if err = c.BindJSON(&r.PostData); err != nil {
			return
		}
	default:
		return errors.New(common.ContentTypeError)
	}
	return nil
}

// 创建救援Pod
func (r *PodResource) Rescue() error {
	podInfo, err := r.Get()
	if err != nil {
		log.Errorf("get podInfo error: %s", err)
		return err
	}
	podSpec := r.getRescuePodSpec(podInfo)
	pod, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Create(podSpec)
	if err != nil {
		log.Errorf("Rescue pod create error:%s; Json:%+v; Name:%s", err, podSpec, r.Params.Name)
		return err
	}
	watcher, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).Watch(metav1.SingleObject(pod.ObjectMeta))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	// 读取watch items，直到达到满足的条件
	_, err = watch.UntilWithoutRetry(ctx, watcher, conditions.PodRunning)
	if err != nil {
		return err
	}
	return nil
}

// 救援Pod创建的描述
func (r *PodResource) getRescuePodSpec(podInfo *v1.Pod) *v1.Pod {
	t := false
	// labels 和 annotations 都需要清除
	podInfo.ObjectMeta = metav1.ObjectMeta{
		Name:      "rescue-pod-" + r.Params.Name, //修改Pod名
		Namespace: r.Params.Namespace,
		OwnerReferences: []metav1.OwnerReference{ // 修改Pod OwnerReferences 在原有Pod清除后此Pod也将被回收
			{
				APIVersion:         "v1",
				Kind:               "Pod",
				Name:               podInfo.Name,
				UID:                podInfo.UID,
				Controller:         &t, // 如果是个控制器
				BlockOwnerDeletion: &t, // 不阻塞Owner的删除
			},
		},
	}
	// 清除istio注入
	if rescueCondition(r.RescueCondition, RescueConditionIstioInject) {
		podInfo.ObjectMeta.Annotations = DisableIsitoInject // 不开启istio注入
		rescueInitContainer := make([]v1.Container, 0)
		// 在initContainer中清除istio的注入container
		for _, initContainer := range podInfo.Spec.InitContainers {
			if initContainer.Name != "istio-init" {
				rescueInitContainer = append(rescueInitContainer, initContainer)
			}
		}
		podInfo.Spec.InitContainers = rescueInitContainer
		// 在volume中清除isito相关volume
		volumes := make([]v1.Volume, 0)
		for _, volume := range podInfo.Spec.Volumes {
			if !strings.Contains(volume.Name, "istio-") {
				volumes = append(volumes, volume)
			}
		}
		podInfo.Spec.Volumes = volumes

	}
	// 清除挂载卷的影响
	if rescueCondition(r.RescueCondition, RescueConditionVolume) {
		podInfo.Spec.Volumes = []v1.Volume{}
	}
	// 清除initContainers
	if rescueCondition(r.RescueCondition, RescueConditionInitContainers) {
		podInfo.Spec.InitContainers = []v1.Container{}
	}
	// 亲和性
	if rescueCondition(r.RescueCondition, RescueConditionAffinity) {
		podInfo.Spec.Affinity = &v1.Affinity{}
	}
	// 容忍性
	if rescueCondition(r.RescueCondition, RescueConditionToleration) {
		podInfo.Spec.Tolerations = []v1.Toleration{}
	}
	rescueContainer := v1.Container{}
	for _, container := range podInfo.Spec.Containers {
		if container.Name == r.Container {
			rescueContainer = container
			rescueContainer.Name = "rescue-" + rescueContainer.Name
			// 清除挂载卷
			if rescueCondition(r.RescueCondition, RescueConditionVolume) {
				rescueContainer.VolumeMounts = []v1.VolumeMount{}
				rescueContainer.VolumeDevices = []v1.VolumeDevice{}
			}
			// 清除环境变量
			if rescueCondition(r.RescueCondition, RescueConditionEnv) {
				rescueContainer.Env = []v1.EnvVar{}
				rescueContainer.EnvFrom = []v1.EnvFromSource{}
			}
			// 替换command
			rescueContainer.Command = []string{"sleep"}
			// 替换args
			rescueContainer.Args = []string{"24h"}
			// 清除Port
			rescueContainer.Ports = nil
			// 清除就绪探针
			rescueContainer.ReadinessProbe = nil
			// 清除存活探针
			rescueContainer.LivenessProbe = nil
			// 清除安全上下文
			rescueContainer.SecurityContext = &v1.SecurityContext{}
			// 特权模式
			if rescueCondition(r.RescueCondition, RescueConditionPrivileged) {
				privileged := true
				rescueContainer.SecurityContext.Privileged = &privileged
			}
		}
	}
	// 指定容器
	podInfo.Spec.Containers = []v1.Container{rescueContainer}
	// 救援Pod不重启
	podInfo.Spec.RestartPolicy = v1.RestartPolicyNever
	// 清除状态信息
	podInfo.Status = v1.PodStatus{}
	log.Infof("Rescue Pod info: %+v", podInfo)
	return podInfo
}

func (r *PodResource) GetDebugPodIPByPod() (string, error) {
	pods, err := r.List()
	if err != nil {
		return "", err
	}
	for _, pod := range pods.Items {
		if DebugPodPrefix+r.Params.Name == pod.Name {
			return pod.Status.PodIP, nil
		}
	}
	return "", nil
}

func rescueCondition(conditions []string, condition string) bool {
	for _, c := range conditions {
		if condition == c {
			return true
		}
	}
	return false
}
