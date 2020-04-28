FROM alpine:3.10

ENV TIME_ZONE Asia/Shanghai

ADD run.sh /usr/local/bin
ADD bin/king-k8s /usr/local/bin
ADD bin/king-k8s-grpc /usr/local/bin

RUN set -xe \
    && sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories \
    && apk --no-cache add --update ca-certificates && apk add --no-cache tzdata \
    && echo "${TIME_ZONE}" > /etc/timezone \
    && ln -sf /usr/share/zoneinfo/${TIME_ZONE} /etc/localtime \
    && mkdir /var/log/kingfisher \
    && mkdir /lib64 \
    && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2 \
    && chmod +x /usr/local/bin/run.sh


CMD ["sh", "/usr/local/bin/run.sh"]


EXPOSE 8080 50000