module github.com/blocktree/openwallet/v2

go 1.23.1

toolchain go1.23.4

require (
	docker.io/go-docker v1.0.0
	github.com/NebulousLabs/entropy-mnemonics v0.0.0-20181203154559-bc7e13c5ccd8
	github.com/asdine/storm v2.1.2+incompatible
	github.com/astaxie/beego v1.12.0
	github.com/awnumar/memguard v0.22.5
	github.com/blocktree/go-owcdrivers v1.2.0
	github.com/blocktree/go-owcrypt v1.1.14
	github.com/bndr/gotabulate v1.1.2
	github.com/bradfitz/gomemcache v0.0.0-20190913173617-a41fca850d0b
	github.com/btcsuite/btcd/btcutil v1.1.0
	github.com/bwmarrin/snowflake v0.3.0
	github.com/codeskyblue/go-sh v0.0.0-20190412065543-76bd3d59ff27
	github.com/couchbase/go-couchbase v0.0.0-20191217190632-b2754d72cc98
	github.com/docker/go-connections v0.4.0
	github.com/ethereum/go-ethereum v1.10.17
	github.com/go-redis/redis v6.15.6+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/gorilla/websocket v1.4.2
	github.com/imroc/req v0.2.4
	github.com/lib/pq v1.3.0
	github.com/mr-tron/base58 v1.1.3
	github.com/pborman/uuid v1.2.0
	github.com/peterh/liner v1.1.1-0.20190123174540-a2c9a5303de7
	github.com/pkg/errors v0.9.1
	github.com/shopspring/decimal v0.0.0-20200105231215-408a2507e114
	github.com/siddontang/ledisdb v0.0.0-20190202134119-8ceb77e66a92
	github.com/ssdb/gossdb v0.0.0-20180723034631-88f6b59b84ec
	github.com/streadway/amqp v0.0.0-20190827072141-edfb9018d271
	github.com/tidwall/gjson v1.9.3
	github.com/tyler-smith/go-bip39 v1.0.2
	go.etcd.io/bbolt v1.3.3
	golang.org/x/crypto v0.33.0
	gopkg.in/urfave/cli.v1 v1.20.0
)

require (
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/awnumar/memcall v0.4.0 // indirect
	github.com/btcsuite/btcd v0.23.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.1.3 // indirect
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1 // indirect
	github.com/codegangsta/inject v0.0.0-20150114235600-33e0aa1cb7c0 // indirect
	github.com/couchbase/gomemcached v0.0.0-20181122193126-5125a94a666c // indirect
	github.com/couchbase/goutils v0.0.0-20180530154633-e865a1461c8a // indirect
	github.com/cupcake/rdb v0.0.0-20161107195141-43ba34106c76 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.0.1 // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/drand/kyber v1.1.4 // indirect
	github.com/edsrzf/mmap-go v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/mattn/go-runewidth v0.0.9 // indirect
	github.com/opencontainers/go-digest v1.0.0-rc1 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pelletier/go-toml v1.2.0 // indirect
	github.com/phoreproject/bls v0.0.0-20200525203911-a88a5ae26844 // indirect
	github.com/shiena/ansicolor v0.0.0-20151119151921-a422bbe96644 // indirect
	github.com/siddontang/go v0.0.0-20180604090527-bdc77568d726 // indirect
	github.com/siddontang/rdb v0.0.0-20150307021120-fc89ed2e418d // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7 // indirect
	github.com/tidwall/match v1.1.1 // indirect
	github.com/tidwall/pretty v1.2.0 // indirect
	golang.org/x/net v0.35.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
)

//replace github.com/blocktree/go-owcdrivers => ../go-owcdrivers
//
//replace github.com/blocktree/go-owcrypt => ../go-owcrypt
