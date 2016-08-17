#!/bin/bash

# TODO put config/secrets into a secret store like vault and/or use consul for config / service discovery
export RABBIT="amqp://guest:guest@localhost:5672/"
export REDIS="localhost:6379"