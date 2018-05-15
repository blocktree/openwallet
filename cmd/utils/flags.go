/*
 * Copyright 2018 The OpenWallet Authors
 * This file is part of the OpenWallet library.
 *
 * The OpenWallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The OpenWallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package utils

import (
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
	"os"
	"fmt"
	"github.com/ethereum/go-ethereum/eth"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/dashboard"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"runtime"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/metrics"
	"strings"
)

const (
	VersionMajor = 1          // Major version component of the current release
	VersionMinor = 0          // Minor version component of the current release
	VersionPatch = 0          // Patch version component of the current release
	VersionMeta  = "unstable" // Version metadata to append to the version string
)

// Version holds the textual version string.
var Version = func() string {
	v := fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}()

var (
	CommandHelpTemplate = `{{.cmd.Name}}{{if .cmd.Subcommands}} command{{end}}{{if .cmd.Flags}} [command options]{{end}} [arguments...]
{{if .cmd.Description}}{{.cmd.Description}}
{{end}}{{if .cmd.Subcommands}}
SUBCOMMANDS:
	{{range .cmd.Subcommands}}{{.cmd.Name}}{{with .cmd.ShortName}}, {{.cmd}}{{end}}{{ "\t" }}{{.cmd.Usage}}
	{{end}}{{end}}{{if .categorizedFlags}}
{{range $idx, $categorized := .categorizedFlags}}{{$categorized.Name}} OPTIONS:
{{range $categorized.Flags}}{{"\t"}}{{.}}
{{end}}
{{end}}{{end}}`
)

func init() {
	cli.AppHelpTemplate = `{{.Name}} {{if .Flags}}[global options] {{end}}command{{if .Flags}} [command options]{{end}} [arguments...]

VERSION:
   {{.Version}}

COMMANDS:
   {{range .Commands}}{{.Name}}{{with .ShortName}}, {{.}}{{end}}{{ "\t" }}{{.Usage}}
   {{end}}{{if .Flags}}
GLOBAL OPTIONS:
   {{range .Flags}}{{.}}
   {{end}}{{end}}
`

	cli.CommandHelpTemplate = CommandHelpTemplate
}

// NewApp creates an app with sane defaults.
func NewApp(gitCommit, usage string) *cli.App {
	app := cli.NewApp()
	app.Name = filepath.Base(os.Args[0])
	app.Author = ""
	//app.Authors = nil
	app.Email = ""
	app.Version = Version
	if len(gitCommit) >= 8 {
		app.Version += "-" + gitCommit[:8]
	}
	app.Usage = usage
	return app
}


var (

	AppNameFlag = cli.StringFlag{
		Name: "name",
		Usage: "Application Name",
	}

	SymbolFlag = cli.StringFlag{
		Name: "symbol, s",
		Usage: "Currency symbol",
	}

	// General settings
	DataDirFlag = DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: DirectoryString{node.DefaultDataDir()},
	}
	KeyStoreDirFlag = DirectoryFlag{
		Name:  "keystore",
		Usage: "Directory for the keystore (default = inside the datadir)",
	}
	NoUSBFlag = cli.BoolFlag{
		Name:  "nousb",
		Usage: "Disables monitoring for and managing USB hardware wallets",
	}
	NetworkIdFlag = cli.Uint64Flag{
		Name:  "networkid",
		Usage: "Network identifier (integer, 1=Frontier, 2=Morden (disused), 3=Ropsten, 4=Rinkeby)",
		Value: eth.DefaultConfig.NetworkId,
	}
	TestnetFlag = cli.BoolFlag{
		Name:  "testnet",
		Usage: "Ropsten network: pre-configured proof-of-work test network",
	}
	RinkebyFlag = cli.BoolFlag{
		Name:  "rinkeby",
		Usage: "Rinkeby network: pre-configured proof-of-authority test network",
	}
	DeveloperFlag = cli.BoolFlag{
		Name:  "dev",
		Usage: "Ephemeral proof-of-authority network with a pre-funded developer account, mining enabled",
	}
	DeveloperPeriodFlag = cli.IntFlag{
		Name:  "dev.period",
		Usage: "Block period to use in developer mode (0 = mine only if transaction pending)",
	}
	IdentityFlag = cli.StringFlag{
		Name:  "identity",
		Usage: "Custom node name",
	}
	DocRootFlag = DirectoryFlag{
		Name:  "docroot",
		Usage: "Document Root for HTTPClient file scheme",
		Value: DirectoryString{homeDir()},
	}
	FastSyncFlag = cli.BoolFlag{
		Name:  "fast",
		Usage: "Enable fast syncing through state downloads",
	}
	LightModeFlag = cli.BoolFlag{
		Name:  "light",
		Usage: "Enable light client mode",
	}
	defaultSyncMode = eth.DefaultConfig.SyncMode
	SyncModeFlag    = TextMarshalerFlag{
		Name:  "syncmode",
		Usage: `Blockchain sync mode ("fast", "full", or "light")`,
		Value: &defaultSyncMode,
	}
	GCModeFlag = cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("full", "archive")`,
		Value: "full",
	}
	LightServFlag = cli.IntFlag{
		Name:  "lightserv",
		Usage: "Maximum percentage of time allowed for serving LES requests (0-90)",
		Value: 0,
	}
	LightPeersFlag = cli.IntFlag{
		Name:  "lightpeers",
		Usage: "Maximum number of LES client peers",
		Value: eth.DefaultConfig.LightPeers,
	}
	LightKDFFlag = cli.BoolFlag{
		Name:  "lightkdf",
		Usage: "Reduce key-derivation RAM & CPU usage at some expense of KDF strength",
	}
	// Dashboard settings
	DashboardEnabledFlag = cli.BoolFlag{
		Name:  "dashboard",
		Usage: "Enable the dashboard",
	}
	DashboardAddrFlag = cli.StringFlag{
		Name:  "dashboard.addr",
		Usage: "Dashboard listening interface",
		Value: dashboard.DefaultConfig.Host,
	}
	DashboardPortFlag = cli.IntFlag{
		Name:  "dashboard.host",
		Usage: "Dashboard listening port",
		Value: dashboard.DefaultConfig.Port,
	}
	DashboardRefreshFlag = cli.DurationFlag{
		Name:  "dashboard.refresh",
		Usage: "Dashboard metrics collection refresh rate",
		Value: dashboard.DefaultConfig.Refresh,
	}
	// Ethash settings
	EthashCacheDirFlag = DirectoryFlag{
		Name:  "ethash.cachedir",
		Usage: "Directory to store the ethash verification caches (default = inside the datadir)",
	}
	EthashCachesInMemoryFlag = cli.IntFlag{
		Name:  "ethash.cachesinmem",
		Usage: "Number of recent ethash caches to keep in memory (16MB each)",
		Value: eth.DefaultConfig.Ethash.CachesInMem,
	}
	EthashCachesOnDiskFlag = cli.IntFlag{
		Name:  "ethash.cachesondisk",
		Usage: "Number of recent ethash caches to keep on disk (16MB each)",
		Value: eth.DefaultConfig.Ethash.CachesOnDisk,
	}
	EthashDatasetDirFlag = DirectoryFlag{
		Name:  "ethash.dagdir",
		Usage: "Directory to store the ethash mining DAGs (default = inside home folder)",
		Value: DirectoryString{eth.DefaultConfig.Ethash.DatasetDir},
	}
	EthashDatasetsInMemoryFlag = cli.IntFlag{
		Name:  "ethash.dagsinmem",
		Usage: "Number of recent ethash mining DAGs to keep in memory (1+GB each)",
		Value: eth.DefaultConfig.Ethash.DatasetsInMem,
	}
	EthashDatasetsOnDiskFlag = cli.IntFlag{
		Name:  "ethash.dagsondisk",
		Usage: "Number of recent ethash mining DAGs to keep on disk (1+GB each)",
		Value: eth.DefaultConfig.Ethash.DatasetsOnDisk,
	}
	// Transaction pool settings
	TxPoolNoLocalsFlag = cli.BoolFlag{
		Name:  "txpool.nolocals",
		Usage: "Disables price exemptions for locally submitted transactions",
	}
	TxPoolJournalFlag = cli.StringFlag{
		Name:  "txpool.journal",
		Usage: "Disk journal for local transaction to survive node restarts",
		Value: core.DefaultTxPoolConfig.Journal,
	}
	TxPoolRejournalFlag = cli.DurationFlag{
		Name:  "txpool.rejournal",
		Usage: "Time interval to regenerate the local transaction journal",
		Value: core.DefaultTxPoolConfig.Rejournal,
	}
	TxPoolPriceLimitFlag = cli.Uint64Flag{
		Name:  "txpool.pricelimit",
		Usage: "Minimum gas price limit to enforce for acceptance into the pool",
		Value: eth.DefaultConfig.TxPool.PriceLimit,
	}
	TxPoolPriceBumpFlag = cli.Uint64Flag{
		Name:  "txpool.pricebump",
		Usage: "Price bump percentage to replace an already existing transaction",
		Value: eth.DefaultConfig.TxPool.PriceBump,
	}
	TxPoolAccountSlotsFlag = cli.Uint64Flag{
		Name:  "txpool.accountslots",
		Usage: "Minimum number of executable transaction slots guaranteed per account",
		Value: eth.DefaultConfig.TxPool.AccountSlots,
	}
	TxPoolGlobalSlotsFlag = cli.Uint64Flag{
		Name:  "txpool.globalslots",
		Usage: "Maximum number of executable transaction slots for all accounts",
		Value: eth.DefaultConfig.TxPool.GlobalSlots,
	}
	TxPoolAccountQueueFlag = cli.Uint64Flag{
		Name:  "txpool.accountqueue",
		Usage: "Maximum number of non-executable transaction slots permitted per account",
		Value: eth.DefaultConfig.TxPool.AccountQueue,
	}
	TxPoolGlobalQueueFlag = cli.Uint64Flag{
		Name:  "txpool.globalqueue",
		Usage: "Maximum number of non-executable transaction slots for all accounts",
		Value: eth.DefaultConfig.TxPool.GlobalQueue,
	}
	TxPoolLifetimeFlag = cli.DurationFlag{
		Name:  "txpool.lifetime",
		Usage: "Maximum amount of time non-executable transaction are queued",
		Value: eth.DefaultConfig.TxPool.Lifetime,
	}
	// Performance tuning settings
	CacheFlag = cli.IntFlag{
		Name:  "cache",
		Usage: "Megabytes of memory allocated to internal caching",
		Value: 1024,
	}
	CacheDatabaseFlag = cli.IntFlag{
		Name:  "cache.database",
		Usage: "Percentage of cache memory allowance to use for database io",
		Value: 75,
	}
	CacheGCFlag = cli.IntFlag{
		Name:  "cache.gc",
		Usage: "Percentage of cache memory allowance to use for trie pruning",
		Value: 25,
	}
	TrieCacheGenFlag = cli.IntFlag{
		Name:  "trie-cache-gens",
		Usage: "Number of trie node generations to keep in memory",
		Value: int(state.MaxTrieCacheGen),
	}
	// Miner settings
	MiningEnabledFlag = cli.BoolFlag{
		Name:  "mine",
		Usage: "Enable mining",
	}
	MinerThreadsFlag = cli.IntFlag{
		Name:  "minerthreads",
		Usage: "Number of CPU threads to use for mining",
		Value: runtime.NumCPU(),
	}
	TargetGasLimitFlag = cli.Uint64Flag{
		Name:  "targetgaslimit",
		Usage: "Target gas limit sets the artificial target gas floor for the blocks to mine",
		Value: params.GenesisGasLimit,
	}
	EtherbaseFlag = cli.StringFlag{
		Name:  "etherbase",
		Usage: "Public address for block mining rewards (default = first account created)",
		Value: "0",
	}
	GasPriceFlag = BigFlag{
		Name:  "gasprice",
		Usage: "Minimal gas price to accept for mining a transactions",
		Value: eth.DefaultConfig.GasPrice,
	}
	ExtraDataFlag = cli.StringFlag{
		Name:  "extradata",
		Usage: "Block extra data set by the miner (default = client version)",
	}
	// Account settings
	UnlockedAccountFlag = cli.StringFlag{
		Name:  "unlock",
		Usage: "Comma separated list of accounts to unlock",
		Value: "",
	}
	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use for non-interactive password input",
		Value: "",
	}

	VMEnableDebugFlag = cli.BoolFlag{
		Name:  "vmdebug",
		Usage: "Record information useful for VM and contract debugging",
	}
	// Logging and debug settings
	EthStatsURLFlag = cli.StringFlag{
		Name:  "ethstats",
		Usage: "Reporting URL of a ethstats service (nodename:secret@host:port)",
	}
	MetricsEnabledFlag = cli.BoolFlag{
		Name:  metrics.MetricsEnabledFlag,
		Usage: "Enable metrics collection and reporting",
	}
	FakePoWFlag = cli.BoolFlag{
		Name:  "fakepow",
		Usage: "Disables proof-of-work verification",
	}
	NoCompactionFlag = cli.BoolFlag{
		Name:  "nocompaction",
		Usage: "Disables db compaction after import",
	}
	// RPC settings
	RPCEnabledFlag = cli.BoolFlag{
		Name:  "rpc",
		Usage: "Enable the HTTP-RPC server",
	}
	RPCListenAddrFlag = cli.StringFlag{
		Name:  "rpcaddr",
		Usage: "HTTP-RPC server listening interface",
		Value: node.DefaultHTTPHost,
	}
	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: node.DefaultHTTPPort,
	}
	RPCCORSDomainFlag = cli.StringFlag{
		Name:  "rpccorsdomain",
		Usage: "Comma separated list of domains from which to accept cross origin requests (browser enforced)",
		Value: "",
	}
	RPCVirtualHostsFlag = cli.StringFlag{
		Name:  "rpcvhosts",
		Usage: "Comma separated list of virtual hostnames from which to accept requests (server enforced). Accepts '*' wildcard.",
		Value: strings.Join(node.DefaultConfig.HTTPVirtualHosts, ","),
	}
	RPCApiFlag = cli.StringFlag{
		Name:  "rpcapi",
		Usage: "API's offered over the HTTP-RPC interface",
		Value: "",
	}
	IPCDisabledFlag = cli.BoolFlag{
		Name:  "ipcdisable",
		Usage: "Disable the IPC-RPC server",
	}
	IPCPathFlag = DirectoryFlag{
		Name:  "ipcpath",
		Usage: "Filename for IPC socket/pipe within the datadir (explicit paths escape it)",
	}
	WSEnabledFlag = cli.BoolFlag{
		Name:  "ws",
		Usage: "Enable the WS-RPC server",
	}
	WSListenAddrFlag = cli.StringFlag{
		Name:  "wsaddr",
		Usage: "WS-RPC server listening interface",
		Value: node.DefaultWSHost,
	}
	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: node.DefaultWSPort,
	}
	WSApiFlag = cli.StringFlag{
		Name:  "wsapi",
		Usage: "API's offered over the WS-RPC interface",
		Value: "",
	}
	WSAllowedOriginsFlag = cli.StringFlag{
		Name:  "wsorigins",
		Usage: "Origins from which to accept websockets requests",
		Value: "",
	}
	ExecFlag = cli.StringFlag{
		Name:  "exec",
		Usage: "Execute JavaScript statement",
	}
	PreloadJSFlag = cli.StringFlag{
		Name:  "preload",
		Usage: "Comma separated list of JavaScript files to preload into the console",
	}

	// Network Settings
	MaxPeersFlag = cli.IntFlag{
		Name:  "maxpeers",
		Usage: "Maximum number of network peers (network disabled if set to 0)",
		Value: 25,
	}
	MaxPendingPeersFlag = cli.IntFlag{
		Name:  "maxpendpeers",
		Usage: "Maximum number of pending connection attempts (defaults used if set to 0)",
		Value: 0,
	}
	ListenPortFlag = cli.IntFlag{
		Name:  "port",
		Usage: "Network listening port",
		Value: 30303,
	}
	BootnodesFlag = cli.StringFlag{
		Name:  "bootnodes",
		Usage: "Comma separated enode URLs for P2P discovery bootstrap (set v4+v5 instead for light servers)",
		Value: "",
	}
	BootnodesV4Flag = cli.StringFlag{
		Name:  "bootnodesv4",
		Usage: "Comma separated enode URLs for P2P v4 discovery bootstrap (light server, full nodes)",
		Value: "",
	}
	BootnodesV5Flag = cli.StringFlag{
		Name:  "bootnodesv5",
		Usage: "Comma separated enode URLs for P2P v5 discovery bootstrap (light server, light nodes)",
		Value: "",
	}
	NodeKeyFileFlag = cli.StringFlag{
		Name:  "nodekey",
		Usage: "P2P node key file",
	}
	NodeKeyHexFlag = cli.StringFlag{
		Name:  "nodekeyhex",
		Usage: "P2P node key as hex (for testing)",
	}
	NATFlag = cli.StringFlag{
		Name:  "nat",
		Usage: "NAT port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
		Value: "any",
	}
	NoDiscoverFlag = cli.BoolFlag{
		Name:  "nodiscover",
		Usage: "Disables the peer discovery mechanism (manual peer addition)",
	}
	DiscoveryV5Flag = cli.BoolFlag{
		Name:  "v5disc",
		Usage: "Enables the experimental RLPx V5 (Topic Discovery) mechanism",
	}
	NetrestrictFlag = cli.StringFlag{
		Name:  "netrestrict",
		Usage: "Restricts network communication to the given IP networks (CIDR masks)",
	}

	// ATM the url is left to the user and deployment to
	JSpathFlag = cli.StringFlag{
		Name:  "jspath",
		Usage: "JavaScript root path for `loadScript`",
		Value: ".",
	}

)