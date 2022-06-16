#!/bin/sh
docker pull ghcr.io/egregors/shoppingcatbot/scbot:latest
docker stop scatb
docker rm -f scatb
docker run -d -v /tmp/dumps:/dumps --name=scatb --env-file=scatb.env ghcr.io/egregors/shoppingcatbot/scbot