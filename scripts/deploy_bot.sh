#!/bin/sh
sleep 1800 # wait for dockerhub to build new image
sudo docker-compose down --rmi all
sudo docker-compose up -d
