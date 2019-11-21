#!/bin/sh
docker-compose -p prodbot -f docker-compose.prod.yml pull
docker-compose -p prodbot -f docker-compose.prod.yml down
docker-compose -p prodbot -f docker-compose.prod.yml up -d
