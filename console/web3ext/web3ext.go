// Copyright 2018 The klaytn Authors
// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
//
// This file is derived from internal/web3ext/web3ext.go (2018/06/04).
// Modified and improved for the klaytn development.

package web3ext

var Modules = map[string]string{
	"admin":      Admin_JS,
	"debug":      Debug_JS,
	"klay":       Klay_JS,
	"miner":      Miner_JS,
	"net":        Net_JS,
	"personal":   Personal_JS,
	"rpc":        RPC_JS,
	"txpool":     TxPool_JS,
	"istanbul":   Istanbul_JS,
	"bridge":     Bridge_JS,
	"clique":     CliqueJs,
	"governance": Governance_JS,
	"bootnode":   Bootnode_JS,
}

const Bootnode_JS = `
web3._extend({
	property: 'bootnode',
	methods: [
		new web3._extend.Method({
			name: 'createUpdateNode',
			call: 'bootnode_createUpdateNode',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getNode',
			call: 'bootnode_getNode',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getTableEntries',
			call: 'bootnode_getTableEntries'
		}),
		new web3._extend.Method({
			name: 'getTableReplacements',
			call: 'bootnode_getTableReplacements'
		}),
		new web3._extend.Method({
			name: 'deleteNode',
			call: 'bootnode_deleteNode',
			params: 1
		})
	],
	properties: []
});
`
const Governance_JS = `
web3._extend({
	property: 'governance',
	methods: [
		new web3._extend.Method({
			name: 'vote',
			call: 'governance_vote',
			params: 2
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'showTally',
			getter: 'governance_showTally',
		}),
		new web3._extend.Property({
			name: 'totalVotingPower',
			getter: 'governance_totalVotingPower',
		}),
		new web3._extend.Property({
			name: 'myVotes',
			getter: 'governance_myVotes',
		}),
		new web3._extend.Property({
			name: 'myVotingPower',
			getter: 'governance_myVotingPower',
		}),
		new web3._extend.Property({
			name: 'chainConfig',
			getter: 'governance_chainConfig',
		}),	
		new web3._extend.Property({
			name: 'nodeAddress',
			getter: 'governance_nodeAddress',
		}),	
	]
});
`
const Admin_JS = `
web3._extend({
	property: 'admin',
	methods: [
		new web3._extend.Method({
			name: 'addPeer',
			call: 'admin_addPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'removePeer',
			call: 'admin_removePeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'exportChain',
			call: 'admin_exportChain',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'importChain',
			call: 'admin_importChain',
			params: 1
		}),
		new web3._extend.Method({
			name: 'sleepBlocks',
			call: 'admin_sleepBlocks',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startRPC',
			call: 'admin_startRPC',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopRPC',
			call: 'admin_stopRPC'
		}),
		new web3._extend.Method({
			name: 'startWS',
			call: 'admin_startWS',
			params: 4,
			inputFormatter: [null, null, null, null]
		}),
		new web3._extend.Method({
			name: 'stopWS',
			call: 'admin_stopWS'
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'admin_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'admin_peers'
		}),
		new web3._extend.Property({
			name: 'datadir',
			getter: 'admin_datadir'
		}),
	]
});
`

const Debug_JS = `
web3._extend({
	property: 'debug',
	methods: [
		new web3._extend.Method({
			name: 'printBlock',
			call: 'debug_printBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getBlockRlp',
			call: 'debug_getBlockRlp',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setHead',
			call: 'debug_setHead',
			params: 1
		}),
		new web3._extend.Method({
			name: 'seedHash',
			call: 'debug_seedHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'dumpBlock',
			call: 'debug_dumpBlock',
			params: 1
		}),
		new web3._extend.Method({
			name: 'chaindbProperty',
			call: 'debug_chaindbProperty',
			params: 1,
			outputFormatter: console.log
		}),
		new web3._extend.Method({
			name: 'chaindbCompact',
			call: 'debug_chaindbCompact',
		}),
		new web3._extend.Method({
			name: 'metrics',
			call: 'debug_metrics',
			params: 1
		}),
		new web3._extend.Method({
			name: 'verbosity',
			call: 'debug_verbosity',
			params: 1
		}),
		new web3._extend.Method({
			name: 'verbosityByName',
			call: 'debug_verbosityByName',
			params: 2
		}),
		new web3._extend.Method({
			name: 'verbosityByID',
			call: 'debug_verbosityByID',
			params: 2
		}),
		new web3._extend.Method({
			name: 'vmodule',
			call: 'debug_vmodule',
			params: 1
		}),
		new web3._extend.Method({
			name: 'backtraceAt',
			call: 'debug_backtraceAt',
			params: 1,
		}),
		new web3._extend.Method({
			name: 'stacks',
			call: 'debug_stacks',
			params: 0,
			outputFormatter: console.log
		}),
		new web3._extend.Method({
			name: 'freeOSMemory',
			call: 'debug_freeOSMemory',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'setGCPercent',
			call: 'debug_setGCPercent',
			params: 1,
		}),
		new web3._extend.Method({
			name: 'memStats',
			call: 'debug_memStats',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'gcStats',
			call: 'debug_gcStats',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'startPProf',
			call: 'debug_startPProf',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'stopPProf',
			call: 'debug_stopPProf',
			params: 0
		}),
		new web3._extend.Method({
			name: 'isPProfRunning',
			call: 'debug_isPProfRunning',
			params: 0
		}),
		new web3._extend.Method({
			name: 'cpuProfile',
			call: 'debug_cpuProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startCPUProfile',
			call: 'debug_startCPUProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopCPUProfile',
			call: 'debug_stopCPUProfile',
			params: 0
		}),
		new web3._extend.Method({
			name: 'goTrace',
			call: 'debug_goTrace',
			params: 2
		}),
		new web3._extend.Method({
			name: 'startGoTrace',
			call: 'debug_startGoTrace',
			params: 1
		}),
		new web3._extend.Method({
			name: 'stopGoTrace',
			call: 'debug_stopGoTrace',
			params: 0
		}),
		new web3._extend.Method({
			name: 'blockProfile',
			call: 'debug_blockProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setBlockProfileRate',
			call: 'debug_setBlockProfileRate',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeBlockProfile',
			call: 'debug_writeBlockProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'mutexProfile',
			call: 'debug_mutexProfile',
			params: 2
		}),
		new web3._extend.Method({
			name: 'setMutexProfileRate',
			call: 'debug_setMutexProfileRate',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeMutexProfile',
			call: 'debug_writeMutexProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'writeMemProfile',
			call: 'debug_writeMemProfile',
			params: 1
		}),
		new web3._extend.Method({
			name: 'traceBlock',
			call: 'debug_traceBlock',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockFromFile',
			call: 'debug_traceBlockFromFile',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockByNumber',
			call: 'debug_traceBlockByNumber',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, null]
		}),
		new web3._extend.Method({
			name: 'traceBlockByHash',
			call: 'debug_traceBlockByHash',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'traceTransaction',
			call: 'debug_traceTransaction',
			params: 2,
			inputFormatter: [null, null]
		}),
		new web3._extend.Method({
			name: 'preimage',
			call: 'debug_preimage',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getBadBlocks',
			call: 'debug_getBadBlocks',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'storageRangeAt',
			call: 'debug_storageRangeAt',
			params: 5,
		}),
		new web3._extend.Method({
			name: 'getModifiedAccountsByNumber',
			call: 'debug_getModifiedAccountsByNumber',
			params: 2,
			inputFormatter: [null, null],
		}),
		new web3._extend.Method({
			name: 'getModifiedAccountsByHash',
			call: 'debug_getModifiedAccountsByHash',
			params: 2,
			inputFormatter:[null, null],
		}),
		new web3._extend.Method({
			name: 'setVMLogTarget',
			call: 'debug_setVMLogTarget',
			params: 1
		}),
	],
	properties: []
});
`

const Klay_JS = `
var blockWithConsensusInfoCall = function (args) {
    return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? "klay_getBlockWithConsensusInfoByHash" : "klay_getBlockWithConsensusInfoByNumber";
};

web3._extend({
	property: 'klay',
	methods: [
		new web3._extend.Method({
			name: 'getBlockReceipts',
			call: 'klay_getBlockReceipts',
			params: 1,
			outputFormatter: function(receipts) {
				var formatted = [];
				for (var i = 0; i < receipts.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionReceiptFormatter(receipts[i]));
				}
				return formatted;
			}
		}),
		new web3._extend.Method({
			name: 'sign',
			call: 'klay_sign',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Method({
			name: 'resend',
			call: 'klay_resend',
			params: 3,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, web3._extend.utils.fromDecimal, web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'klay_signTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'getValidators',
			call: 'klay_getValidators',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.Method({
			name: 'accountCreated',
			call: 'klay_accountCreated'
			params: 2,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter, web3._extend.formatters.inputDefaultBlockNumberFormatter],
		}),
		new web3._extend.Method({
			name: 'getBlockWithConsensusInfo',
			call: blockWithConsensusInfoCall,
			params: 1,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.Method({
			name: 'getBlockWithConsensusInfoRange',
			call: 'klay_getBlockWithConsensusInfoByNumberRange',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.formatters.inputBlockNumberFormatter]
		}),
		new web3._extend.Method({
			name: 'submitTransaction',
			call: 'klay_submitTransaction',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter]
		}),
		new web3._extend.Method({
			name: 'getRawTransaction',
			call: 'klay_getRawTransactionByHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getRawTransactionFromBlock',
			call: function(args) {
				return (web3._extend.utils.isString(args[0]) && args[0].indexOf('0x') === 0) ? 'klay_getRawTransactionByBlockHashAndIndex' : 'klay_getRawTransactionByBlockNumberAndIndex';
			},
			params: 2,
			inputFormatter: [web3._extend.formatters.inputBlockNumberFormatter, web3._extend.utils.toHex]
		}),
		new web3._extend.Method({
			name: 'writeThroughCaching',
			call: 'klay_writeThroughCaching',
		}),
		new web3._extend.Method({
			name: 'isParallelDBWrite',
			call: 'klay_isParallelDBWrite',
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'pendingTransactions',
			getter: 'klay_pendingTransactions',
			outputFormatter: function(txs) {
				var formatted = [];
				for (var i = 0; i < txs.length; i++) {
					formatted.push(web3._extend.formatters.outputTransactionFormatter(txs[i]));
					formatted[i].blockHash = null;
				}
				return formatted;
			}
		}),
        new web3._extend.Property({
            name : 'rewardbase',
            getter: 'klay_rewardbase',
           
        }),
	]
});
`

const Miner_JS = `
web3._extend({
	property: 'miner',
	methods: [
		new web3._extend.Method({
			name: 'start',
			call: 'miner_start',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'stop',
			call: 'miner_stop'
		}),
		new web3._extend.Method({
			name: 'setCoinbase',
			call: 'miner_setCoinbase',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.Method({
			name: 'setRewardbase',
			call: 'miner_setRewardbase',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.Method({
			name: 'setRewardContract',
			call: 'miner_setRewardContract',
			params: 1,
			inputFormatter: [web3._extend.formatters.inputAddressFormatter]
		}),
		new web3._extend.Method({
			name: 'setExtra',
			call: 'miner_setExtra',
			params: 1
		}),
		new web3._extend.Method({
			name: 'setGasPrice',
			call: 'miner_setGasPrice',
			params: 1,
			inputFormatter: [web3._extend.utils.fromDecimal]
		}),
		new web3._extend.Method({
			name: 'getHashrate',
			call: 'miner_getHashrate'
		}),
	],
	properties: []
});
`

const Net_JS = `
web3._extend({
	property: 'net',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'version',
			getter: 'net_version'
		}),
		new web3._extend.Property({
			name: 'networkID',
			getter: 'net_networkID'
		}),
	]
});
`

const Personal_JS = `
web3._extend({
	property: 'personal',
	methods: [
		new web3._extend.Method({
			name: 'importRawKey',
			call: 'personal_importRawKey',
			params: 2
		}),
		new web3._extend.Method({
			name: 'sign',
			call: 'personal_sign',
			params: 3,
			inputFormatter: [null, web3._extend.formatters.inputAddressFormatter, null]
		}),
		new web3._extend.Method({
			name: 'ecRecover',
			call: 'personal_ecRecover',
			params: 2
		}),
		new web3._extend.Method({
			name: 'openWallet',
			call: 'personal_openWallet',
			params: 2
		}),
		new web3._extend.Method({
			name: 'deriveAccount',
			call: 'personal_deriveAccount',
			params: 3
		}),
		new web3._extend.Method({
			name: 'signTransaction',
			call: 'personal_signTransaction',
			params: 2,
			inputFormatter: [web3._extend.formatters.inputTransactionFormatter, null]
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'listWallets',
			getter: 'personal_listWallets'
		}),
	]
})
`

const RPC_JS = `
web3._extend({
	property: 'rpc',
	methods: [],
	properties: [
		new web3._extend.Property({
			name: 'modules',
			getter: 'rpc_modules'
		}),
	]
});
`

const TxPool_JS = `
web3._extend({
	property: 'txpool',
	methods: [],
	properties:
	[
		new web3._extend.Property({
			name: 'content',
			getter: 'txpool_content'
		}),
		new web3._extend.Property({
			name: 'inspect',
			getter: 'txpool_inspect'
		}),
		new web3._extend.Property({
			name: 'status',
			getter: 'txpool_status',
			outputFormatter: function(status) {
				status.pending = web3._extend.utils.toDecimal(status.pending);
				status.queued = web3._extend.utils.toDecimal(status.queued);
				return status;
			}
		}),
	]
});
`

const Istanbul_JS = `
web3._extend({
	property: 'istanbul',
	methods:
	[
		new web3._extend.Method({
			name: 'getSnapshot',
			call: 'istanbul_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSnapshotAtHash',
			call: 'istanbul_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getValidators',
			call: 'istanbul_getValidators',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getValidatorsAtHash',
			call: 'istanbul_getValidatorsAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'propose',
			call: 'istanbul_propose',
			params: 2
		}),
		new web3._extend.Method({
			name: 'discard',
			call: 'istanbul_discard',
			params: 1
		})
	],
	properties:
	[
		new web3._extend.Property({
			name: 'candidates',
			getter: 'istanbul_candidates'
		}),
	]
});
`
const Bridge_JS = `
web3._extend({
	property: 'bridge',
	methods:
	[
		new web3._extend.Method({
			name: 'addPeer',
			call: 'bridge_addPeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'removePeer',
			call: 'bridge_removePeer',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getChildChainIndexingEnabled',
			call: 'bridge_getChildChainIndexingEnabled'
		}),
		new web3._extend.Method({
			name: 'convertChildChainBlockHashToParentChainTxHash',
			call: 'bridge_convertChildChainBlockHashToParentChainTxHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getReceiptFromParentChain',
			call: 'bridge_getReceiptFromParentChain',
			params: 1
		}),
		new web3._extend.Method({
			name: 'deployGateway',
			call: 'bridge_deployGateway',
			params: 0
		}),
		new web3._extend.Method({
			name: 'deployGatewayOnLocal',
			call: 'bridge_deployGatewayOnLocalChain',
			params: 0
		}),
		new web3._extend.Method({
			name: 'deployGatewayOnRemote',
			call: 'bridge_deployGatewayOnParentChain',
			params: 0
		}),
		new web3._extend.Method({
			name: 'subscribeGateway',
			call: 'bridge_subscribeEventGateway',
			params: 2
		}),
		new web3._extend.Method({
			name: 'anchoring',
			call: 'bridge_anchoring',
			params: 1
		}),
		new web3._extend.Method({
			name: 'registerGateway',
			call: 'bridge_registerGateway',
			params: 2
		}),
		new web3._extend.Method({
			name: 'unRegisterGateway',
			call: 'bridge_unRegisterGateway',
			params: 1
		}),
		new web3._extend.Method({
			name: 'registerToken',
			call: 'bridge_registerToken',
			params: 2
		}),
		new web3._extend.Method({
			name: 'unRegisterToken',
			call: 'bridge_unRegisterToken',
			params: 1
		}),
	],
    properties: [
		new web3._extend.Property({
			name: 'nodeInfo',
			getter: 'bridge_nodeInfo'
		}),
		new web3._extend.Property({
			name: 'peers',
			getter: 'bridge_peers'
		}),
		new web3._extend.Property({
			name: 'chainAccount',
			getter: 'bridge_getChainAccountAddr'
		}),
		new web3._extend.Property({
			name: 'anchoringPeriod',
			getter: 'bridge_getAnchoringPeriod'
		}),
		new web3._extend.Property({
			name: 'sendChainTxslimit',
			getter: 'bridge_getSentChainTxsLimit'
		}),
		new web3._extend.Property({
			name: 'chainAccountNonce',
			getter: 'bridge_getChainAccountNonce'
		}),
		new web3._extend.Property({
			name: 'listGateway',
			getter: 'bridge_listDeployedGateway'
		}),
		new web3._extend.Property({
			name: 'txPendingCount',
			getter: 'bridge_txPendingCount'
		}),
		new web3._extend.Property({
			name: 'latestAnchoredBlockNumber',
			getter: 'bridge_getLatestAnchoredBlockNumber'
		}),
	]
});
`
const CliqueJs = `
web3._extend({
	property: 'clique',
	methods: [
		new web3._extend.Method({
			name: 'getSnapshot',
			call: 'clique_getSnapshot',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSnapshotAtHash',
			call: 'clique_getSnapshotAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'getSigners',
			call: 'clique_getSigners',
			params: 1,
			inputFormatter: [null]
		}),
		new web3._extend.Method({
			name: 'getSignersAtHash',
			call: 'clique_getSignersAtHash',
			params: 1
		}),
		new web3._extend.Method({
			name: 'propose',
			call: 'clique_propose',
			params: 2
		}),
		new web3._extend.Method({
			name: 'discard',
			call: 'clique_discard',
			params: 1
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'proposals',
			getter: 'clique_proposals'
		}),
	]
});
`
