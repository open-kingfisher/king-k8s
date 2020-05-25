FROM golang:1.14.3-alpine3.11 as builder
ARG NAME="king-k8s"
ARG GIT_URL="https://github.com/open-kingfisher/$NAME.git"
RUN set -xe \
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk add --no-cache protoc git make \
    && export GO111MODULE=on  && go get github.com/golang/protobuf/protoc-gen-go@v1.3 \
    && git clone $GIT_URL /$NAME && cd /$NAME && make generate && make build

FROM alpine:3.10

ARG NAME="king-k8s"
ENV TIME_ZONE Asia/Shanghai
RUN set -xe \
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --no-cache add --update ca-certificates && apk add --no-cache tzdata \
    && echo "${TIME_ZONE}" > /etc/timezone \
    && ln -sf /usr/share/zoneinfo/${TIME_ZONE} /etc/localtime \
    && mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 

COPY --from=builder /$NAME/entrypoint.sh /entrypoint.sh
COPY --from=builder /$NAME/bin/$NAME /usr/local/bin
COPY --from=builder /$NAME/bin/$NAME-grpc /usr/local/bin

ENTRYPOINT ["/bin/sh","/entrypoint.sh"]
