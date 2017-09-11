#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/torigemubot/
APPFOLDER=/usr/local/lib/torigemubot/
SYSTEMDFOLDER=/etc/systemd/system/

if [ ! -d "$DATAFOLDER" ]; then
	mkdir -p $DATAFOLDER
fi

if [ ! -d "$APPFOLDER" ]; then
	mkdir -p $APPFOLDER
fi

systemctl --now disable torigemubot.service

cp torigemubot.service $SYSTEMDFOLDER
cp torigemubot $APPFOLDER

if [ ! -e "$DATAFOLDER" ]; then
	cp torigemubotsrv.sh $APPFOLDER
fi

systemctl --force enable torigemubot.service
