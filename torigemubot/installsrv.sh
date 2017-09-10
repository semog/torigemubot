#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/torigemubot/
APPFOLDER=/usr/local/lib/torigemubot/
SYSTEMDFOLDER=/etc/systemd/system/

if [ ! -d "$DATAFOLDER" ]; then
	mkdir $DATAFOLDER
fi

if [ ! -d "$APPFOLDER" ]; then
	mkdir $APPFOLDER
fi

systemctl --now disable torigemubot.service

cp torigemubot.service $SYSTEMDFOLDER
cp torigemubot torigemubotsrv.sh $APPFOLDER
pushd $APPFOLDER
chmod +x torigemubotsrv.sh torigemubot
popd
systemctl --force enable torigemubot.service
