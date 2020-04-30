package main

import (
	"net"

	"github.com/open-kingfisher/king-k8s/grpc/impl"
	pb "github.com/open-kingfisher/king-k8s/grpc/proto"
	"github.com/open-kingfisher/king-utils/common"
	"github.com/open-kingfisher/king-utils/common/log"
	"google.golang.org/grpc"
)

func main() {
	listen, err := net.Listen("tcp", common.GRPCPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Infof("grpc server start ok listen%s", common.GRPCPort)
	s := grpc.NewServer()
	pb.RegisterEchoServer(s, &impl.Say{})
	pb.RegisterServiceServer(s, &impl.Service{})
	pb.RegisterDeploymentServer(s, &impl.Deployment{})
	if err := s.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
