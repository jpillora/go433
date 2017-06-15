#!/bin/bash
PROG=$1
cd $PROG || exit 1
echo "building ($PROG)" &&
GOOS=linux GOARCH=arm go build -o /tmp/gobin &&
echo "uploading" &&
rsync -e "ssh -p 22" --progress --compress /tmp/gobin root@jamopi:/usr/local/bin/$PROG &&
echo "done" &&
rm /tmp/gobin
