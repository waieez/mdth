#!/bin/bash

export RABBIT="amqp://guest:guest@localhost:5672/"

docker stop rabbit
docker rm -f rabbit
docker run -d --hostname rabbit --name rabbit -p 15672:15672 -p 5672:5672 rabbitmq:3-management