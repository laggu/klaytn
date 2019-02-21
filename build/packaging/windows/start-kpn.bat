@echo off

set HOME=%~dp0
set CONF=%HOME%\conf

call %CONF%\kpn-conf.cmd

REM Check if exist data directory
set "NOT_INIT="
IF NOT EXIST %KLAY_HOME% (
    set NOT_INIT=1
)
IF NOT EXIST %DATA_DIR% (
    set NOT_INIT=1
)

IF DEFINED NOT_INIT (
    echo "[ERROR] : kpn is not initiated, Initiate kpn with genesis file first."
    GOTO end
)

set OPTIONS=--networkid %NETWORK_ID%  --datadir %DATA_DIR%  --port %PORT%  --rpc --rpcapi klay --rpcport %RPC_PORT%  --rpcaddr 0.0.0.0 ^
--rpccorsdomain *  --rpcvhosts * --ws  --wsaddr 0.0.0.0 --wsport %WS_PORT% --wsorigins * --srvtype fasthttp --metrics --prometheus ^
--verbosity 3 --txpool.globalslots 2048 --txpool.globalqueue 2048 --txpool.accountslots 2048 --txpool.accountqueue 2048 --txpool.nolocals ^
--nodiscover  --syncmode full  --mine  --maxpeers 5000 --db.leveldb.cache-size 10240

%HOME%\bin\kpn.exe %OPTIONS%

:end
@pause
