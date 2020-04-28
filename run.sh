#!/bin/bash

# grpc
/usr/local/bin/king-k8s-grpc -dbURL='user:password@tcp(192.168.10.100:3306)/kingfisher' -listen=0.0.0.0:8080 -listen=0.0.0.0:50000 &

# k8s
/usr/local/bin/king-k8s -dbURL='user:password@tcp(192.168.10.100:3306)/kingfisher' -listen=0.0.0.0:8080 -rabbitMQURL='amqp://user:password@king-rabbitmq:5672/'

