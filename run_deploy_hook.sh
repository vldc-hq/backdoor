#!/bin/bash -e
echo "starting deploy hook" | systemd-cat -p info
cd /home/ci
sudo -u ci ./ci
