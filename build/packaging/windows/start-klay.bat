@echo off

set HOME=%~dp0
set CONF=%HOME%\conf

call %CONF%\klay-conf.cmd

REM Check if exist data directory
set "NOT_INIT="
IF NOT EXIST %KLAY_HOME% (
    set NOT_INIT=1
)
IF NOT EXIST %DATA_DIR% (
    set NOT_INIT=1
)

IF DEFINED NOT_INIT (
    echo "[ERROR] : klaytn is not initiated, Initiate a klaytn with genesis file first."
    GOTO end
)

IF "%NODE_TYPE%"=="CN" (
    set OPTIONS=--nodetype cn --networkid %NETWORK_ID% --datadir %DATA_DIR% --port %PORT% --srvtype fasthttp --metrics --prometheus --verbosity 3 ^
--txpool.globalslots 4096 --txpool.globalqueue 4096 --txpool.accountslots 4096 --txpool.accountqueue 4096 --nodiscover ^
--syncmode full --mine --maxpeers 5000 --db.leveldb.cache-size 10240
) ELSE IF "%NODE_TYPE%"=="BN" (
    set OPTIONS=--nodetype bn --networkid %NETWORK_ID%  --datadir %DATA_DIR%  --port %PORT%  --rpc --rpcapi klay --rpcport %RPC_PORT%  --rpcaddr 0.0.0.0 ^
--rpccorsdomain *  --rpcvhosts * --ws  --wsaddr 0.0.0.0 --wsport %WS_PORT% --wsorigins * --srvtype fasthttp --metrics --prometheus ^
--verbosity 3 --txpool.globalslots 2048 --txpool.globalqueue 2048 --txpool.accountslots 2048 --txpool.accountqueue 2048 --txpool.nolocals ^
--nodiscover  --syncmode full  --mine  --maxpeers 5000 --db.leveldb.cache-size 10240
) ELSE IF "%NODE_TYPE%"=="RN" (
    set OPTIONS=--nodetype rn --networkid %NETWORK_ID% --datadir %DATA_DIR% --port %PORT% --rpc --rpcapi klay --rpcport %RPC_PORT% --rpcaddr 0.0.0.0 ^
--rpccorsdomain * --rpcvhosts * --ws --wsaddr 0.0.0.0 --wsport %WS_PORT% --wsorigins * --srvtype fasthttp --metrics --prometheus ^
--verbosity 3 --txpool.globalslots 1024 --txpool.globalqueue 1024 --txpool.accountslots 1024 --txpool.accountqueue 1024 --nodiscover ^
--syncmode full --mine
) ELSE (
    set OPTIONS=--nodetype rn --networkid %NETWORK_ID% --datadir %DATA_DIR% --port %PORT% --rpc --rpcapi klay --rpcport %RPC_PORT% --rpcaddr 0.0.0.0 ^
--rpccorsdomain * --rpcvhosts * --ws --wsaddr 0.0.0.0 --wsport %WS_PORT% --wsorigins * --srvtype fasthttp --metrics --prometheus ^
--verbosity 3 --txpool.globalslots 1024 --txpool.globalqueue 1024 --txpool.accountslots 1024 --txpool.accountqueue 1024 --nodiscover ^
--syncmode full --mine
)

%HOME%\bin\klay.exe %OPTIONS%

:end
@pause
