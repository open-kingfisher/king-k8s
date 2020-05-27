#!/bin/sh

#!/bin/sh
[ "$DB_URL" ] || DB_URL='user:password@tcp(192.168.10.100:3306)/kingfisher'
[ "$LISTEN" ] || LISTEN=0.0.0.0
[ "$PORT" ] || PORT=8080
[ "$RPCPORT" ] || RPCPORT=50000
[ "$RABBITMQ_URL" ] || RABBITMQ_URL='amqp://user:password@king-rabbitmq:5672/'
[ "$TIME_ZONE" ] || TIME_ZONE="Asia/Shanghai"
[ "$ALPINE_REPO" ] || ALPINE_REPO="mirrors.aliyun.com"

sed -i 's/dl-cdn.alpinelinux.org/${ALPINE_REPO}/g' /etc/apk/repositories     
apk --no-cache add tzdata 
echo "${TIME_ZONE}" > /etc/timezone 
ln -sf /usr/share/zoneinfo/${TIME_ZONE} /etc/localtime 
mkdir /lib64 
ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

mkdir -p /var/log/kingfisher

# grpc
/usr/local/bin/king-k8s-grpc -dbURL=$DB_URL -listen=$LISTEN:$PORT -listen=$LISTEN:$RPCPORT &

# k8s
/usr/local/bin/king-k8s -dbURL=$DB_URL  -listen=$LISTEN:$PORT -rabbitMQURL=$RABBITMQ_URL

