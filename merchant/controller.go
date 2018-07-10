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

package merchant

import (
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/pkg/errors"
)

const (

	//订阅类型，1：钱包余额，2：充值记录，3：汇总日志
	SubscribeTypeBalance    = 1
	SubscribeTypeCharge     = 2
	SubscribeTypeSummaryLog = 3
)

var (
	//商户节点
	merchantNode *MerchantNode

	/* 异常错误 */

	//节点断开
	ErrMerchantNodeDisconnected = errors.New("Merchant node is not connected!")
)

/********** 钱包管理相关方法【被动】 **********/

//setupRouter 配置路由
func (m *MerchantNode) setupRouter() {

	m.Node.HandleFunc("subscribe", m.subscribe)
	m.Node.HandleFunc("configWallet", m.configWallet)
	m.Node.HandleFunc("getWalletInfo", m.getWalletInfo)
	m.Node.HandleFunc("submitTrasaction", m.submitTrasaction)
}

func (m *MerchantNode) configWallet(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func (m *MerchantNode) getWalletInfo(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func (m *MerchantNode) submitTrasaction(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

//responseSuccess 成功结果响应
func responseSuccess(ctx *owtp.Context, result interface{}) {
	ctx.Response(result, 1000, "success")
}

//responseError 失败结果响应
func responseError(ctx *owtp.Context, err error) {
	ctx.Response(nil, 1001, err.Error())
}
