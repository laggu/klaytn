#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SOURCE_DIR/common.sh

(cd $GXP_DATAPATH && $GXP_BIN_PATH/istanbul setup --num $NUM_NODES --nodes --verbose --bft --save)
