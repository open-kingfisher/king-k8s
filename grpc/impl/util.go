package impl

import (
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/access"
	"github.com/open-kingfisher/king-utils/common/handle"
	"github.com/open-kingfisher/king-utils/common/log"
)

func GenerateCommonParams(cluster, namespace, name string) (*handle.Resources, error) {
	// 获取clientSet，如果失败直接返回错误
	clientSet, err := access.Access(cluster)
	if err != nil {
		log.Errorf("%s%s", common.K8SClientSetError, err)
		return nil, err
	}
	commonParams := &handle.Resources{
		Namespace: namespace,
		Cluster:   cluster,
		Name:      name,
		ClientSet: clientSet,
	}
	return commonParams, nil
}
