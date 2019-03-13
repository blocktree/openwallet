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

package sia

import (
	"encoding/base64"
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

	//Client *req.Req
}

type Response struct {
	Code    int         `json:"code,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Message string      `json:"message,omitempty"`
	Id      string      `json:"id,omitempty"`
}

// Call calls for batch address
func (c *Client) CallBatchAddress(path, method string, request interface{}) ([]byte, error) {

	url := c.BaseURL + "/" + path

	authHeader := req.Header{
		"Accept":        "application/json",
		"User-Agent":    "Sia-Agent",
		"Authorization": "Basic " + basicAuth("", c.Auth),
	}

	r, err := req.Do(
		method,
		url,
		request,
		authHeader)

	if err != nil {
		return nil, err
	}

	return r.Bytes(), nil
}

func (c *Client) Call(path, method string, request interface{}) ([]byte, error) {

	url := c.BaseURL + "/" + path

	authHeader := req.Header{
		"Accept":        "application/json",
		"User-Agent":    "Sia-Agent",
		"Authorization": "Basic " + basicAuth("", c.Auth),
	}

	if c.Debug {
		log.Println("Start Request API...")
	}

	r, err := req.Do(
		method,
		url,
		request,
		authHeader)

	if c.Debug {
		log.Println("Request API Completed")
	}

	if c.Debug {
		log.Printf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	if r.Response().StatusCode != http.StatusOK {
		message := gjson.GetBytes(r.Bytes(), "message").String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		return nil, errors.New(message)
	}

	return r.Bytes(), nil
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
