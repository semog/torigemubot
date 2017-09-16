#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/torigemubot
APPFOLDER=/usr/local/lib/torigemubot
SYSTEMDFOLDER=/etc/systemd/system

mkdir -p $DATAFOLDER/
mkdir -p $APPFOLDER/

systemctl --now disable torigemubot.service

cp torigemubot.service $SYSTEMDFOLDER/
cp torigemubot $APPFOLDER/

if [ ! -e "$APPFOLDER/torigemubotsrv.sh" ]; then
	cp torigemubotsrv.sh $APPFOLDER/
fi

systemctl --force enable torigemubot.service
systemctl start torigemubot.service
