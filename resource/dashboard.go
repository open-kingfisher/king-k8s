package resource

import (
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
	"sync"
)

type InfoCard struct {
	Title string `json:"title"`
	Count int    `json:"count"`
}

type Container struct {
	Name    string `json:"name"`
	Cause   string `json:"cause"`
	Message string `json:"message"`
}

type Reason struct {
	Name       string      `json:"name"`
	Node       string      `json:"node"`
	Containers []Container `json:"container"`
}

type Application struct {
	Name                string   `json:"name"`
	Replicas            *int32   `json:"replicas"`
	AvailableReplicas   int32    `json:"availableReplicas"`
	UnAvailableReplicas int32    `json:"unavailableReplicas"`
	CreationTimestamp   string   `json:"creationTimestamp"`
	Status              string   `json:"status"`
	LastTransitionTime  string   `json:"lastTransitionTime"`
	Reasons             []Reason `json:"reason"`
}
type DashboardResource struct {
	Params *handle.Resources
}

// sync.WaitGroup方法
func (r *DashboardResource) ListInfoCard() ([]*InfoCard, error) {
	wg := sync.WaitGroup{}
	infoCardList := make([]*InfoCard, 0)
	for i := 1; i < 5; i++ {
		wg.Add(1)
	}
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				log.DPanicf("ListInfoCard Ingress panic:%v", err)
			}
		}()
		ingress := IngressResource{
			Params: &handle.Resources{
				Namespace: r.Params.Namespace,
				ClientSet: r.Params.ClientSet,
			},
		}
		if data, err := ingress.List(); err == nil {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Ingress",
				Count: len(data.Items),
			})
		} else {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Ingress",
			})
		}
	}()
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				log.DPanicf("ListInfoCard Node panic:%v", err)
			}
		}()
		node := NodeResource{
			Params: &handle.Resources{
				Namespace: r.Params.Namespace,
				ClientSet: r.Params.ClientSet,
			},
		}
		if data, err := node.List(); err == nil {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Node",
				Count: len(data.Items),
			})
		} else {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Node",
			})
		}
	}()
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				log.DPanicf("ListInfoCard Service panic:%v", err)
			}
		}()
		service := ServiceResource{
			Params: &handle.Resources{
				Namespace: r.Params.Namespace,
				ClientSet: r.Params.ClientSet,
			},
		}
		if data, err := service.List(); err == nil {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Service",
				Count: len(data.Items),
			})
		} else {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Service",
			})
		}
	}()
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				log.DPanicf("ListInfoCard Deployment panic:%v", err)
			}
		}()
		if data, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Deployment",
				Count: len(data.Items),
			})
		} else {
			infoCardList = append(infoCardList, &InfoCard{
				Title: "Deployments",
			})
		}
	}()
	wg.Wait()
	return infoCardList, nil
}

func (r *DashboardResource) ListApplication() ([]*Application, error) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Application list panic: %s", err)
		}
	}()
	wg := sync.WaitGroup{}
	ApplicationList := make([]*Application, 0)
	for i := 1; i < 2; i++ {
		wg.Add(1)
	}
	go func() {
		defer func() {
			wg.Done()
			if err := recover(); err != nil {
				log.DPanicf("ListApplication panic:%v", err)
			}
		}()
		if data, err := r.Params.ClientSet.AppsV1().Deployments(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
			for _, v := range data.Items {
				selector := "metadata.namespace=" + r.Params.Namespace + ",status.phase!=Running,spec.restartPolicy=Always"
				status := string(v.Status.Conditions[0].Status)
				reason := make([]Reason, 0)
				if status != "True" {
					pod, _ := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).List(metav1.ListOptions{FieldSelector: selector})
					for _, p := range pod.Items {
						d := strings.Split(p.Name, "-")[:2]
						if strings.Join(d, "-") == v.Name {
							container := make([]Container, 0)
							for _, s := range p.Status.ContainerStatuses {
								container = append(container, Container{
									Name:    s.Name,
									Cause:   s.State.Waiting.Reason,
									Message: s.State.Waiting.Message,
								})
							}
							reason = append(reason, Reason{
								Name:       p.Name,
								Containers: container,
								Node:       p.Spec.NodeName,
							})
						}
					}
				}
				ApplicationList = append(ApplicationList, &Application{
					Name:                v.Name,
					Replicas:            v.Spec.Replicas,
					AvailableReplicas:   v.Status.AvailableReplicas,
					UnAvailableReplicas: v.Status.UnavailableReplicas,
					CreationTimestamp:   v.CreationTimestamp.Format("2006-01-02 15:04:05"),
					LastTransitionTime:  v.Status.Conditions[0].LastTransitionTime.Format("2006-01-02 15:04:05"),
					Status:              status,
					Reasons:             reason,
				})
			}
		} else {
			log.Errorf("Application list error:%s", err)
		}
	}()
	wg.Wait()
	return ApplicationList, nil
}

func (r *DashboardResource) ListHistory() ([]*common.AuditLog, error) {
	audit := make([]*common.AuditLog, 0)
	if err := db.List(common.DataField, common.AuditLogTable, &audit, "order by data -> '$.action_time' desc limit 7"); err == nil {
		return audit, err
	} else {
		return audit, nil
	}
}

func (r *DashboardResource) ListPodStatus() ([]map[string]interface{}, error) {
	pending, running, succeeded, failed, unknown := 0, 0, 0, 0, 0
	if pods, err := r.Params.ClientSet.CoreV1().Pods(r.Params.Namespace).List(metav1.ListOptions{}); err != nil {
		return nil, err
	} else {
		for _, pod := range pods.Items {
			switch pod.Status.Phase {
			case corev1.PodPending:
				pending += 1
			case corev1.PodRunning:
				running += 1
			case corev1.PodSucceeded:
				succeeded += 1
			case corev1.PodFailed:
				failed += 1
			case corev1.PodUnknown:
				unknown += 1
			}
		}
	}
	podStatus := []map[string]interface{}{
		{"value": pending, "name": corev1.PodPending},
		{"value": running, "name": corev1.PodRunning},
		{"value": succeeded, "name": corev1.PodSucceeded},
		{"value": failed, "name": corev1.PodFailed},
		{"value": unknown, "name": corev1.PodUnknown},
	}
	return podStatus, nil
}
