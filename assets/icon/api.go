/*
 * Copyright 2018 The openwallet Authors
 * This file is part of the openwallet library.
 *
 * The openwallet library is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The openwallet library is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Lesser General Public License for more details.
 */

package icon

import (
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"github.com/blocktree/openwallet/log"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
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
	c.Client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (c *Client) Call(path string, request map[string]interface{}) (*gjson.Result, error) {

	var (
		body = make(map[string]interface{}, 0)
	)

	if c.Client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	authHeader := req.Header{
		"Accept":        "application/json",
	}

	//json-rpc
	body["jsonrpc"] = "2.0"
	body["id"] = "1"
	body["method"] = path
	body["params"] = request

	if c.Debug {
		log.Std.Info("Start Request API...")
	}

	r, err := c.Client.Post(c.BaseURL, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Std.Info("Request API Completed")
	}

	if c.Debug {
		log.Std.Info("%+v", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func BasicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	/*
		// Response - error
		{
			"jsonrpc": "2.0",
			"id": 1234,
			"error": {
				"code": -32602,
				"message": "Invalid address"
			}
		}
	*/

	if result.Get("result").Exists() {
		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String())
	err = errors.New(errInfo)

	return err
}

func (c *Client) Call_icx_getBalance(address string) (string, error) {
	request := map[string]interface{} {
		"address":address,
	}

	ret, err := c.Call("icx_getBalance", request)
	if err != nil {
		return "", err
	}

	bigint, _ := hexutil.DecodeBig(ret.String())
	b := decimal.NewFromBigInt(bigint, 0).Div(coinDecimal)

	return b.String(), nil
}

func (c *Client) Call_icx_sendTransaction(request map[string]interface{}) (string, error) {
	ret, err := c.Call("icx_sendTransaction", request)
	if err != nil {
		return "", err
	}

	return ret.String(), nil
}

func (c *Client) Call_icx_getTransactionByHash(txhash string) (string, error) {
	request := map[string]interface{} {
		"txHash":txhash,
	}

	ret, err := c.Call("icx_getTransactionByHash", request)
	if err != nil {
		return "", err
	}

	return ret.String(), nil
}