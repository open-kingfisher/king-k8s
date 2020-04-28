package resource

import (
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/open-kingfisher/king-utils/common/handle"
)

type EventResource struct {
	Params   *handle.Resources
	PostData *v1.Event
}

func (r *EventResource) List() (*v1.EventList, error) {
	events := &v1.EventList{}
	if eventList, err := r.Params.ClientSet.CoreV1().Events(r.Params.Namespace).List(metav1.ListOptions{}); err == nil {
		// 通过uid获取对应资源的Event
		if r.Params.Uid != "" {
			for _, event := range eventList.Items {
				if string(event.InvolvedObject.UID) == r.Params.Uid {
					events.Items = append(events.Items, event)
				}
			}
			return events, nil
		} else {
			return eventList, nil
		}
	} else {
		return nil, err
	}
}
