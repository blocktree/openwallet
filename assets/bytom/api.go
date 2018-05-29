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

package bytom

import (
	"encoding/base64"
	"github.com/imroc/req"
	"log"
)

var (
//client *Client
//url    = "http://192.168.2.224:10031"
//rpcuser = "wallet"
//rpcpassword = "walletPassword2017"
)

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL     string
	AccessToken string
	Debug       bool

	//Client *req.Req
}

type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

// Call calls a remote procedure on another node, specified by the path.
func (c *Client) Call(path string, request interface{}) []byte {

	//var (
	//	body = make(map[string]interface{}, 0)
	//)

	//authHeader := req.Header{
	//	"Accept":        "application/json",
	//	"Authorization": "Basic " + basicAuth(rpcuser, rpcpassword),
	//}

	//json-rpc
	//body["jsonrpc"] = "2.0"
	//body["id"] = "1"
	//body["method"] = path
	//body["params"] = request

	url := c.BaseURL + "/" + path

	if c.Debug {log.Println("Start Request API...")}

	r, err := req.Post(url, req.BodyJSON(&request))
	if err != nil {
		log.Printf("unexpected err: %v\n", err)
		return nil
	}

	if c.Debug {log.Println("Request API Completed")}

	if c.Debug {
		log.Printf("%+v\n", r)
	}

	return r.Bytes()
}

// See 2 (end of page 4) http://www.ietf.org/rfc/rfc2617.txt
// "To receive authorization, the client sends the userid and password,
// separated by a single colon (":") character, within a base64
// encoded string in the credentials."
// It is not meant to be urlencoded.
func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}
