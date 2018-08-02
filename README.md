[![CircleCI](https://circleci.com/gh/ground-x/go-gxplatform/tree/master.svg?style=svg&circle-token=28de86a436dbe6af811bff7079606433baa43344)](https://circleci.com/gh/ground-x/go-gxplatform/tree/master)

# Table of Contents
<!-- vim-markdown-toc GFM -->

* [Go Klaytn](#go-klaytn)
* [Building the source](#building-the-source)
* [Executables](#executables)
* [Running klay](#running-klay)
  * [Full node on the main Klaytn network](#full-node-on-the-main-klaytn-network)
  * [Full node on the Klaytn test network](#full-node-on-the-klaytn-test-network)
  * [Configuration (PoW)](#configuration-pow)
  * [Programatically interfacing Klaytn nodes](#programatically-interfacing-klay-nodes)
  * [Operating a private network](#operating-a-private-network)
    * [Defining the private genesis state](#defining-the-private-genesis-state)
  * [Configuration (istanbul-BFT)](#configuration-istanbul-bft)
  * [sol2proto](#sol2proto)
  * [grpc-contract](#grpc-contract)
* [License](#license)

<!-- vim-markdown-toc -->

## Go Klaytn

Official golang implementation of the Klaytn protocol.

## Building the source

Building klay requires both a Go (version 1.7 or later) and a C compiler.
You can install them using your favourite package manager.
Once the dependencies are installed, run

    make klay   or  make all

## Executables

The go-klaytn project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **`klay`** | Our main Klaytn CLI client. It is the entry point into the Klaytn network (main-, test- or private net), capable of running as a full node (default) archive node (retaining all historical state) or a light node (retrieving data live). It can be used by other processes as a gateway into the Klaytn network via JSON RPC endpoints exposed on top of HTTP, WebSocket and/or IPC transports. `klay --help`|
| `istanbul` | tools for configuring Istanbul BFT (IBFT) network.
| `abigen` | Source code generator to convert Klaytn contract definitions into easy to use, compile-time type-safe Go packages. |
| `sol2proto` | The Klaytn ABI to gRPC protobuf transpiler. |
| `grpc-contract` | A tool to generate the grpc server code for a contract. |

## Running klay

Going through all the possible command line flags is out of scope here, but we've
enumerated a few common parameter combos to get you up to speed quickly on how you can run your
own Klaytn instance.

### Full node on the main Klaytn network

By far the most common scenario is people wanting to simply interact with the Klaytn network:
create accounts; transfer funds; deploy and interact with contracts. For this particular use-case
the user doesn't care about years-old historical data, so we can fast-sync quickly to the current
state of the network. To do so:

```
$ klay console
```

This command will:

 * Start klay in fast sync mode (default, can be changed with the `--syncmode` flag), causing it to
   download more data in exchange for avoiding processing the entire history of the Klaytn network,
   which is very CPU intensive.

### Full node on the Klaytn test network

Transitioning towards developers, if you'd like to play around with creating Klaytn contracts, you
almost certainly would like to do that without any real money involved until you get the hang of the
entire system. In other words, instead of attaching to the main network, you want to join the **test**
network with your node, which is fully equivalent to the main network, but with play-Klaytn only.

```
$ klay --testnet console
```

The `console` subcommand have the exact same meaning as above and they are equally useful on the
testnet too. Please see above for their explanations if you've skipped to here.

Specifying the `--testnet` flag however will reconfigure your Klaytn instance a bit:

 * Instead of using the default data directory (`~/.klay` on Linux for example), Klaytn will nest
   itself one level deeper into a `testnet` subfolder (`~/.klay/testnet` on Linux). Note, on OSX
   and Linux this also means that attaching to a running testnet node requires the use of a custom
   endpoint since `klay attach` will try to attach to a production node endpoint by default. E.g.
   `klay attach <datadir>/testnet/klay.ipc`. Windows users are not affected by this.
 * Instead of connecting the main Klaytn network, the client will connect to the test network,
   which uses different P2P bootnodes, different network IDs and genesis states.
   
*Note: Although there are some internal protective measures to prevent transactions from crossing
over between the main network and test network, you should make sure to always use separate accounts
for play-money and real-money. Unless you manually move accounts, Klaytn will by default correctly
separate the two networks and will not make any accounts available between them.*

### Configuration (PoW)

As an alternative to passing the numerous flags to the `klay` binary, you can also pass a configuration file via:

```
$ klay --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig` subcommand to export your existing configuration:

```
$ klay --your-favourite-flags dumpconfig
```


Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other containers and/or hosts. By default, `klay` binds to the local interface and RPC endpoints is not accessible from the outside.

### Programatically interfacing Klaytn nodes

As a developer, sooner rather than later you'll want to start interacting with Klaytn and the Klaytn
network via your own programs and not manually through the console. To aid this, Klaytn has built-in
support for a JSON-RPC based APIs. These can be
exposed via HTTP, WebSockets and IPC (unix sockets on unix based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by Klaytn, whereas the HTTP
and WS interfaces need to manually be enabled and only expose a subset of APIs due to security reasons.
These can be turned on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr` HTTP-RPC server listening interface (default: "localhost")
  * `--rpcport` HTTP-RPC server listening port (default: 8545)
  * `--rpcapi` API's offered over the HTTP-RPC interface (default: "klay,net,web3")
  * `--rpccorsdomain` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr` WS-RPC server listening interface (default: "localhost")
  * `--wsport` WS-RPC server listening port (default: 8546)
  * `--wsapi` API's offered over the WS-RPC interface (default: "klay,net,web3")
  * `--wsorigins` Origins from which to accept websockets requests
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi` API's offered over the IPC-RPC interface (default: "admin,klay,miner,net,personal,txpool,web3")
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)
  * `--srvtype` HTTP-RPC/WebSocket-RPC server module [http, fasthttp] (default: "http")

You'll need to use your own programming environments' capabilities (libraries, tools, etc) to connect
via HTTP, WS or IPC to a Klaytn node configured with the above flags and you'll need to speak [JSON-RPC](http://www.jsonrpc.org/specification)
on all transports. You can reuse the same connection for multiple requests!

### Operating a private network

Maintaining your own private network is more involved as a lot of configurations taken for granted in
the official networks need to be manually set up.

#### Defining the private genesis state

First, you'll need to create the genesis state of your networks, which all nodes need to be aware of
and agree upon. This consists of a small JSON file (e.g. call it `genesis.json`):

```json
{
  "config": {
        "chainId": 0,
        "homesteadBlock": 0
    },
  "alloc"      : {},
  "coinbase"   : "0x0000000000000000000000000000000000000000",
  "difficulty" : "0x20000",
  "extraData"  : "",
  "gasLimit"   : "0x2fefd8",
  "nonce"      : "0x0000000000000042",
  "mixhash"    : "0x0000000000000000000000000000000000000000000000000000000000000000",
  "parentHash" : "0x0000000000000000000000000000000000000000000000000000000000000000",
  "timestamp"  : "0x00"
}
```

The above fields should be fine for most purposes, although we'd recommend changing the `nonce` to
some random value so you prevent unknown remote nodes from being able to connect to you. If you'd
like to pre-fund some accounts for easier testing, you can populate the `alloc` field with account
configs:

```json
"alloc": {
  "0x0000000000000000000000000000000000000001": {"balance": "111111111"},
  "0x0000000000000000000000000000000000000002": {"balance": "222222222"}
}
```

With the genesis state defined in the above JSON file, you'll need to initialize **every** Klaytn node
with it prior to starting it up to ensure all blockchain parameters are correctly set:

```
$ klay init path/to/genesis.json
```

### Configuration (klaytn-tools - istanbul)

참조 : https://github.com/ground-x/klaytn-tools/blob/master/README.md

istanbul의 setup 명령어로 만들어진 nodekey를 각 노드의 nodekey 화일에 저장한다.

```nodekey
81a74dc939a2e023a3743396cc9beb04cc092e11aceadf07e5f5e4299bb9a8c6
```

미리 생성된 계정이 있는 경우에는 genesis.json 파일에 alloc 설정에 계정에 대한 balance 값을 추가함

nodekey 화일과 static-nodes.json 화일을 datadir로 복사한다.
Genesis 블락을 다음의 명령어로 생성한다.
```
./klay --datadir $DATAPATH init genesis.json

```
각 노드를 다음 명령어로 실행함. 실행시에 --mine 옵션으로 마이닝을 실행하고 --gasprice 0 옵션으로 gasprice를 0으로 세팅함.
```
klay --srvtype fasthttp --datadir $DATAPATH --port 30303 --rpc --rpcaddr 0.0.0.0 --rpcport "8123" --rpccorsdomain "*"
--nodiscover --networkid 3900 --nat "any" --wsport "8546" --ws --wsaddr 0.0.0.0 --wsorigins="*"
--rpcapi "db,txpool,klay,net,web3,miner,personal,admin,rpc" --mine --gasprice 0 console

```
동일 머신에서 수행시에는 `--port`, `--rpcport`, `--wsport` 옵션을 다르게 설정하고 `--networkid`는 동일하게 설정함.
In this case, please make sure the IP and port number for each validator in
`static-nodes.json` are the same as your local IP (e.g., 127.0.0.1) and the
value used with`--port` when running `klay`.

### sol2proto
Solidity ABI to gRPC protobuf IDL transpiler

The `context` is in the standard library Go 1.7 already. Make sure the latest version of grpc and protoc plugin are installed.
```
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
```

```bash
sol2proto --pkg awesome --abi MyAwesomeContract.abi > my_awesome_contract.proto
```
### grpc-contract
A tool to generate the grpc server code for a contract

```bash
grpc-contract --types $(filename) --path ./protobuf --pb-path ./protobuf
```

## License

The go-klaytn library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also
included in our repository in the `COPYING.LESSER` file.

The go-klaytn binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.
