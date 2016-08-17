#!/bin/bash

# TODO: make this better

docker stop rabbit
docker rm -f rabbit
docker stop redis
docker rm -f redis

docker run -d --hostname rabbit --name rabbit -p 15672:15672 -p 5672:5672 rabbitmq:3-management
docker run -d --name redis -p 6379:6379 redis