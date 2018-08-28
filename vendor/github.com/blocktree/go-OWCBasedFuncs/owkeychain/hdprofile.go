package owkeychain

import "github.com/blocktree/go-OWCrypt"

var (
	openwalletPrePath = "m/44'/88'"
)

type CoinType struct {
	hdIndex   uint32
	curveType uint32
}

//XXX[0]:hd扩展索引
//XXX[1]:曲线类型
var (
	Bitcoin  = CoinType{uint32(0), owcrypt.ECC_CURVE_SECP256K1}
	Ethereum = CoinType{uint32(1), owcrypt.ECC_CURVE_SECP256K1}
)

var (
	owprvPrefix = []byte{0x07, 0xa8, 0x10, 0x0c, 0x28}
	owpubPrefix = []byte{0x07, 0xa8, 0x10, 0x31, 0xa2}

	BitcoinPubkeyPrefix = []byte{0}
	BitcoinScriptPrefix = []byte{5}
)
