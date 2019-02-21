@echo off

set HOME=%~dp0
set CONF=%HOME%\conf

call %CONF%\ken-conf.cmd

REM Check if exist data directory
set "NOT_INIT="
IF NOT EXIST %KLAY_HOME% (
    set NOT_INIT=1
)
IF NOT EXIST %DATA_DIR% (
    set NOT_INIT=1
)

IF DEFINED NOT_INIT (
    echo "[ERROR] : ken is not initiated, Initiate ken with genesis file first."
    GOTO end
)

set OPTIONS=--networkid %NETWORK_ID% --datadir %DATA_DIR% --port %PORT% --rpc --rpcapi klay --rpcport %RPC_PORT% --rpcaddr 0.0.0.0 ^
--rpccorsdomain * --rpcvhosts * --ws --wsaddr 0.0.0.0 --wsport %WS_PORT% --wsorigins * --srvtype fasthttp --metrics --prometheus ^
--verbosity 3 --txpool.globalslots 1024 --txpool.globalqueue 1024 --txpool.accountslots 1024 --txpool.accountqueue 1024 --nodiscover ^
--syncmode full --mine

%HOME%\bin\ken.exe %OPTIONS%

:end
@pause
