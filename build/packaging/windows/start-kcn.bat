@echo off

set HOME=%~dp0
set CONF=%HOME%\conf

call %CONF%\kcn-conf.cmd

REM Check if exist data directory
set "NOT_INIT="
IF NOT EXIST %KLAY_HOME% (
    set NOT_INIT=1
)
IF NOT EXIST %DATA_DIR% (
    set NOT_INIT=1
)

IF DEFINED NOT_INIT (
    echo "[ERROR] : kcn is not initiated, Initiate kcn with genesis file first."
    GOTO end
)

set OPTIONS=--networkid %NETWORK_ID% --datadir %DATA_DIR% --port %PORT% --srvtype fasthttp --metrics --prometheus --verbosity 3 ^
--txpool.globalslots 4096 --txpool.globalqueue 4096 --txpool.accountslots 4096 --txpool.accountqueue 4096 --nodiscover ^
--syncmode full --mine --maxpeers 5000 --db.leveldb.cache-size 10240

%HOME%\bin\kcn.exe %OPTIONS%

:end
@pause
