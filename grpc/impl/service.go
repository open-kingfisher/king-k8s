package impl

import (
	"context"
	"encoding/json"
	pb "github.com/open-kingfisher/king-k8s/grpc/proto"
	"github.com/open-kingfisher/king-k8s/resource"
	"github.com/open-kingfisher/king-utils/common/log"
)

type Service struct{}

func (s *Service) Get(ctx context.Context, in *pb.ServiceRequest) (*pb.ServiceResponse, error) {
	// 生成通用参数
	commonParams, err := GenerateCommonParams(in.Cluster, in.Namespace, in.Name)
	if err != nil {
		log.Errorf("Generate Common Params Error: %v", err)
		return nil, err
	}
	// 构建对应结构体
	r := resource.ServiceResource{Params: commonParams}
	// 调用结构体方法
	data, err := r.Get()
	if err != nil {
		log.Errorf("Get Service Error: %v", err)
		return nil, err
	}
	// 序列化
	b, err := json.Marshal(data)
	if err != nil {
		log.Errorf("Marshal Service Error: %v", err)
		return nil, err
	}
	return &pb.ServiceResponse{Data: b}, nil
}
