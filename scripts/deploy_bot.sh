#!/bin/sh
docker-compose -p devbot -f docker-compose.dev.yml pull
docker-compose -p devbot -f docker-compose.dev.yml down
docker-compose -p devbot -f docker-compose.dev.yml up -d
