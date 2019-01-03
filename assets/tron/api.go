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
	"strconv"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/imroc/req"
	"github.com/shopspring/decimal"
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
}

// NewClient create new client to connect
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

	r, err := req.Post(url, req.BodyJSON(&param), authHeader)
	if err != nil {
		log.Errorf("Failed: %+v >\n", err)
		return nil, err
	}
	// log.Std.Info("%+v", r)

	if r.Response().StatusCode != http.StatusOK {
		message := gjson.ParseBytes(r.Bytes()).String()
		message = fmt.Sprintf("[%s]%s", r.Response().Status, message)
		log.Error(message)
		return nil, errors.New(message)
	}

	res := gjson.ParseBytes(r.Bytes())
	return &res, nil
}

//getBalanceByExplorer 获取地址余额
func (wm *WalletManager) getBalanceByExplorer(address string) (*openwallet.Balance, error) {
	account, err := wm.GetAccount(address)

	//params := req.Param{"address": address}
	//result, err := wm.WalletClient.Call("/wallet/getaccount", params)
	// result, err := wm.WalletClient.Call(path, "Account")
	if err != nil {
		return nil, err
	}
	balance, err := strconv.ParseInt(account.Balance, 10, 64)
	if err != nil {
		return nil, err
	}
	floatBalance := (float64(balance) / 1000000)
	stringBalance := strconv.FormatFloat(floatBalance, 'f', -1, 64)
	obj := openwallet.Balance{}
	obj.Address = address
	obj.ConfirmBalance = stringBalance
	obj.UnconfirmBalance = "0.0"
	u, _ := decimal.NewFromString(obj.ConfirmBalance)
	b, _ := decimal.NewFromString(obj.UnconfirmBalance)
	obj.Balance = u.Add(b).StringFixed(6)
	return &obj, nil
}

func newBalanceByExplorer(json *gjson.Result) *openwallet.Balance {

	obj := openwallet.Balance{}
	//解析json
	obj.Address = gjson.Get(json.Raw, "addrStr").String()
	obj.ConfirmBalance = gjson.Get(json.Raw, "balance").String()
	obj.UnconfirmBalance = gjson.Get(json.Raw, "unconfirmedBalance").String()
	u, _ := decimal.NewFromString(obj.ConfirmBalance)
	b, _ := decimal.NewFromString(obj.UnconfirmBalance)
	obj.Balance = u.Add(b).StringFixed(6)
	return &obj
}
