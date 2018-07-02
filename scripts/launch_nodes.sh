#!/bin/bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source $SOURCE_DIR/common.sh

for (( i = 0; i < $NUM_NODES; i++ )); do
  PORT=$(( ($PORT_START + $i * 10101) % 65536 ))
  RPC_PORT=$(( ($RPC_PORT_START + $i * 10000) % 65536))
  WS_PORT=$(( ($WS_PORT_START + $i * 10000) % 65536 ))
  GASPRICE=0
  echo "Launching a node with port=$PORT RPC_PORT=$RPC_PORT WS_PORT=$WS_PORT"

  $GXP_BIN_PATH/gxp --datadir $GXP_DATAPATH/$i --port $PORT --rpc --rpcaddr 0.0.0.0 --rpcport $RPC_PORT --rpccorsdomain "*" --nodiscover --networkid $NETWORK_ID --nat "any" --wsport $WS_PORT --ws --wsaddr 0.0.0.0 --wsorigins="*" --rpcapi "db,txpool,gxp,net,web3,miner,personal,admin,rpc" --mine --gasprice $GASPRICE 1>$GXP_DATAPATH/$i/stdout 2>$GXP_DATAPATH/$i/stderr &
done
