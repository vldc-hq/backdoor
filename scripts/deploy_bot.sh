#!/bin/sh
sleep 1800 # wait for dockerhub to build new image
docker-compose down -rmi bot
docker-compose up -d bot
