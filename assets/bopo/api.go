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

package bopo

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/imroc/req"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

// A Client is a Bitcoin RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL string
	//AccessToken string
	Debug  bool
	client *req.Req
}

// A struct of Response for BOPO RESTful response
type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

func NewClient(url string, debug bool) *Client {

	c := &Client{
		BaseURL: url,
		// AccessToken: token,
		Debug: debug,
	}

	reqC := req.New()
	// trans, _ := reqC.Client().Transport.(*http.Transport)
	// trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.client = reqC

	return c
}

func (c *Client) Call(path, method string, request interface{}) ([]byte, error) {

	if c == nil || c.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	url := c.BaseURL + "/" + path
	authHeader := req.Header{"Accept": "application/json"}

	r, err := c.client.Do(method, url, request, authHeader)
	if err != nil {
		log.Printf("Failed: %+v >\n", err)
		return nil, err
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
			return nil, errors.New(fmt.Sprintf("Bopo returns invalid! \nReturn: %s", gjson.ParseBytes(rs).String()))
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
