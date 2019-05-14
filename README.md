[![CircleCI](https://circleci.com/gh/ground-x/klaytn/tree/master.svg?style=svg&circle-token=28de86a436dbe6af811bff7079606433baa43344)](https://circleci.com/gh/ground-x/klaytn/tree/master)
[![codecov](https://codecov.io/gh/ground-x/klaytn/branch/master/graph/badge.svg?token=Tb7cRhQUsU)](https://codecov.io/gh/ground-x/klaytn)

# Table of Contents
<!-- vim-markdown-toc GFM -->

* [Klaytn](#klaytn)
* [Building the source](#building-the-source)
* [Executables](#executables)
* [Running a Core Cell](#running-a-core-cell)
* [Running an Endpoint Node](#running-an-endpoint-node)
  * [Full node on the main Klaytn network](#full-node-on-the-main-klaytn-network)
  * [Full node on the Klaytn test network](#full-node-on-the-klaytn-test-network)
  * [Configuration](#configuration)
  * [Programmatically interfacing Klaytn nodes](#programmatically-interfacing-klaytn-nodes)
* [License](#license)

<!-- vim-markdown-toc -->

## Klaytn

Official golang implementation of the Klaytn protocol.

## Building the source

Building the Klaytn node binaries, such as `kcn`, `kpn`, or `ken`, requires
both a Go (version 1.7 or later) and a C compiler.  You can install them using
your favorite package manager.
Once the dependencies are installed, run

    make all   (or make {kcn, kpn, ken})

## Executables

The klaytn project comes with several wrappers/executables found in the `cmd` directory.

| Command    | Description |
|:----------:|-------------|
| **`kcn`** | The CLI client for Klaytn Consensus Node. |
| **`kpn`** | The CLI client for Klaytn Proxy Node. |
| **`ken`** | The CLI client for Klaytn Endpoint Node, which is the entry point into the Klaytn network (main-, test- or private net).  It can be used by other processes as a gateway into the Klaytn network via JSON RPC endpoints exposed on top of HTTP, WebSocket, gRPC, and/or IPC transports. |
| **`kscn`** | The CLI client for Klaytn ServiceChain Node.  Run `kscn --help` for command-line flags. |
| `abigen` | Source code generator to convert Klaytn contract definitions into easy to use, compile-time type-safe Go packages. |

Both `kcn` and `ken` are capable of running as a full node (default) archive
node (retaining all historical state) or a light node (retrieving data live).

## Running a Core Cell

Core Cell (CC) is a set of consensus node (CN) and one or more proxy nodes
(PNs) and plays a role of generating blocks in the Klaytn network.  Since
setting up and operating a CC is a bit complicated work, we recommend to visit
the [CC Operation Guide](https://docs.klaytn.com/node/cc)
for the detail of CC bootstrapping process.

## Running an Endpoint Node

Going through all the possible command-line flags is out of scope here, but
we've enumerated a few common parameter combos to get you up to speed quickly
on how you can run your own Klaytn Endpoint Node instance.

### Full node on the main Klaytn network

By far the most common scenario is people wanting to simply interact with the
Klaytn network: create accounts; transfer funds; deploy and interact with
contracts. For this particular use-case the user does not care about years-old
historical data, so we can fast-sync quickly to the current state of the
network. To do so:

```
$ ken console
```

This command will:

 * Start a Klaytn Endpoint Node in fast sync mode (default, can be changed with
   the `--syncmode` flag), causing it to download more data in exchange for
   avoiding processing the entire history of the Klaytn network, which is very
   CPU intensive.

### Full node on the Klaytn test network

Transitioning towards developers, if you'd like to play around with creating
Klaytn contracts, you almost certainly would like to do that without any real
money involved until you get the hang of the entire system.  In other words,
instead of attaching to the main network, you want to join the Baoba test
network with your node, which is fully equivalent to the main network, but with
play-Klaytn only.

```
$ ken --baobab console
```

The `console` subcommand have the exact same meaning as above and they are
equally useful on the Baoba testnet too.  Please see above for their
explanations if you've skipped to here.

Specifying the `--baobab` flag however will reconfigure your Klaytn instance a bit:

 * Instead of using the default data directory (`~/.klay` on Linux for example), Klaytn will nest
   itself one level deeper into a `testnet` subfolder (`~/.klay/testnet` on Linux). Note, on OSX
   and Linux this also means that attaching to a running testnet node requires the use of a custom
   endpoint since `ken attach` will try to attach to a production node endpoint by default. E.g.
   `ken attach <datadir>/testnet/klay.ipc`. Windows users are not affected by this.
 * Instead of connecting the main Klaytn network, the client will connect to the test network,
   which uses different P2P bootnodes, different network IDs and genesis states.

*Note: Although there are some internal protective measures to prevent
transactions from crossing over between the main network and test network, you
should make sure to always use separate accounts for play-money and real-money.
Unless you manually move accounts, Klaytn will by default correctly separate
the two networks and will not make any accounts available between them.*

### Configuration

As an alternative to passing the numerous flags to the `ken` binary, you can
also pass a configuration file via:

```
$ ken --config /path/to/your_config.toml
```

To get an idea how the file should look like you can use the `dumpconfig`
subcommand to export your existing configuration:

```
$ ken --your-favourite-flags dumpconfig
```

Do not forget `--rpcaddr 0.0.0.0`, if you want to access RPC from other
containers and/or hosts. By default, `ken` binds to the local interface and RPC
endpoints is not accessible from the outside.

### Programmatically interfacing Klaytn nodes

As a developer, sooner rather than later you'll want to start interacting with
Klaytn and the Klaytn network via your own programs and not manually through
the console. To aid this, Klaytn has built-in support for a JSON-RPC based
APIs. These can be exposed via HTTP, WebSockets, gRPC, and IPC (unix sockets on
unix based platforms, and named pipes on Windows).

The IPC interface is enabled by default and exposes all the APIs supported by
Klaytn, whereas the HTTP, WS, and gRPC interfaces need to manually be enabled
and only expose a subset of APIs due to security reasons.  These can be turned
on/off and configured as you'd expect.

HTTP based JSON-RPC API options:

  * `--rpc` Enable the HTTP-RPC server
  * `--rpcaddr value` HTTP-RPC server listening interface (default: "localhost")
  * `--rpcport value` HTTP-RPC server listening port (default: 8545)
  * `--rpcapi value` API's offered over the HTTP-RPC interface (default: "klay,net,web3")
  * `--rpccorsdomain value` Comma separated list of domains from which to accept cross origin requests (browser enforced)
  * `--ws` Enable the WS-RPC server
  * `--wsaddr value` WS-RPC server listening interface (default: "localhost")
  * `--wsport value` WS-RPC server listening port (default: 8546)
  * `--wsapi value` API's offered over the WS-RPC interface (default: "klay,net,web3")
  * `--wsorigins value` Origins from which to accept websockets requests
  * `--grpc` Enable the gRPC server
  * `--grpcaddr value` gRPC server listening interface (default: "localhost")
  * `--grpcport value` gRPC server listening port (default: 8547)
  * `--ipcdisable` Disable the IPC-RPC server
  * `--ipcapi value` API's offered over the IPC-RPC interface (default: "admin,klay,miner,net,personal,txpool,web3")
  * `--ipcpath` Filename for IPC socket/pipe within the datadir (explicit paths escape it)
  * `--srvtype` HTTP-RPC/WebSocket-RPC server module [http, fasthttp] (default: "http")

You'll need to use your own programming environments' capabilities (libraries,
tools, etc) to connect via HTTP, WS, gRPC or IPC to a Klaytn node configured
with the above flags and you'll need to speak
[JSON-RPC](http://www.jsonrpc.org/specification) on all transports. You can
reuse the same connection for multiple requests!


## License

The klaytn library (i.e. all code outside of the `cmd` directory) is licensed under the
[GNU Lesser General Public License v3.0](https://www.gnu.org/licenses/lgpl-3.0.en.html), also
included in our repository in the `COPYING.LESSER` file.

The klaytn binaries (i.e. all code inside of the `cmd` directory) is licensed under the
[GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html), also included
in our repository in the `COPYING` file.
