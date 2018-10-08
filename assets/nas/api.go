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

package nas

import (
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"strconv"

	//	"log"
	"errors"
	"github.com/nebulasio/go-nebulas/rpc/pb"
)

type Client struct {
	BaseURL     string
	Debug       bool
	Client      *req.Req
	Header      req.Header
}

//定义全局变量Nonce用于记录真正交易上链的nonce值和记录在DB中的nonce值
var Nonce int

func NewClient(url string, debug bool) *Client {
	c := Client{
		BaseURL:     url,
		Debug:       debug,
	}

	api := req.New()

	c.Client = api
	c.Header = req.Header{"Content-Type": "application/json"}

	return &c
}

func (c *Client) CallTestJson() (){

	trx := make(map[string]interface{},0)

	var Nonce string = "100"
	nonce,_:= strconv.ParseUint(Nonce,10,64)

	trx["from"] = "qwerty"
	trx["to"] = "asdf"
	trx["value"] = "123"
	trx["nonce"] = nonce
	trx["gasLimit"] = "1212"
	trx["gasPrice"] = "10000"

	fmt.Printf("trx=%v\n\n",trx)

	tx := &rpcpb.TransactionRequest{
		"qwerty",
		"asdf",
		"123",
		123 ,
		"123" ,
		"1222",
		nil,
		nil,
		"",
	}
	fmt.Printf("tx=%v\n",tx)
}



//确定nonce值
func (c *Client) CheckNonce(key *Key) uint64{

	nonce_get,_ := c.CallGetaccountstate(key.Address,"nonce")
	nonce_chain ,_ := strconv.Atoi(nonce_get) 	//当前链上nonce值
	nonce_db,_ := strconv.Atoi(key.Nonce)	//本地记录的nonce值

	//如果本地nonce_local > 链上nonce,采用本地nonce,否则采用链上nonce
	if nonce_db > nonce_chain{
		Nonce = nonce_db + 1
		log.Std.Info("%s nonce_db=%d > nonce_chain=%d,Use nonce_db+1...",key.Address,nonce_db,nonce_chain)
	}else{
		Nonce = nonce_chain + 1
		log.Std.Info("%s nonce_db=%d <= nonce_chain=%d,Use nonce_chain+1...",key.Address,nonce_db,nonce_chain)
	}

	return uint64(Nonce)
}

//查询每个地址balance、nonce
//address:n1S8ojaa9Pz8TduXEm8vXrxBs6Kz5dyp7km
//query:balance、nonce
func (c *Client) CallGetaccountstate( address string ,query string) (string, error) {

	url := c.BaseURL + "/v1/user/accountstate"
//	fmt.Printf("url=%s\n",url)

	var (
		body = make(map[string]interface{}, 0)
	)

	if c.Client == nil {
		return "", errors.New("API url is not setup. ")
	}

	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": "Basic " ,
	}

	//json-rpc
//	body["jsonrpc"] = "2.0"
//	body["id"] = "1"
//	body["method"] = 10
//	body["params"] = 12121
	body["address"] = address

	if c.Debug {
		log.Info("Start Request API...")
	}

	r, err := c.Client.Post(url, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return "", err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return "", err
	}
	//resp :  {"result":{"address":"n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"}}
	//result:  "result" : {"address":"n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"}
	//result:  "result.address" : "n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"
	dst := "result." + query
	result := resp.Get(dst)

	return result.Str, nil
}


//查询区块链chain_id，testnet:	mainnet:
func (c *Client) CallGetchain_id() uint32 {
	url := c.BaseURL + "/v1/user/nebstate"
	param := make(req.QueryParam, 0)

	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Info(err)
		return 0
	}

	//	return r.Bytes()
	if c.Debug {
		log.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return 0
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return 0
	}

	result := resp.Get("result.chain_id")
	return uint32(result.Num)
}


//查询GasPrice
func (c *Client) CallGetGasPrice() string {
	url := c.BaseURL + "/v1/user/getGasPrice"
	param := make(req.QueryParam, 0)

	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Info(err)
		return ""
	}

	if c.Debug {
		log.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return ""
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return ""
	}

	result := resp.Get("result.gas_price")
	return (result.Str)
}


//发送广播签名后的交易单数据
func (c *Client) CallSendRawTransaction( data string ) (string, error) {

	url := c.BaseURL + "/v1/user/rawtransaction"

	var (
		body = make(map[string]interface{}, 0)
	)

	if c.Client == nil {
		return "", errors.New("API url is not setup. ")
	}

	authHeader := req.Header{
		"Accept":        "application/json",
		"Authorization": "Basic " ,
	}

	//json-rpc
	//	body["jsonrpc"] = "2.0"
	//	body["id"] = "1"
	//	body["method"] = path
	//	body["params"] = request
	body["data"] = data

	if c.Debug {
		log.Info("Start Request API...")
	}

	r, err := c.Client.Post(url, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return "", err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return "", err
	}
	//resp :  {"result":{"address":"n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"}}
	//result:  "result" : {"address":"n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"}
	//result:  "result.address" : "n1Qmnmuebg4xxvnuHUoSLDjLFMznxMdsDng"

	result := resp.Get("result.txhash")
	return result.Str, nil
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	if !result.Get("error").IsObject() {

		if !result.Get("result").Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String())
	err = errors.New(errInfo)

	return err
}


