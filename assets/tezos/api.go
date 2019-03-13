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

package tezos

import (
	"github.com/imroc/req"
	"log"
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
	c.Header = req.Header{"Content-Type": "application/json"}

	return &c
}

func (c *Client) CallGetHeader() []byte {
	url := c.BaseURL + "/chains/main/blocks/head/header"
	param := make(req.QueryParam, 0)

	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func (c *Client) CallGetCounter(pkh string) []byte {
	url := c.BaseURL + "/chains/main/blocks/head/context/contracts/" + pkh + "/counter"

	r, err := c.Client.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	//因为结果为"number"\n ，所以去掉双引号和\n
	lenght := len(r.Bytes())
	return r.Bytes()[1:lenght-2]
}

func (c *Client) CallGetManagerKey(pkh string) []byte {
	url := c.BaseURL + "/chains/main/blocks/head/context/contracts/" + pkh + "/manager_key"

	r, err := c.Client.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func (c *Client) CallForgeOps(chain_id string, head_hash string, body interface{}) string {
	url := c.BaseURL + "/chains/" + chain_id + "/blocks/" + head_hash + "/helpers/forge/operations"
	param := make(req.Param, 0)

	//log.Println(body)
	r, err := c.Client.Post(url, param, c.Header, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return ""
	}
	//因为结果为"hex"\n ，所以去掉双引号和\n
	lenght := len(r.Bytes())
	return string(r.Bytes()[1:lenght-2])
}

func (c *Client) CallPreapplyOps(body interface{}) []byte{
	url := c.BaseURL + "/chains/main/blocks/head/helpers/preapply/operations"
	param := make(req.Param, 0)

	r, err := c.Client.Post(url, param, c.Header, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func (c *Client) CallInjectOps(body string) []byte {
	url := c.BaseURL + "/injection/operation"
	param := make(req.Param, 0)

	r, err := c.Client.Post(url, param, c.Header, req.BodyJSON(&body))

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return r.Bytes()
}

func (c *Client) CallGetbalance(addr string) []byte {
	url := c.BaseURL + "/chains/main/blocks/head/context/contracts/" + addr + "/balance"
	param := make(req.QueryParam, 0)

	r, err := c.Client.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	//因为结果为"number"\n ，所以去掉双引号和\n
	lenght := len(r.Bytes())
	return r.Bytes()[1:lenght-2]
}

