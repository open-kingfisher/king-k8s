package resource

import (
	"context"
	"fmt"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"regexp"
	"time"
)

type PrometheusResource struct {
	Params    *handle.Resources
	PostData  *common.ClusterPluginDB
	ClientSet v1.API
}

func (r *PrometheusResource) Status() (interface{}, error) {
	pods, err := r.Params.ClientSet.CoreV1().Pods("default").List(metav1.ListOptions{})
	if err == nil {
		for _, p := range pods.Items {
			if p.Name == common.KubectlPodName {
				return map[string]int{"Status": 1}, nil
			}
		}
	}
	return map[string]int{"Status": 0}, err
}

func (r *PrometheusResource) NodeMetric() (interface{}, error) {
	r.getCPU()

	return map[string]interface{}{
		"CPU":    r.getCPU(),
		"Memory": r.getMem(),
	}, nil
}

func (r *PrometheusResource) getCPU() interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	allocatable := fmt.Sprintf("sum(kube_node_status_allocatable_cpu_cores{node='%s'})", r.Params.Name) // 查询节点可分配的CPU内核
	requests := fmt.Sprintf("sum(kube_pod_container_resource_requests_cpu_cores{node='%s'})", r.Params.Name)
	query := fmt.Sprintf("%s/%s", requests, allocatable)
	result, warnings, err := r.ClientSet.Query(ctx, query, time.Now())
	if err != nil {
		log.Errorf("Error querying Prometheus: %v", err)
	}
	if len(warnings) > 0 {
		log.Errorf("Warnings: %v", warnings)
	}
	res := result.String()
	// 查找对应的值
	re, _ := regexp.Compile("=> (.*?) @")
	s := re.FindSubmatch([]byte(res))

	return string(s[1]) // s[0]全部内容，是s[1]括号中的内容
}

func (r *PrometheusResource) getMem() interface{} {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	allocatable := fmt.Sprintf("sum(kube_node_status_allocatable_memory_bytes{node='%s'})", r.Params.Name) // 查询节点可分配的CPU内核
	requests := fmt.Sprintf("sum(kube_pod_container_resource_requests_memory_bytes{node='%s'})", r.Params.Name)
	query := fmt.Sprintf("%s/%s", requests, allocatable)
	result, warnings, err := r.ClientSet.Query(ctx, query, time.Now())
	if err != nil {
		log.Errorf("Error querying Prometheus: %v", err)
	}
	if len(warnings) > 0 {
		log.Errorf("Warnings: %v", warnings)
	}
	res := result.String()
	// 查找对应的值
	re, _ := regexp.Compile("=> (.*?) @")
	s := re.FindSubmatch([]byte(res))

	return string(s[1]) // s[0]全部内容，是s[1]括号中的内容
}
