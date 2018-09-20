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

package tron

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/blocktree/OpenWallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

// A Client is a Tron RPC client. It performs RPCs over HTTP using JSON
// request and responses. A Client must be configured with a secret token
// to authenticate with other Cores on the network.
type Client struct {
	BaseURL string
	// AccessToken string
	Debug  bool
	client *req.Req
	//Client *req.Req
}

// type Response struct {
// 	Code    int         `json:"code,omitempty"`
// 	Error   interface{} `json:"error,omitempty"`
// 	Result  interface{} `json:"result,omitempty"`
// 	Message string      `json:"message,omitempty"`
// 	Id      string      `json:"id,omitempty"`
// }

func NewClient(url, token string, debug bool) *Client {
	c := Client{
		BaseURL: url,
		// AccessToken: token,
		Debug: debug,
	}

	api := req.New()
	//trans, _ := api.Client().Transport.(*http.Transport)
	//trans.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.client = api

	return &c
}

// Call calls a remote procedure on another node, specified by the path.
func (c *Client) Call(path string, param interface{}) (*gjson.Result, error) {

	if c == nil || c.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	url := c.BaseURL + path
	authHeader := req.Header{"Accept": "application/json"}
	// if c.Debug {
	// 	log.Std.Info("Start Request API...")
	// }

	// r, err := c.client.Do("POST", url, request, authHeader)
	r, err := req.Post(url, req.BodyJSON(&param), authHeader)
	// r, err := c.client.Post(c.BaseURL+path, req.BodyJSON(&body))

	// if c.Debug {
	// 	log.Std.Info("Request API Completed")
	// }

	// if c.Debug {
	// 	log.Std.Info("%+v", r)
	// }

	if err != nil {
		log.Error("Failed: %+v >\n", err)
		return nil, err
	}

	if r.Response().StatusCode != http.StatusOK {
		message := gjson.ParseBytes(r.Bytes()).String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		log.Error(message)
		return nil, errors.New(message)
	}

	resp := gjson.ParseBytes(r.Bytes())

	return &resp, nil
}

// Call calls a remote procedure on another node, specified by the path.
// func (c *Client) Call(path string, request []interface{}) (*gjson.Result, error) {
func (c *Client) Call2(path string, param interface{}) ([]byte, error) {

	if c == nil || c.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	url := c.BaseURL + path
	authHeader := req.Header{"Accept": "application/json"}

	r, err := req.Post(url, req.BodyJSON(&param), authHeader)
	// r, err := c.client.Do("POST", url, request, authHeader)
	if err != nil {
		log.Error("Failed: %+v >\n", err)
		return nil, err
	}

	if c.Debug {
		// log.Std.Info("%+v", r)
	}

	if r.Response().StatusCode != http.StatusOK {
		message := gjson.ParseBytes(r.Bytes()).String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		log.Error(message)
		return nil, errors.New(message)
	}

	return r.Bytes(), nil
}
