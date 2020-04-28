/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pb "github.com/open-kingfisher/king-k8s/grpc/proto"
	"github.com/open-kingfisher/king-utils/common/log"
	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
)

const (
	address     = "localhost:50000"
	defaultName = "world"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.Get(ctx, &pb.ServiceRequest{Cluster: "c_6812529878I", Namespace: "istio-test", Name: "reviews"})
	if err != nil {
		log.Fatalf("could not grpc request: %v", err)
	}
	service := v1.Service{}
	json.Unmarshal(r.Data, &service)
	fmt.Println(service.Spec.Selector)
}
