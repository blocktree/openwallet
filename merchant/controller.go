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
	"log"
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

//subscribe 订阅方法
func (m *MerchantNode) subscribe(ctx *owtp.Context) {

	log.Printf("Merchat Call: subscribe \n")
	log.Printf("params: %v\n", ctx.Params())

	var (
		subscriptions []*Subscription
	)

	db, err := m.OpenDB()
	if err != nil {
		responseError(ctx, err)
		return
	}
	defer db.Close()

	//每次订阅都先清除旧订阅
	db.Drop("subscribe")

	//1. 把订阅记录写入到数据库
	for _, p := range ctx.Params().Get("subscriptions").Array() {
		s := NewSubscription(p)
		subscriptions = append(subscriptions, s)
		db.Save(s)
	}

	//2. 向商户获取订阅的地址列表

	//3. 启动定时器，检查订阅地址的最新状态（交易记录，余额）

	responseSuccess(ctx, nil)
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
