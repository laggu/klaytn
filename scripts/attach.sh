#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SOURCE_DIR/common.sh
NODE_TO_ATTACH=0

if [ -n "$1" ]; then
  NODE_TO_ATTACH=$1
fi

RPC_PORT=$(( ($RPC_PORT_START + $NODE_TO_ATTACH * 10000) % 65536))

ATTACH="http://localhost:$RPC_PORT"
echo "!!!Attaching to $ATTACH"

$GXP_BIN_PATH/gxp attach $ATTACH
