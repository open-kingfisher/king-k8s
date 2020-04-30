package resource

import (
	"fmt"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"k8s.io/api/apps/v1"
	"testing"
)

func TestControllerResource_Watch(t *testing.T) {
	clientSet, err := access.Access("c_9941048464f")
	//patch := common.PatchData{
	//	Op:"add",
	//	Path: "/spec/paused",
	//	Value: false,
	//}
	r := ControllerResource{
		Params: &handle.Resources{
			Name:       "nginx01",
			Namespace:  "default",
			Cluster:    "c_9941048464f",
			Controller: "deployment",
			ClientSet:  clientSet,
			PatchData: &common.PatchJson{
				Patches: []common.PatchData{
					{
						Op:    "add",
						Path:  "/spec/paused",
						Value: true,
					},
				},
			},
		},
	}
	event, err := r.Watch()
	if err != nil {
		fmt.Println("error:", err)
	}
	deployment, err := r.Get()
	if err != nil {
		fmt.Println("get:", err)
	}
	var maxUnavailable int32
	var updatedReplicas int32
	var replicas int32
	de, ok := deployment.(*v1.Deployment)
	if ok {
		a
		maxUnavailable = int32(de.Spec.Strategy.RollingUpdate.MaxUnavailable.IntValue())
		updatedReplicas = de.Status.UpdatedReplicas
		replicas = *de.Spec.Replicas
	}
	if updatedReplicas == replicas {
		updatedReplicas = 0
	}
	for v := range event.ResultChan() {
		fmt.Printf("%v+\n", v.Object)
		if d, ok := v.Object.(*v1.Deployment); ok {
			re := d.Status.UpdatedReplicas
			fmt.Println(re)
			fmt.Println("updatedReplicas", updatedReplicas)
			fmt.Println("maxUnavailable", maxUnavailable)
			if re > updatedReplicas && re <= maxUnavailable+updatedReplicas {
				if _, err := r.Patch(); err != nil {
					fmt.Println("patch:", err)
				} else {
					if maxUnavailable+updatedReplicas >= replicas {
						r.Params.PatchData.Patches[0].Value = false
						r.Patch()
					}
					event.Stop()
					fmt.Println("patch:ok")
				}
			}

		} else {
			fmt.Println(ok)
		}
	}
}
