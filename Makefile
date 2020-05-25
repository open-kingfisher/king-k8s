GOCMD=GOOS=linux go
PROJECT=kingfisher
SERVICE=king-k8s
GRPC=king-k8s-grpc
REGISTRY=registry.kingfisher.com.cn
REVISION=latest

build:
	go build -o bin/$(SERVICE) main.go
	go build -o bin/$(GRPC) grpc/server/main.go

push:
	docker build -f Dockerfile -t $(REGISTRY)/$(PROJECT)/$(SERVICE):$(REVISION) .
	docker push $(REGISTRY)/$(PROJECT)/$(SERVICE):$(REVISION)

generate:
	protoc -I./grpc/proto ./grpc/proto/*.proto --go_out=plugins=grpc:./grpc/proto/
