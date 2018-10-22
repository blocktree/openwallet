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

package qtum

import (
	"encoding/hex"
	"fmt"
	"testing"
	"github.com/blocktree/OpenWallet/assets/qtum/btcLikeTxDriver"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
)

//案例一：
//花费指向公钥哈希地址的UTXO
//from pkh
//to   pkh
func Test_case(t *testing.T) {
	// 第一个输入 0.097
	in1 := btcLikeTxDriver.Vin{"83b27c3fd8fa2ffd8770ccfd2e8025c707a93a213ea5499c62c39d8dcb0ea625", uint32(1)}
	// 第二个输入 0.1
	in2 := btcLikeTxDriver.Vin{"e1a9c2596c38de005dccb61b737ecde7b5472c7b7791d251ec920c47913c9141", uint32(1)}

	// 目标地址与数额
	// 向 QQfTuAKdRrTawjiPZRcQ6iaK9BgxwMDgXN 发送 0.15
	// out 单位为聪
	out := btcLikeTxDriver.Vout{"QQfTuAKdRrTawjiPZRcQ6iaK9BgxwMDgXN", uint64(15000000)}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateEmptyRawTransaction([]btcLikeTxDriver.Vin{in1, in2}, []btcLikeTxDriver.Vout{out}, lockTime, replaceable)

	if err != nil {
		t.Error("构建空交易单失败")
	} else {
		fmt.Println("空交易单：")
		fmt.Println(emptyTrans)
	}

	// 获取in1 和 in2 的锁定脚本
	// 填充TxUnlock结构体
	in1Lock := "76a914f0ed48938dfa7dea31c4d12a1461b9f77560500e88ac"
	in2Lock := "76a914f3ecec22a336e205f6fbcb95ea459b6ed859a04f88ac"

	address1 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"
	address2 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"

	//针对此类指向公钥哈希地址的UTXO，此处仅需要锁定脚本即可计算待签交易单
	unlockData1 := btcLikeTxDriver.TxUnlock{nil, in1Lock, "", uint64(0),address1}
	unlockData2 := btcLikeTxDriver.TxUnlock{nil, in2Lock, "", uint64(0),address2}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("获取待签名交易单哈希失败")
	} else {
		for i, t := range transHash {
			fmt.Println("第", i+1, "个交易单哈希值为")
			fmt.Println(t)
		}
	}

	//将交易单哈希与每条哈希对应的地址发送给客户端
	//客户端根据对应地址派生私钥对哈希进行签名

	// 获取私钥
	// in1 address QiZtY5ssbVis9MntBdqmcYuJWsP5BCGBX3
	address := "L4QmwiNtobd4xTpvjxhu7mZ4oNuBhrR2kcYsqk8fS5yKZdraUNYv"
	ret, err := addressEncoder.AddressDecode(address, addressEncoder.QTUM_mainnetPrivateWIFCompressed)
	if err != nil {
		t.Error("decode error")
	} else {
		fmt.Println(ret)
	}
	in1Prikey := ret

	// in2 address Qiqk8a4ezUci9s6xeoBZHMTE1CtyjKJhNq
	address = "KxRGsMrnSRhcjmKDeajpWQXQi6agP8WiJ19djdGQ8gdWmzAsTFBe"
	ret2, err := addressEncoder.AddressDecode(address, addressEncoder.QTUM_mainnetPrivateWIFCompressed)
	if err != nil {
		t.Error("decode error")
	} else {
		fmt.Println(ret2)
	}
	in2Prikey := ret2

	//客户端使用私钥填充TxUnlock结构体，并进行签名
	//仅需要私钥，
	//此处为了测试方便，不再清除TxUnlock结构体内预填充的数据
	unlockData1.PrivateKey = in1Prikey
	unlockData2.PrivateKey = in2Prikey

	/////////交易单哈希签名
	sigPub, err := btcLikeTxDriver.SignRawTransactionHash(transHash, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})

	if err != nil {
		t.Error("交易单哈希签名失败")
	} else {
		fmt.Println("签名结果")
		for i, s := range sigPub {
			fmt.Println("第", i+1, "个签名结果")
			fmt.Println(hex.EncodeToString(s.Signature))
			fmt.Println("对应的公钥为")
			fmt.Println(hex.EncodeToString(s.Pubkey))
		}
	}

	//客户端返回签名结果和每个签名对应的公钥
	//将返回结果填充到空交易单中

	////////填充签名结果到空交易单
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例的对应字段置为 "" 即可
	signedTrans, err := btcLikeTxDriver.InsertSignatureIntoEmptyTransaction(emptyTrans, sigPub, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("交易单拼接失败")
	} else {
		fmt.Println("拼接后的交易单")
		fmt.Println(signedTrans)
	}

	/////////验证交易单
	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcLikeTxDriver.VerifyRawTransaction(signedTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if pass {
		fmt.Println("验证通过")
	} else {
		t.Error("验证失败")
	}

	///////////
	//验证通过，则可以直接发送该笔交易
}