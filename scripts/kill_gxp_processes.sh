#!/bin/bash

ps -ef | grep "gxp --datadir" | awk '{print $2}' | xargs -I{} kill -9 {}
