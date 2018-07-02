# Table of contents

<!-- vim-markdown-toc GFM -->

* [How to run test network in a single node](#how-to-run-test-network-in-a-single-node)
  * [Getting Started](#getting-started)
    * [Launching GXP nodes](#launching-gxp-nodes)
    * [Attaching to a GXP node](#attaching-to-a-gxp-node)
  * [Log files](#log-files)
  * [Killing GXP processes](#killing-gxp-processes)
  * [Change Options](#change-options)
    * [number of nodes](#number-of-nodes)
    * [network ID](#network-id)
    * [Change ports](#change-ports)

<!-- vim-markdown-toc -->

# How to run test network in a single node

## Getting Started

### Launching GXP nodes
You can run GXP with the default setting by executing the following commands:
```
$ cd scripts
$ export GXP_DATAPATH=~/.gxp/data
$ ./run_gxp.sh
```

### Attaching to a GXP node
```
$ ./attach.sh

> gxp.blockNumber
```

## Log files
You can find logfiles in `$GXP_DATAPATH/node_id/stdout` and `$GXP_DATAPATH/node_id/stderr`

## Killing GXP processes
You can kill all GXP processes like below:
```
$ ./kill_gxp_processes.sh
```

## Change Options
### number of nodes
You can change the number of nodes by changing `$NUM_NODES` before executing the scripts.

```
$ export NUM_NODES=5
```

### network ID
You can change the network ID by changing `$NETWORK_ID` before executing the scripts.

```
$ export NETWORK_ID=3095
```

### Change ports
You can change various PORTs before executing the scripts.

```
$ export PORT_START=30303
$ export RPC_PORT_START=8123
$ export WS_PORT_START=8546
```
