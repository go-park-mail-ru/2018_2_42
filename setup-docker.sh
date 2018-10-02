#!/bin/bash

#Возможно не сработает, попробуй вызвать команды в данной последовательносьти из консоли

# build the flask container
docker build -t auth .

# create the network
docker network create docknet

# start the ES container
docker run -d --net docknet -p 5432:5432 --name pg postgres

# start the flask app container
docker run -d --net docknet -p 8080:8080 --name auth auth:latest