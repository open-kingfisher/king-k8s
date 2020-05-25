#!/bin/sh

#!/bin/sh
[ "$DB_URL" ] || DB_URL='user:password@tcp(192.168.10.100:3306)/kingfisher'
[ "$LISTEN" ] || LISTEN=0.0.0.0
[ "$PORT" ] || LISTEN=8080
[ "$RPCPORT" ] || LISTEN=50000
[ "$RABBITMQ_URL" ] || RABBITMQ_URL='amqp://user:password@king-rabbitmq:5672/'

mkdir /var/log/kingfisher

# grpc
/usr/local/bin/king-k8s-grpc -dbURL=$DB_URL -listen=$LISTEN:$PORT -listen=$LISTEN:$RPCPORT &

# k8s
/usr/local/bin/king-k8s -dbURL=$DB_URL  -listen=$LISTEN:$PORT -rabbitMQURL=$RABBITMQ_URL

