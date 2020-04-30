package resource

import (
	"errors"
	"github.com/open-kingfisher/king-utils/common/handle"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/metrics/pkg/apis/custom_metrics/v1beta2"
	"k8s.io/metrics/pkg/apis/metrics/v1beta1"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	customMetrics "k8s.io/metrics/pkg/client/custom_metrics"
)

type MetricResource struct {
	Params              *handle.Resources
	MetricsClient       *metrics.Clientset
	CustomMetricsClient customMetrics.CustomMetricsClient
}

func (r *MetricResource) GetNodeMetrics() (*v1beta1.NodeMetrics, error) {
	return r.MetricsClient.MetricsV1beta1().NodeMetricses().Get(r.Params.Name, metav1.GetOptions{})
}

func (r *MetricResource) ListNodeMetrics() (*v1beta1.NodeMetricsList, error) {
	return r.MetricsClient.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
}

func (r *MetricResource) GetPodMetrics() (*v1beta1.PodMetrics, error) {
	return r.MetricsClient.MetricsV1beta1().PodMetricses(r.Params.Namespace).Get(r.Params.Name, metav1.GetOptions{})
}

func (r *MetricResource) ListPodMetrics() (*v1beta1.PodMetricsList, error) {
	return r.MetricsClient.MetricsV1beta1().PodMetricses(r.Params.Namespace).List(metav1.ListOptions{})
}

func (r *MetricResource) GetCustomMetrics() (*v1beta2.MetricValue, error) {
	return r.CustomMetricsClient.NamespacedMetrics(r.Params.Namespace).GetForObject(schema.GroupKind{Group: "", Kind: r.Params.MetricsKind}, r.Params.Name, r.Params.MetricsName, labels.NewSelector())
}

func (r *MetricResource) ListCustomMetrics() (*v1beta2.MetricValueList, error) {
	if filter, err := labels.NewRequirement(r.Params.LabelKey, selection.Equals, []string{r.Params.Name}); err != nil {
		return nil, errors.New("couldn't create a label filter")
	} else {
		if metric, err := r.CustomMetricsClient.NamespacedMetrics(r.Params.Namespace).GetForObjects(schema.GroupKind{Group: "", Kind: r.Params.MetricsKind}, labels.NewSelector().Add(*filter), r.Params.MetricsName, labels.NewSelector()); err != nil {
			return nil, err
		} else {
			return metric, nil
		}
	}
}
