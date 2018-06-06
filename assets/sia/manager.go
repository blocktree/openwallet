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

package sia

import (
	"github.com/shopspring/decimal"
	"github.com/tidwall/gjson"
	"time"
	"github.com/imroc/req"
	"github.com/tyler-smith/go-bip39"
)

var (
	//钱包服务API
	serverAPI = "http://127.0.0.1:10031"
	//钱包主链私钥文件路径
	walletPath = ""
	//小数位长度
	coinDecimal decimal.Decimal = decimal.NewFromFloat(100000000)
	//参与汇总的钱包
	//walletsInSum = make(map[string]*AccountBalance)
	//汇总阀值
	threshold decimal.Decimal = decimal.NewFromFloat(12).Mul(coinDecimal)
	//最小转账额度
	minSendAmount decimal.Decimal = decimal.NewFromFloat(10).Mul(coinDecimal)
	//最小矿工费
	minFees decimal.Decimal = decimal.NewFromFloat(0.005).Mul(coinDecimal)
	//汇总地址
	sumAddress = ""
	//汇总执行间隔时间
	cycleSeconds = time.Second * 10
	// 节点客户端
	client *Client
)

//GetWalletInfo 获取钱包信息
func GetWalletInfo() ([]*Wallet, error) {

	var (
		wallets = make([]*Wallet, 0)
	)

	result, err := client.Call("wallet", "GET", nil)
	if err != nil {
		return nil, err
	}

	a := gjson.ParseBytes(result)
	wallets = append(wallets, NewWallet(a))

	return wallets, err

}




//BackupWallet 备份钱包私钥数据
func BackupWallet(destination string) (string, error) {

	request := req.Param{
		"destination": destination,
	}

	_, err := client.Call("wallet/backup", "GET", request)
	if err != nil {
		return "", err
	}

	return destination, nil
}

//RestoreWallet 通过keystore恢复钱包
func RestoreWallet(keystore []byte) error {


	return nil
}

//UnlockWallet 解锁钱包
func UnlockWallet(password string) error {

	request := req.Param{
		"encryptionpassword": password,
	}

	_, err := client.Call("wallet/unlock", "POST", request)
	if err != nil {
		return err
	}

	return nil
}

//CreateNewWallet 创建钱包
func CreateNewWallet(password string, force bool) (string, error) {

	request := req.Param{
		"encryptionpassword": password,
		"force": force,
	}

	result, err := client.Call("wallet/init", "POST", request)
	if err != nil {
		return "", err
	}

	primaryseed := gjson.GetBytes(result, "seed").String()

	return primaryseed, err

}

//CreateNewWallet 创建钱包
func CreateAddress() (string, error) {

	result, err := client.Call("wallet/address", "GET", nil)
	if err != nil {
		return "", err
	}

	address := gjson.GetBytes(result, "address").String()

	return address, err

}

//GetAddressInfo 获取地址列表
func GetAddressInfo() ([]string, error) {

	result, err := client.Call("wallet/addresses", "GET", nil)
	if err != nil {
		return nil, err
	}

	content := gjson.GetBytes(result, "addresses").Array()

	addresses := make([]string, 0)
	for _, a := range content {
		addresses = append(addresses, a.String())
	}

	return addresses, err

}

func GetConsensus() error {

	_, err := client.Call("consensus", "GET", nil)
	if err != nil {
		return err
	}

	return nil
}

//genMnemonic 随机创建密钥
func genMnemonic() string {
	entropy, _ := bip39.NewEntropy(128)
	mnemonic, _ := bip39.NewMnemonic(entropy)
	return mnemonic
}