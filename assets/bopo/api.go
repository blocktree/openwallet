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

package bopo

import (
	"fmt"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
)

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL string
	Auth    string
	Debug   bool
}

// // A struct of Response for Bitcoincash RPC response
// type Response struct {
// 	Code    int         `json:"code,omitempty"`
// 	Error   interface{} `json:"error,omitempty"`
// 	Result  interface{} `json:"result,omitempty"`
// 	Message string      `json:"message,omitempty"`
// 	Id      string      `json:"id,omitempty"`
// }

func (c *Client) Call(path, method string, request interface{}) ([]byte, error) {

	url := c.BaseURL + "/" + path
	authHeader := req.Header{"Accept": "application/json"}

	if c.Debug {
		log.Println("Start Request API...")
	}

	r, err := req.Do(method, url, request, authHeader)
	if err != nil {
		log.Printf("%+v\n", r)
		return nil, err
	}

	if c.Debug {
		log.Println("Request API Completed")
	}

	if r.Response().StatusCode != http.StatusOK {
		message := gjson.GetBytes(r.Bytes(), "message").String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		return nil, errors.New(message)
	}

	// Bopo 方面 API 在变，暂不验证^
	res := gjson.ParseBytes(r.Bytes()).Map()
	if code, ok := res["code"]; !ok || code.Int() != 0 {
		if msg, ok := res["msg"]; ok {
			log.Println(errors.New(msg.String()))
		} else {
			log.Println(errors.New("Invalid data format of Bopo Network!"))
		}
		// return nil, errors.New(res["msg"].String())	// 500
	}
	return r.Bytes(), nil
}
