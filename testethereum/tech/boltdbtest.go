package tech

import (
	"sync"
	"time"

	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/hdkeystore"
	"github.com/btcsuite/btcutil/hdkeychain"
)

type Coin struct {
	Symbol     string `json:"symbol"`
	IsContract bool   `json:"isContract"`
	ContractID string `json:"contractID"` //代币的编号, 后续会定义, 目前先不管
}

type Address struct {
	AccountID string    `json:"accountID" storm:"index"` //钱包ID
	Address   string    `json:"address" storm:"id"`      //地址字符串
	Alias     string    `json:"alias"`                   //地址别名，可绑定用户
	Tag       string    `json:"tag"`                     //标签
	Index     uint64    `json:"index"`                   //账户ID，索引位
	HDPath    string    `json:"hdPath"`                  //地址公钥根路径
	WatchOnly bool      `json:"watchOnly"`               //是否观察地址，true的时候，Index，RootPath，Alias都没有。
	Symbol    string    `json:"symbol"`                  //币种类别
	Balance   string    `json:"balance"`                 //余额
	IsMemo    bool      `json:"isMemo"`                  //是否备注
	Memo      string    `json:"memo"`                    //备注
	CreatedAt time.Time `json:"createdAt"`               //创建时间
	IsChange  bool      `json:"isChange"`                //是否找零地址

	//核心地址指针
	Core interface{}
}

// 1. 地址也有别名?
// 2. 地址标签是什么?index? HDPath应该是APP端维护吧? 观察地址是什么? 备注是谁来备注?
// 3. 以太坊不需要找零.
// 4. balance是节点来赋值么? 哪一步来赋值?
// 5. 钱包里有publickey, 地址是不是应该也有, 验签要用

type KeySignature struct {
	EccType    uint32   //曲线类型
	Address    *Address //提供签名的地址
	Signatures string   //未花签名
	Message    string   //被签消息
}

// 1. RawTransaction中的RawHex是不是这个Message

type AssetsAccount struct {
	WalletID  string   `json:"walletID"`             //钱包ID
	Alias     string   `json:"alias"`                //别名
	AccountID string   `json:"accountID" storm:"id"` //账户ID，合成地址
	Index     uint64   `json:"index"`                //账户ID，索引位
	HDPath    string   `json:"hdPath"`               //衍生路径
	PublicKey string   `json:"publicKey"`            //主公钥
	OwnerKeys []string `json:"ownerKeys"`            //公钥数组，大于1为多签
	//Owners          map[string]AccountOwner //拥有者列表, 账户公钥: 拥有者
	ContractAddress string      `json:"contractAddress"` //多签合约地址
	Required        uint64      `json:"required"`        //必要签名数
	Symbol          string      `json:"symbol"`          //资产币种类别
	AddressIndex    int         `json:"addressCount"`
	Balance         string      `json:"balance"`
	core            interface{} //核心账户指针
}

type RawTransaction struct {
	Coin        Coin                      //-指明以太坊交 智能合约地址
	TxID        string                    //
	RawHex      string                    //区块链协议构造的交易原生数据
	Amount      string                    //
	FeeRate     string                    //gas price
	To          []string                  //
	Account     *AssetsAccount            //对象中包含了WalletWrapper用于查询db, 获取账户地址的一个
	Signatures  map[string][]KeySignature //传入CreateRawTransaction的时候, map为空; 由CreateRawTransaction对KeySignature进行初始化
	Required    uint64                    //必要签名
	IsBuilt     bool                      //是否完成构建建议单
	IsCompleted bool                      //是否完成所有签名
	IsSubmit    bool                      //是否已广播
}

// 1. FeeRate是gas limit还是gas price?
// 2. 以太坊是否会有多个目的地址?
// 3. Account就是钱包, 找到钱包中的一个地址作为from地址, 地址存在哪里?
// 4. 智能合约只会调用代币的转账接口吧?
// 5. Required, 有不需要签名的情况么?
// 6. 缺少from, nonce, input, chainID
// 7. TransactionDecoder是多线程调用的么? 钱包的余额更新的问题.
// 8. address里应该有公钥吧

/*
{
  raw: "0xf8ab81ad850430e234008301d8a8948847e5f841458ace82dbb0692c97115799fe28d380b844a9059cbb0000000000000000000000002a63b2203955b84fefe52baca3881b3614991b3400000000000000000000000000000000000000000000000000000000000000643ca0f8b6f37211c43ef90a76972e9b3fba42e09ff12dc7e2c73667a12164086a7994a0123c98401fe09d352d2cb4e0bc0b0aafc86e3dd7ca169e14561bfdf2c3d16040",
  tx: {
    gas: "0x1d8a8",
    gasPrice: "0x430e23400",
    hash: "0x6c1dfbb6bd00c08819e75f0b1240dd5251b6ae4f0fdc291729828a9adb236536",
    input: "0xa9059cbb0000000000000000000000002a63b2203955b84fefe52baca3881b3614991b340000000000000000000000000000000000000000000000000000000000000064",
	nonce: "0xad",
	to: "0x8847e5f841458ace82dbb0692c97115799fe28d3",
	value: "0x0",

    r: "0xf8b6f37211c43ef90a76972e9b3fba42e09ff12dc7e2c73667a12164086a7994",
    s: "0x123c98401fe09d352d2cb4e0bc0b0aafc86e3dd7ca169e14561bfdf2c3d16040",
    v: "0x3c"
  }

*/

type StormDB struct {
	*storm.DB
	FileName string
	Opened   bool
}
type Wrapper struct {
	sourceDB     *StormDB     //存储钱包相关数据的数据库，目前使用boltdb作为持久方案
	mu           sync.RWMutex //锁
	isExternalDB bool         //是否外部加载的数据库，非内部打开，内部打开需要关闭
	sourceFile   string       //钱包数据库文件路径，用于内部打开
}

type unlocked struct {
	Key   *hdkeychain.ExtendedKey
	abort chan struct{}
}

type Wallet struct {
	AppID        string `json:"appID"`
	WalletID     string `json:"walletID"  storm:"id"`
	Alias        string `json:"alias"`
	Password     string `json:"password"`
	RootPub      string `json:"rootpub"` //弃用
	RootPath     string `json:"rootPath"`
	KeyFile      string `json:"keyFile"`      //钱包的密钥文件
	DBFile       string `json:"dbFile"`       //钱包的数据库文件
	WatchOnly    bool   `json:"watchOnly"`    //创建watchonly的钱包，没有私钥文件，只有db文件
	IsTrust      bool   `json:"isTrust"`      //是否托管密钥
	AccountIndex int    `json:"accountIndex"` //账户索引数，-1代表未创建账户

	key      *hdkeystore.HDKey
	fileName string              //钱包文件命名，所有与钱包相关的都以这个filename命名
	core     interface{}         //核心钱包指针
	unlocked map[string]unlocked // 已解锁的钱包，集合（钱包地址, 钱包私钥）
}

type WalletWrapper struct {
	*AppWrapper
	wallet  *Wallet //需要包装的钱包
	keyFile string  //钱包密钥文件路径

}

type AppWrapper struct {
	*Wrapper
	appID string
}
