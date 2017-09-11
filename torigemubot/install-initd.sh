#!/bin/bash
service torigemubot stop
cp torigemubot-initd /etc/init.d/torigemubot
insserv torigemubot
service torigemubot start
