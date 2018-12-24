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
	"fmt"
	"github.com/imroc/req"
	"log"
	"strings"
	"crypto/x509"
	"io/ioutil"
)

type Client struct {
	BaseURL     string
	Debug       bool
	Client      *req.Req
	Header      req.Header
}

func NewClient(url string, debug bool) *Client {
	c := Client{
		BaseURL:     url,
		Debug:       debug,
	}

	api := req.New()
/*
	//tls
	pair, e := tls.LoadX509KeyPair(certsDir + "/client.crt", certsDir + "/client.key")
	if e != nil {
		log.Fatal("LoadX509KeyPair:", e)
	}
	trans, _ := api.Client().Transport.(*http.Transport)
	trans.TLSClientConfig = &tls.Config{
		RootCAs:      loadCA(certsDir + "/ca.crt"),
		Certificates: []tls.Certificate{pair},
	}
*/
	c.Client = api

	return &c
}

//加载ca证书
func loadCA(caFile string) *x509.CertPool {
	pool := x509.NewCertPool()

	if ca, e := ioutil.ReadFile(caFile); e != nil {
		log.Fatal("ReadFile: ", e)
	} else {
		pool.AppendCertsFromPEM(ca)
	}
	return pool
}

//callCreateWalletAPI 调用创建钱包接口
func (c *Client) callCreateWalletAPI(name, words, passphrase string, create bool) []byte {

	var (
		body           = make(map[string]interface{}, 0)
		method         string
	)

	if create {
		method = "create"
	} else {
		method = "restore"
	}

	url := c.BaseURL + "wallets"

	body["operation"] = method
	//分割助记词
	mnemonicArray := strings.Split(words, " ")
	body["backupPhrase"] = mnemonicArray

	body["assuranceLevel"] = "normal"
	body["name"] = name

	if len(passphrase) > 0 {
		//设置密码
		body["spendingPassword"] = passphrase
	}

	r, err := c.Client.Post(url, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%+v\n", r)
	}
	return r.Bytes()
}

//GetWalletInfo 查询钱包信息,输入一个表示查询指定钱包，空则查询全部
func (c *Client) callGetWalletAPI(wid ...string) []byte {
	method := "wallets"
	url := c.BaseURL + method
	param := make(req.Param, 0)
	if len(wid) > 0 && len(wid[0]) > 0 {
		//查询wid的钱包信息
		url = fmt.Sprint(url, "/", wid[0])
	} else {
		//查询全部钱包信息

	}
	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	/*
	success 返回结果
	{
	  "data": {
		"createdAt": "2032-07-26T13:46:01.035803",
		"syncState": {
		  "tag": "synced",
		  "data": null
		},
		"balance": 41984918983627330,
		"hasSpendingPassword": false,
		"assuranceLevel": "normal",
		"name": "My wallet",
		"id": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg",
		"spendingPasswordLastUpdate": "2029-04-05T12:13:13.241896"
	  },
	  "status": "success",
	  "meta": {
		"pagination": {
		  "totalPages": 0,
		  "page": 1,
		  "perPage": 10,
		  "totalEntries": 0
		}
	  }
	}
	*/

	return r.Bytes()
}

// callCreateNewAccountAPI 调用创建账户API，传入账户名，钱包id
func (c *Client) callCreateNewAccountAPI(name, wid, passphrase string) []byte {
	var (
		body  map[string]interface{}
		param = make(req.QueryParam, 0)
	)
	///api/v1/wallets/{walletId}/accounts
	method := "accounts"
	url := c.BaseURL + "wallets/" + wid + "/" + method

	/*
		传入参数
		{
		"name": "𶂯",
		"spendingPassword": ""
		}
	*/

	body = map[string]interface{}{
		"name": name,
		"spendingPassword": passphrase,
	}

	param["walletId"] = wid

	r, err := c.Client.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	/*
		success 返回结果
		{
		  "data": {
			"amount": 11091176260625604,
			"addresses": [
			  {
				"used": true,
				"changeAddress": true,
				"id": "Ae2tdPwUPEZ3hmyBxBGfSLZHzETDof8afvg1vFogbukKJS75xChNyvwiuR6"
			  }
			],
			"name": "My account",
			"walletId": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg",
			"index": 1
		  },
		  "status": "success",
		  "meta": {
			"pagination": {
			  "totalPages": 0,
			  "page": 1,
			  "perPage": 10,
			  "totalEntries": 0
			}
		  }
	}
	*/

	return r.Bytes()
}

// callGetAccountsAPI 获取账户信息
func (c *Client) callGetAccountsAPI(wid string, accountId ...string) []byte {

	var (
		param = make(req.QueryParam, 0)
	)
	///api/v1/wallets/{walletId}/accounts/{accountId}
	method := "accounts"
	url := c.BaseURL + "wallets/" + wid + "/" + method
	if len(accountId) > 0 && len(accountId[0]) > 0 {
		//查询wid的钱包信息
		url = fmt.Sprint(url, "/", accountId[0])
	}
	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	return r.Bytes()

}

// callGetAccountByIDAPI 获取置顶aid账户的信息
func (c *Client) callGetAccountByIDAPI(wid, accountId string) []byte {
	///api/v1/wallets/{walletId}/accounts/{accountId}
	method := "accounts"
	url := fmt.Sprintf("%swallets/%s/%s/%s", c.BaseURL, wid, method, accountId)
	r, err := c.Client.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%+v\n", r)
	}

	return r.Bytes()

}

// callCreateNewAddressAPI 调用创建账户地址
// @param wid        钱包id
// @param aid        账户index
//@param passphrase  钱包密码
func (c *Client) callCreateNewAddressAPI(wid string, aid int64, passphrase string) []byte {
	var (
		//result map[string]interface{}
	    body = make(map[string]interface{})
		param = make(req.QueryParam, 0)
	)

	method := "addresses"
	url := c.BaseURL + method

	/*
		传入参数
		{
			"accountIndex": 2,
			"walletId": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg",
			"spendingPassword": "0001010001010101010200020001010201000200010201020202000200010000"
		}
	*/


	//设置密码
	body["spendingPassword"] = passphrase
	body["accountIndex"] = aid
	body["walletId"] = wid

	r, err := c.Client.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%+v\n", r)
	}

	/*
		success 返回结果
	{
	  "data": {
		"used": true,
		"changeAddress": false,
		"id": "2reD6PLk2vQ44znyXUjiNZ6GBu4U4ibssQMv1f1B4sMhZecZ5yE2ZPtSZmFBqdA7QA88h337S9s4t6qBNeTYRsGsKPTP8xdxkFpXsJs4J9qPRX3zi1rCrvsuF2sF6s9TSKokGvigySsm73jRuusBaCyvYY4tiGdoY5X9XRihqR56PpgA2Ty5XCrb3gVhZyfxiJtn5ZpS1baxQMGrcJ8dFn76Ko8HXqo9C62EqdCiWcFEmkUPq15Kj5oUS9eZryU5izRCjp9EWmubNp3YgEgRJjFx6wr8aZxxFyphaUYueuRGzcjGDv8vaATXfTmLVtWrusjwc45QscHvPv6yk1",
		"ownership": "ambiguousOwnership"
	  },
	  "status": "success",
	  "meta": {
		"pagination": {
		  "totalPages": 0,
		  "page": 1,
		  "perPage": 10,
		  "totalEntries": 0
		}
	  }
	}
	*/

	return r.Bytes()
}

// callSendTxAPI 发送交易
// @param wid           钱包ID
// @param aid           账户index
// @param to            接收地址
// @param amount        发送数量
// @param passphrase    钱包密码
func (c *Client) callSendTxAPI(wid string, aid int64, to string, amount uint64, passphrase string) ([]byte, error) {
	var (
		body  map[string]interface{}
		param = make(req.QueryParam, 0)
	)

	///api/v1/transactions
	method := "transactions"
	url := c.BaseURL + method
	//参数
	/*
	{
		  "groupingPolicy": null,
		  "destinations": [
			{
			  "amount": 27889372662178400,
			  "address": "AL91N9VXRTCv3vnvq89Wj4DxEV6atyovsJoi7fBR7vU9bvecgDi4wxmS78FfdjxcBB99JeqkFjjxCAZL8czYPwSGX6frfPFt8nLnyp3V4Yh362d48Mz"
			},
			{
			  "amount": 33829516096667580,
			  "address": "2cWKMJemoBamMBMBYJWe5aQR3xitBLbjVbt4yxQUWUWC8KGGHvYF8BfUQwmAs5je5fSii"
			}
		  ],
		  "source": {
			"accountIndex": 2,
			"walletId": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg"
		  },
		  "spendingPassword": "0200020002020202010000010101020100020001020201020001000101010101"
		}
	 */

	body = map[string]interface{}{
		"groupingPolicy": "OptimizeForHighThroughput",
		"spendingPassword": passphrase,
	}

	source := map[string]interface{} {
		"accountIndex": aid,
		"walletId": wid,
	}

	dst1 := map[string]interface{} {
		"amount": amount,
		"address": to,
	}

	body["source"] = source

	var dst []interface{}
	dst = append(dst, dst1)
	body["destinations"] = dst

	r, err := c.Client.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	return r.Bytes(), nil
}

//callEstimateFeesAPI 计算矿工费
func (c *Client) callEstimateFeesAPI(wid string, aid int64, to string, amount uint64, passphrase string) ([]byte, error){
	var (
		body  map[string]interface{}
		param = make(req.QueryParam, 0)
	)

	///api/v1/transactions
	method := "transactions/fees"
	url := c.BaseURL + method
	//参数参考例子
	/*
	{
		  "groupingPolicy": null,
		  "destinations": [
			{
			  "amount": 27889372662178400,
			  "address": "AL91N9VXRTCv3vnvq89Wj4DxEV6atyovsJoi7fBR7vU9bvecgDi4wxmS78FfdjxcBB99JeqkFjjxCAZL8czYPwSGX6frfPFt8nLnyp3V4Yh362d48Mz"
			},
			{
			  "amount": 33829516096667580,
			  "address": "2cWKMJemoBamMBMBYJWe5aQR3xitBLbjVbt4yxQUWUWC8KGGHvYF8BfUQwmAs5je5fSii"
			}
		  ],
		  "source": {
			"accountIndex": 2,
			"walletId": "J7rQqaLLHBFPrgJXwpktaMB1B1kQBXAyc2uRSfRPzNVGiv6TdxBzkPNBUWysZZZdhFG9gRy3sQFfX5wfpLbi4XTFGFxTg"
		  },
		  "spendingPassword": "0200020002020202010000010101020100020001020201020001000101010101"
		}
	 */

	body = map[string]interface{}{
		"groupingPolicy": "OptimizeForHighThroughput",
		"spendingPassword": passphrase,
	}

	source := map[string]interface{} {
		"accountIndex": aid,
		"walletId": wid,
	}

	dst1 := map[string]interface{} {
		"amount": amount,
		"address": to,
	}

	body["source"] = source

	var dst []interface{}
	dst = append(dst, dst1)
	body["destinations"] = dst

	//设置密码
	param["passphrase"] = passphrase

	r, err := c.Client.Post(url, param, req.BodyJSON(&body))

	if err != nil {
		log.Println(err)
		return nil, err
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	return r.Bytes(), nil
}

//callDeleteWallet 调用删除钱包所有信息API
func (c *Client) callDeleteWallet(wid string) []byte {
	var (
	//result map[string]interface{}
	//body  map[string]interface{}
	//param = make(req.QueryParam, 0)
	)

	//https://127.0.0.1:8090/api/wallets/{walletId}
	method := "wallets"
	url := fmt.Sprintf("%s%s/%s", c.BaseURL, method, wid)

	r, err := c.Client.Delete(url)

	if err != nil {
		log.Println(err)
		return nil
	}

	if c.Debug {
		log.Printf("%v\n", r)
	}

	return r.Bytes()
}

//查询节点状态信息
func (c *Client) callGetNodeInfo() []byte {
	url := c.BaseURL + "node-info"
	r, err := c.Client.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}