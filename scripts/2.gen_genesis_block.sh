#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SOURCE_DIR/common.sh

for (( i = 0; i < $NUM_NODES; i++ )); do
  (cd $GXP_DATAPATH && 
    cp static-nodes.json $GXP_DATAPATH/$i &&
    $GXP_BIN_PATH/gxp --datadir $GXP_DATAPATH/$i init genesis.json)
done
