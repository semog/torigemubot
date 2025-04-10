#!/bin/bash
DATAFOLDER=/usr/local/share/appdata/torigemubot
APPFOLDER=/usr/local/lib/torigemubot
SYSTEMDFOLDER=/etc/systemd/system

if [ ! -e "torigemubot" ]; then
	echo "torigemubot binary not found. Please build it first."
	exit 1
fi

# If the service is being installed for the first time, then a bot
# token must be provided. If the service is being reinstalled, then
# the token is optional.
if [ -z "$1" ] && [ ! -e "$APPFOLDER/torigemubotsrv.sh" ]; then
	echo "Usage: $0 <bot token>"
	exit 1
fi

mkdir -p $DATAFOLDER/
mkdir -p $APPFOLDER/

systemctl --now disable torigemubot.service

cp torigemubot.service $SYSTEMDFOLDER/
cp torigemubot $APPFOLDER/

# Only copy the game database if it doesn't exist
if [ ! -e "$DATAFOLDER/torigemu.db" ]; then
	if [ ! -e "../mkkanjidb/torigemu.db" ]; then
		echo You must build the torigemu.db file using the mkkanjidb tool first.
		exit 1
	fi
	cp ../mkkanjidb/torigemu.db $DATAFOLDER/
fi

# Only copy the torigemubotsrv.sh if it doesn't exist
if [ ! -e "$APPFOLDER/torigemubotsrv.sh" ]; then
	cp torigemubotsrv.sh $APPFOLDER/
fi

if [ ! -z "$1" ]; then
	# Replace/update the bot token
	sed -i "s/BOTTOKEN=.*/BOTTOKEN=$1/g" $APPFOLDER/torigemubotsrv.sh
fi

systemctl --force enable torigemubot.service
systemctl start torigemubot.service
