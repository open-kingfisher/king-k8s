package util

import (
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"strings"
)

// 根据Label生成对应的labelSelector
// k8s-app=nginx,app=nginx01 多标签查询
func GenerateLabelSelector(selector map[string]string) string {
	var labelSelector string
	for k, v := range selector {
		labelSelector += k + "=" + v + ","
	}
	return strings.TrimRight(labelSelector, ",")
}

func GetPodBySelectorLabel(labelSelector, nameSpace string, clientSet *kubernetes.Clientset) (*v1.PodList, error) {
	if pods, err := clientSet.CoreV1().Pods(nameSpace).List(metav1.ListOptions{LabelSelector: labelSelector}); err != nil {
		return nil, err
	} else {
		return pods, nil
	}
}

func GetDeploymentBySelectorLabel(labelSelector, nameSpace string, clientSet *kubernetes.Clientset) (*appsv1.DeploymentList, error) {
	if deployment, err := clientSet.AppsV1().Deployments(nameSpace).List(metav1.ListOptions{LabelSelector: labelSelector}); err != nil {
		return nil, err
	} else {
		return deployment, nil
	}
}
