package main

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/go-OWCBasedFuncs/owkeychain"
)

func main() {
	seed := [32]byte{0x95, 0x59, 0xdb, 0xab, 0xf4, 0xd0, 0xb9, 0xf8, 0xae, 0x9a, 0x09, 0x5c, 0x93, 0x0e, 0xed, 0xe9, 0x32, 0xa5, 0x14, 0x76, 0x51, 0x86, 0xf8, 0xeb, 0x6d, 0xc3, 0x61, 0x6d, 0xcd, 0xf6, 0x68, 0xde}

	//获取比特币的扩展的根公钥
	//绝对路径为 ---  m/44'/88'/0'
	btcRootPub, err := owkeychain.GetCoinRootPublicKey(seed[:], owkeychain.Bitcoin)

	if err != nil {
		fmt.Println("fail")
	} else {
		fmt.Println("-----------------------------------------------------------------")
	}

	//将根公钥编码成owpubXXXX格式
	owkey := btcRootPub.OWEncode()
	fmt.Println("owpub格式开头的根公钥")
	fmt.Println(owkey)
	fmt.Println("-----------------------------------------------------------------")

	//通过根公钥扩展子公钥
	//扩展10个,并转为P2PK地址
	index := uint32(0)

	for ; index < 10; index++ {
		tmpKey, err := btcRootPub.DerivedPublicKeyFromSerializes(index)
		if err != nil {
			fmt.Println("fail")
		} else {
			fmt.Println("第", index+1, "个地址:", owkeychain.Base58checkEncode(tmpKey.GetPublicKeyBytes(), owkeychain.BitcoinPubkeyPrefix))

		}
	}
	fmt.Println("-----------------------------------------------------------------")

	//现在要获取第4个子公钥的私钥来进行签名，index = 3
	prikey, err := owkeychain.DerivedPrivateKeyBytes(seed[:], owkeychain.Bitcoin, uint32(3))
	if err != nil {
		fmt.Println("fail")
	} else {
		fmt.Println("第4个地址对应的私钥", hex.EncodeToString(prikey[:]))
	}

}
