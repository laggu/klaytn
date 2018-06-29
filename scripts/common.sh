#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

if [ -z "$GXP_DATAPATH" ]; then
  echo "GXP_DATAPATH should be set first!!!"
  exit 1
fi

if [ -z "$GXP_BIN_PATH" ]; then
  GXP_BIN_PATH=$SOURCE_DIR/../build/bin
fi

if [ -z "$NETWORK_ID" ]; then
  NETWORK_ID=3905
fi

if [ -z "$NUM_NODES" ]; then
  NUM_NODES=4
fi

PORT_START=30303
RPC_PORT_START=8123
WS_PORT_START=8546

mkdir -p $GXP_DATAPATH

echo "GXP_BIN_PATH = $GXP_BIN_PATH"
echo "GXP_DATAPATH = $GXP_DATAPATH"
echo "NETWORK_ID = $NETWORK_ID"
echo "Number of validator nodes = $NUM_NODES"
