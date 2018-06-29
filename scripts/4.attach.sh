#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SOURCE_DIR/common.sh

$GXP_BIN_PATH/gxp attach http://localhost:$RPC_PORT_START
