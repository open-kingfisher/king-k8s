GOCMD=GOOS=linux go
PROJECT=kingfisher
SERVICE=king-k8s
REGISTRY=registry.kingfisher.com.cn
REVISION=latest

build:
	go build -o bin/$(SERVICE) main.go
	go build -o bin/$(GRPC) grpc/server/main.go

push:
	docker build -f Dockerfile -t $(REGISTRY)/$(PROJECT)/$(SERVICE):$(REVISION) .
	docker push $(REGISTRY)/$(PROJECT)/$(SERVICE):$(REVISION)

generate:
	protoc -I$(GOPATH)/src/github.com/open-kingfisher/king-k8s/grpc/proto $(GOPATH)/src/github.com/open-kingfisher/king-k8s/grpc/proto/*.proto --go_out=plugins=grpc:$(GOPATH)/src/github.com/open-kingfisher/king-k8s/grpc/proto/
