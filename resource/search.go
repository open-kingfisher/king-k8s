package resource

import (
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
	"github.com/open-kingfisher/king-utils/db"
	"strings"
)

type SearchResource struct {
	Params *handle.Resources
}

type search struct {
	Product   []string          `json:"product"`
	Cluster   string            `json:"cluster"`
	Namespace string            `json:"namespace"`
	Kind      map[string]string `json:"kind"`
	Pod       string            `json:"pod"`
	PodIP     string            `json:"podIp"`
	HostIP    string            `json:"hostIp"`
	NodeName  string            `json:"nodeName"`
}

func (r *SearchResource) Get() ([]*search, error) {
	var g errgroup.Group
	searchList := make([]*search, 0)
	clusters := make([]*common.ClusterDB, 0)
	if err := db.List(common.DataField, common.Cluster, &clusters, ""); err != nil {
		return searchList, err
	}
	for _, cluster := range clusters {
		cluster := cluster
		g.Go(func() error {
			if clientSet, err := access.Access(cluster.Id); err != nil {
				log.Errorf("Search cluster %s clientSet  error: %s", cluster.Id, err)
			} else {
				if podList, err := clientSet.CoreV1().Pods(corev1.NamespaceAll).List(metav1.ListOptions{}); err != nil {
					log.Errorf("Search list pod error: %s", err)
				} else {
					for _, v := range podList.Items {
						if strings.Contains(v.Status.PodIP, r.Params.Name) ||
							strings.Contains(v.Status.HostIP, r.Params.Name) ||
							strings.Contains(v.Spec.NodeName, r.Params.Name) ||
							strings.Contains(v.GetName(), r.Params.Name) {
							product, _ := getProduct(cluster.Id, v.GetNamespace())
							name := ""
							kind := make(map[string]string)
							if v.OwnerReferences != nil {
								if v.OwnerReferences[0].Kind == "ReplicaSet" {
									nameList := strings.Split(v.GetName(), "-")
									for _, n := range nameList[:len(nameList)-2] {
										name += n + "-"
									}
									name = strings.TrimSuffix(name, "-")
									v.OwnerReferences[0].Kind = "Deployment"
								} else {
									name = v.OwnerReferences[0].Name
								}
								kind = map[string]string{
									v.OwnerReferences[0].Kind: name,
								}
							}
							search := search{
								Product:   product,
								Cluster:   cluster.Name,
								Namespace: v.GetNamespace(),
								Kind:      kind,
								Pod:       v.GetName(),
								PodIP:     v.Status.PodIP,
								HostIP:    v.Status.HostIP,
								NodeName:  v.Spec.NodeName,
							}
							searchList = append(searchList, &search)
						}
					}
				}
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return searchList, err
	}
	return searchList, nil
}

func getProduct(cluster, namespace string) ([]string, error) {
	product := make([]string, 0)
	products := make([]*common.ProductDB, 0)
	if err := db.List(common.DataField, common.Product, &products, ""); err != nil {
		return nil, err
	}
	for _, v := range products {
		namespaceId, _ := getNamespaceId(cluster, namespace)
		if sliceContain(cluster, v.Cluster) && sliceContain(namespaceId, v.Namespace) {
			productName, _ := getProductName(v.Id)
			product = append(product, productName)
		}
	}
	return product, nil
}

func getProductName(productId string) (string, error) {
	productTree := common.ProductTree{}
	if err := db.GetById(common.ProductTreeTable, productId, &productTree); err != nil {
		return "", err
	}
	return productTree.Name, nil
}

func getNamespaceId(clusterId, namespaceName string) (string, error) {
	var namespaceId string
	namespaceList := make([]*common.NamespaceDB, 0)
	if err := db.List(common.DataField, common.Namespace, &namespaceList, "WHERE data-> '$.name'=? and data-> '$.cluster'=?", namespaceName, clusterId); err != nil {
		return namespaceId, err
	}
	if len(namespaceList) > 0 {
		namespaceId = namespaceList[0].Id
	}
	return namespaceId, nil
}

func sliceContain(s string, slice []string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
