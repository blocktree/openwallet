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
	"github.com/blocktree/go-OWCBasedFuncs/btcLikeTxDriver"
	"github.com/blocktree/go-OWCBasedFuncs/addressEncoder"
)

//案例一：
//花费指向公钥哈希地址的UTXO
//from pkh
//to   pkh
func Test_case1(t *testing.T) {
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

//案例二
//花费指向脚本哈希地址的UTXO
// from SH
// to   PKH
func Test_case2(t *testing.T) {
	// 第一个输入
	in1 := btcLikeTxDriver.Vin{"871b7a30ab16f6bdd1a27de40713249e395bde029796289044a9c26fe16e6e0a", uint32(0)}
	// 第二个输入
	in2 := btcLikeTxDriver.Vin{"53c40f36fd44c11cc9381d9eac6b0817678180200b90613131c80b5655754969", uint32(0)}

	// 目标地址与数额
	// 向 QMuCdu2y2tjQw8KtFYiL5zfMDpYZCHSz6k 发送 0.13
	// out 单位为聪
	out := btcLikeTxDriver.Vout{"QMuCdu2y2tjQw8KtFYiL5zfMDpYZCHSz6k", uint64(13000000)}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	///////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateEmptyRawTransaction([]btcLikeTxDriver.Vin{in1, in2}, []btcLikeTxDriver.Vout{out}, lockTime, replaceable)

	if err != nil {
		t.Error("构建空交易单失败")
	} else {
		fmt.Println("空交易单：")
		fmt.Println(emptyTrans)
	}

	// 获取in1 和 in2 的锁定脚本
	// 获取in1 和 in2 的赎回脚本
	// 获取in1 和 in2 的数额amount
	// 填充TxUnlock结构体
	in1Lock := "a91421f5946fcec43caa5d905d6e7c4d34aad57e20b387"
	in2Lock := "a914cfed6d5ae483deda33e92eb4c96fc1a281fbe06487"
	in1Redeem := "0014a972da7198dadfa4fb8886091d523a64a9e95a88"
	in2Redeem := "001469b6968a3d6917d0e1270b0b21d3605b86f3dee5"
	in1Amount := uint64(17411199)
	in2Amount := uint64(5559614)
	address1 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"
	address2 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"

	//针对此类指向脚本哈希地址的UTXO，此需要锁定脚本、赎回脚本以及该UTXO包含的数额方可计算待签交易单
	unlockData1 := btcLikeTxDriver.TxUnlock{nil, in1Lock, in1Redeem, in1Amount,address1}
	unlockData2 := btcLikeTxDriver.TxUnlock{nil, in2Lock, in2Redeem, in2Amount,address2}

	/////////计算待签名交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("创建待签交易单哈希失败")
	} else {
		for i, t := range transHash {
			fmt.Println("第", i+1, "个交易单哈希值为")
			fmt.Println(t)
		}
	}

	//将交易单哈希与每条哈希对应的地址发送给客户端
	//客户端根据对应地址派生私钥对哈希进行签名

	// 获取私钥
	// in1 address 2MvLnUoMyYmfxCqSbh7tpGpTxj18UPCvRqb
	//5ae604201f2e7dc1004603dd231a8c922b92abe836fe48eddaafe9320fa0cc60
	in1Prikey := []byte{0x5a, 0xe6, 0x04, 0x20, 0x1f, 0x2e, 0x7d, 0xc1, 0x00, 0x46, 0x03, 0xdd, 0x23, 0x1a, 0x8c, 0x92, 0x2b, 0x92, 0xab, 0xe8, 0x36, 0xfe, 0x48, 0xed, 0xda, 0xaf, 0xe9, 0x32, 0x0f, 0xa0, 0xcc, 0x60}
	// in2 address 2NCCeHip41kqwNJwopWmwqxrgM3VJiGDCsx   d25cf4f6744013b7421966de5e27e1fdc82f465fc863a751441dd74e5ace02bc
	in2Prikey := []byte{0xd2, 0x5c, 0xf4, 0xf6, 0x74, 0x40, 0x13, 0xb7, 0x42, 0x19, 0x66, 0xde, 0x5e, 0x27, 0xe1, 0xfd, 0xc8, 0x2f, 0x46, 0x5f, 0xc8, 0x63, 0xa7, 0x51, 0x44, 0x1d, 0xd7, 0x4e, 0x5a, 0xce, 0x02, 0xbc}

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
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例中需要设置
	//  前面已经设置，此处不再重复操作
	signedTrans, err := btcLikeTxDriver.InsertSignatureIntoEmptyTransaction(emptyTrans, sigPub, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("交易单拼接失败")
	} else {
		fmt.Println("拼接后的交易单")
		fmt.Println(signedTrans)
	}

	//交易单验证

	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcLikeTxDriver.VerifyRawTransaction(signedTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if pass {
		fmt.Println("验证通过")
	} else {
		t.Error("验证失败")
	}
}

//案例三
//花费混合了指向公钥哈希和脚本哈希的UTXO
// from pkh + sh
// to pkh
func Test_case3(t *testing.T) {

	//输入一
	//指向公钥哈希地址的UTXO
	in1 := btcLikeTxDriver.Vin{"302759ff352b436db5b9c1700d6a1e5f29c324a9e3d69190b65b3553e05c9308", uint32(0)}
	//输入二
	//指向脚本哈希地址的UTXO
	in2 := btcLikeTxDriver.Vin{"184d6c95f2d4c394f7ff63ce3388a65e8daa182351f64bd69abd64ac9fc51a23", uint32(1)}

	//输出
	// 向 mwmXzRM19gg5AB5Vu16dvfuhWujTq5PzvK 发送 0.673
	// out 单位为聪
	out := btcLikeTxDriver.Vout{"mwmXzRM19gg5AB5Vu16dvfuhWujTq5PzvK", uint64(67300000)}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	///////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateEmptyRawTransaction([]btcLikeTxDriver.Vin{in1, in2}, []btcLikeTxDriver.Vout{out}, lockTime, replaceable)

	if err != nil {
		t.Error("构建空交易单失败")
	} else {
		fmt.Println("空交易单：")
		fmt.Println(emptyTrans)
	}

	// 获取in1 和 in2 的锁定脚本
	// 获取in2 的赎回脚本
	// 获取in1 和 in2 的数额amount
	// 填充TxUnlock结构体
	in1Lock := "76a914d46043209073ad39879356295562d952cd9dae3a88ac"
	in2Lock := "a914bd52d5e36ab0cbd34d981843d1b620705a67927d87"
	in1Redeem := ""
	in2Redeem := "0014774334d4657dc2d251eff89af58d0a92bde2ec05"
	in1Amount := uint64(0)
	in2Amount := uint64(823237)
	address1 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"
	address2 := "mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq"

	//针对此类指向脚本哈希地址的UTXO，此需要锁定脚本、赎回脚本以及该UTXO包含的数额方可计算待签交易单
	unlockData1 := btcLikeTxDriver.TxUnlock{nil, in1Lock, in1Redeem, in1Amount,address1}
	unlockData2 := btcLikeTxDriver.TxUnlock{nil, in2Lock, in2Redeem, in2Amount,address2}

	/////////计算待签名交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("创建待签交易单哈希失败")
	} else {
		for i, t := range transHash {
			fmt.Println("第", i+1, "个交易单哈希值为")
			fmt.Println(t)
		}
	}

	//将交易单哈希与每条哈希对应的地址发送给客户端
	//客户端根据对应地址派生私钥对哈希进行签名

	// 获取私钥
	// in1 address mzsts8xiVWv8uGEYUrAB6XzKXZPiX9j6jq
	//80bc398d7c4a674daa977566c2e6cd5040520027e57fe806dfaa868df4cc43ab
	in1Prikey := []byte{0x80, 0xbc, 0x39, 0x8d, 0x7c, 0x4a, 0x67, 0x4d, 0xaa, 0x97, 0x75, 0x66, 0xc2, 0xe6, 0xcd, 0x50, 0x40, 0x52, 0x00, 0x27, 0xe5, 0x7f, 0xe8, 0x06, 0xdf, 0xaa, 0x86, 0x8d, 0xf4, 0xcc, 0x43, 0xab}
	// in2 address 2NAWGw3wHZTnHnXRT1GZF2eB6a6DUqqHCu8
	// 7e41219c19076595dc17421ffd894188354d151539c1be1c15e975fcc9c47789
	in2Prikey := []byte{0x7e, 0x41, 0x21, 0x9c, 0x19, 0x07, 0x65, 0x95, 0xdc, 0x17, 0x42, 0x1f, 0xfd, 0x89, 0x41, 0x88, 0x35, 0x4d, 0x15, 0x15, 0x39, 0xc1, 0xbe, 0x1c, 0x15, 0xe9, 0x75, 0xfc, 0xc9, 0xc4, 0x77, 0x89}

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
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例中需要设置
	//  前面已经设置，此处不再重复操作
	signedTrans, err := btcLikeTxDriver.InsertSignatureIntoEmptyTransaction(emptyTrans, sigPub, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if err != nil {
		t.Error("交易单拼接失败")
	} else {
		fmt.Println("拼接后的交易单")
		fmt.Println(signedTrans)
	}

	//交易单验证

	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcLikeTxDriver.VerifyRawTransaction(signedTrans, []btcLikeTxDriver.TxUnlock{unlockData1, unlockData2})
	if pass {
		fmt.Println("验证通过")
	} else {
		t.Error("验证失败")
	}
}
