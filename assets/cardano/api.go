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

package cardano

import (
	"crypto/tls"
	"fmt"
	"github.com/imroc/req"
	"log"
	"net/http"
	"strings"
)

const (
	serverAPI = "https://192.168.2.224:10026/api/"
)

var (
	api = req.New()
)

func init() {

	//改http
	trans, _ := api.Client().Transport.(*http.Transport)
	trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	//开启调试
	req.Debug = false
}

//callCreateWalletAPI 调用创建钱包接口
func callCreateWalletAPI(name, words string, passphrase ...string) []byte {

	var (
		//result         map[string]interface{}
		cwInitMeta     = make(map[string]interface{}, 0)
		body           = make(map[string]interface{}, 0)
		cwBackupPhrase = make(map[string]interface{}, 0)
	)

	method := "wallets/new"
	url := serverAPI + method
	param := make(req.QueryParam, 0)
	if len(passphrase) > 0 {
		//设置密码
		param["passphrase"] = passphrase[0]
	}

	//分割助记词
	mnemonicArray := strings.Split(words, " ")
	cwBackupPhrase["bpToList"] = mnemonicArray

	cwInitMeta["cwName"] = name
	cwInitMeta["cwAssurance"] = "CWANormal"
	cwInitMeta["cwUnit"] = 0

	body["cwInitMeta"] = cwInitMeta
	body["cwBackupPhrase"] = cwBackupPhrase

	log.Println("开始请求钱包API")

	r, err := api.Post(url, param, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return nil
	}
	//r.ToJSON(&result)
	log.Printf("%v\n", r)
	return r.Bytes()
}

//GetWalletInfo 查询钱包信息，因ADA的接口限制，只接收1个wid
func callGetWalletAPI(wid ...string) []byte {

	var (
	//result map[string]interface{}
	)

	method := "wallets"
	url := serverAPI + method
	param := make(req.Param, 0)
	if len(wid) > 0 && len(wid[0]) > 0 {
		//查询wid的钱包信息
		url = fmt.Sprint(url, "/", wid[0])
	} else {
		//查询全部钱包信息

	}
	r, err := api.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	log.Printf("%+v\n", r)

	/*
		//success 返回结果
		{
			"Right": [{
				"cwId": "Ae2tdPwUPEZ2CVLXYWiatEb2aSwaR573k4NY581fMde9N2GiCqtL7h6ybhU",
				"cwMeta": {
					"cwName": "Personal Wallet 1",
					"cwAssurance": "CWANormal",
					"cwUnit": 0
				},
				"cwAccountsNumber": 3,
				"cwAmount": {
					"getCCoin": "3829106"
				},
				"cwHasPassphrase": false,
				"cwPassphraseLU": 1.517391540346574584e9
			}, ...]
		}
	*/

	//r.ToJSON(&result)

	return r.Bytes()
}

// callCreateNewAccountAPI 调用创建账户API，传入账户名，钱包id
func callCreateNewAccountAPI(name, wid, passphrase string) []byte {
	var (
		//result map[string]interface{}
		body  map[string]interface{}
		param = make(req.QueryParam, 0)
	)

	method := "accounts"
	url := serverAPI + method

	/*
		传入参数
		{
			"caInitMeta": {
				"caName": "string"
			},
			"caInitWId": "string"
		}
	*/

	body = map[string]interface{}{
		"caInitMeta": map[string]interface{}{
			"caName": name,
		},
		"caInitWId": wid,
	}

	//设置密码
	param["passphrase"] = passphrase

	r, err := api.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil
	}

	log.Printf("%+v\n", r)

	/*
		success 返回结果
		{
			"Right": {
				"caId": "Ae2tdPwUPEZKrS6hL1E9XgKSet8ydFYHQkV3gmEG8RGUwh5XKugoUFEk7Lx@2237891267",
				"caMeta": {
					"caName": "chance"
				},
				"caAddresses": [{
					"cadId": "DdzFFzCqrht5DnfoXe47MEs12Tkhzd9JHkZ1f1QDMy1FEFsZd3UDRCHcjQmkNSX5j6w8L9gbm5J1VGbz59yjjsX2AL85A7FYUmHMrNXe",
					"cadAmount": {
						"getCCoin": "0"
					},
					"cadIsUsed": false,
					"cadIsChange": false
				}],
				"caAmount": {
					"getCCoin": "0"
				}
			}
		}

		//failed 返回错误
		{
			"Left": {
				"tag": "RequestError",
				"contents": "Passphrase doesn't match"
			}
		}
	*/
	//r.ToJSON(&result)

	return r.Bytes()
}

// callGetAccountsAPI 获取账户信息
func callGetAccountsAPI(accountId ...string) []byte {

	var (
		//result map[string]interface{}
		param = make(req.QueryParam, 0)
	)

	method := "accounts"
	url := serverAPI + method
	if len(accountId) > 0 && len(accountId[0]) > 0 {
		//查询wid的钱包信息
		param["accountId"] = accountId[0]
	}
	r, err := api.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}
	//r.ToJSON(&result)
	//log.Printf("%+v\n", r)

	return r.Bytes()

}

// callGetAccountByIDAPI 获取置顶aid账户的信息
func callGetAccountByIDAPI(accountId string) []byte {

	var (
	//result map[string]interface{}
	//param = make(req.QueryParam, 0)
	)

	method := "accounts"
	url := fmt.Sprintf("%s%s/%s", serverAPI, method, accountId)
	r, err := api.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}
	//r.ToJSON(&result)
	//log.Printf("%+v\n", r)

	return r.Bytes()

}

// callCreateNewAddressAPI 调用创建账户地址
// @param aid
// @param passphrase
func callCreateNewAddressAPI(aid, passphrase string) []byte {
	var (
		//result map[string]interface{}
		body  interface{}
		param = make(req.QueryParam, 0)
	)

	method := "addresses"
	url := serverAPI + method

	/*
		传入参数
		CAccountId
	*/

	body = aid

	//设置密码
	param["passphrase"] = passphrase

	r, err := api.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil
	}

	//log.Printf("%+v\n", r)

	/*
		success 返回结果
		{
		  "cadId": "string",
		  "cadAmount": {
			"getCCoin": "string"
		  },
		  "cadIsUsed": true,
		  "cadIsChange": true
		}

	*/
	//r.ToJSON(&result)

	return r.Bytes()
}

// callSendTxAPI 发送交易
// @param from
// @param to
// @param amount
// @param passphrase
func callSendTxAPI(from, to string, amount uint64, passphrase string) []byte {
	var (
		//result map[string]interface{}
		body  map[string]interface{}
		param = make(req.QueryParam, 0)
	)

	//https://127.0.0.1:8090/api/txs/payments/{from}/{to}/{amount}
	method := "txs/payments"
	url := fmt.Sprintf("%s%s/%s/%s/%d", serverAPI, method, from, to, amount)

	/*
		传入参数
		CAccountId
	*/

	body = map[string]interface{}{
		"groupingPolicy": "OptimizeForHighThroughput",
	}

	//设置密码
	param["passphrase"] = passphrase

	r, err := api.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil
	}

	log.Printf("%v\n", r)

	return r.Bytes()
}
