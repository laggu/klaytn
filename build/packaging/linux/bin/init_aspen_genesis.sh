#!/usr/bin/env bash

BIN=$(cd "$(dirname $0)"; pwd)
CMD_HOME=$(dirname $BIN)
CONF=$CMD_HOME/conf

source $CONF/klay.conf

if [ ! -d $KLAY_HOME ]; then
    mkdir -p $KLAY_HOME
fi
if [ ! -d $DATA_DIR ]; then
    mkdir -p $DATA_DIR
fi

cp $CONF/aspen/static-nodes.json $DATA_DIR/

echo "Init genesis for klaytn aspen network"

$BIN/klay init --datadir $DATA_DIR $CONF/aspen/genesis.json
