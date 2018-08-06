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
	"github.com/codeskyblue/go-sh"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"log"
	"net/http"
	"strings"
)

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL string
	Auth    string
	Debug   bool
}

// A struct of Response for BOPO RESTful response
type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

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
		message := gjson.ParseBytes(r.Bytes()).String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		return nil, errors.New(message)
	}

	// 500: Bopo API 变动较大，测试链和公链返回值结构不稳定，注意验证过程^
	rs := r.Bytes()
	code := gjson.GetBytes(rs, "code")
	if code.Exists() != true || code.Int() != 0 {
		msg := gjson.GetBytes(rs, "msg")
		if msg.Exists() != true {
			return nil, errors.New(fmt.Sprintf("Bopo returns invalid! \nReturn: ", gjson.ParseBytes(rs).String()))
		}
		// log.Println(fmt.Sprintf("Bopo returns invalid! \nReturn: %s", gjson.ParseBytes(rs).String()))
		return nil, errors.New(msg.String())
	}
	data := gjson.GetBytes(rs, "data")
	if data.Exists() != true {
		return nil, errors.New(fmt.Sprintf("Bopo fullnode returns invalid 'data'! \nReturn: %s", gjson.ParseBytes(rs).String()))
	}

	return []byte(data.String()), nil
}

//cmdCall 执行命令
func cmdCall(cmd string, wait bool) error {
	var (
		cmdName string
		args    []string
	)

	cmds := strings.Split(cmd, " ")
	if len(cmds) > 0 {
		cmdName = cmds[0]
		args = cmds[1:]
	} else {
		return errors.New("command not found ")
	}
	session := sh.Command(cmdName, args)
	if wait {
		return session.Run()
	} else {
		return session.Start()
	}
}
