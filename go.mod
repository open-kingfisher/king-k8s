module github.com/open-kingfisher/king-k8s

go 1.14

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.6.2
	github.com/golang/protobuf v1.4.0
	github.com/open-kingfisher/king-utils v0.0.0-20200526073742-a2d1d095795b
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1
	golang.org/x/net v0.0.0-20200425230154-ff2c4b7c35a0
	golang.org/x/sync v0.0.0-20200317015054-43a5402ce75a
	google.golang.org/grpc v1.29.1
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.18.2
	k8s.io/apimachinery v0.18.2
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/metrics v0.18.2
)

replace (
	github.com/docker/docker => github.com/docker/docker v0.7.3-0.20190924004649-91870ed38213
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/metrics => k8s.io/metrics v0.17.3
)
