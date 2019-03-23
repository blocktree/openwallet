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

package eosio

import (
	"encoding/base64"
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
	"strings"
)

type ClientInterface interface {
	Call(path string, request []interface{}) (*gjson.Result, error)
}

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
}

func NewClient(url string, debug bool) *Client {
	c := Client{
		BaseURL: strings.TrimSuffix(url, "/"),
		Debug:   debug,
	}

	api := req.New()
	c.client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (c *Client) Call(path string, request interface{}) (*gjson.Result, error) {

	if request == nil {
		request = struct {}{}
	}

	if c.client == nil {
		return nil, fmt.Errorf("API url is not setup. ")
	}

	if c.Debug {
		log.Std.Info("Start Request API...")
	}
	urlString := c.BaseURL + "/" + strings.TrimPrefix(path, "/")
	r, err := c.client.Post(urlString, req.BodyJSON(&request))

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

	return &resp, nil
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
		//failed 返回错误
		{
		    "code": 500,
		    "message": "Internal Service Error",
		    "error": {
		        "code": 7,
		        "name": "bad_cast_exception",
		        "what": "Bad Cast",
		        "details": [
		            {
		                "message": "Invalid cast from type 'string_type' to Object",
		                "file": "variant.cpp",
		                "line_number": 583,
		                "method": "get_object"
		            }
		        ]
		    }
		}
	*/

	if result.Get("error").IsObject() {

		errInfo := fmt.Sprintf("[%d]%s",
			result.Get("error.code").Int(),
			result.Get("error.what").String())
		err = fmt.Errorf(errInfo)

		return err
	}

	return nil
}
