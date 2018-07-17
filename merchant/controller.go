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
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
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
	m.Node.HandleFunc("subscribe", m.createWallet)
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
	//
	////每次订阅都先清除旧订阅
	//db.Drop("subscribe")

	for _, p := range ctx.Params().Get("subscriptions").Array() {
		s := NewSubscription(p)
		subscriptions = append(subscriptions, s)

		//检查订阅的钱包是否存在，不存在随机创建
		wallet := openwallet.NewWatchOnlyWallet(s.WalletID)
		db.Save(wallet)
	}

	//重置订阅内容
	m.resetSubscriptions(subscriptions)



	//启动订阅交易记录任务

	responseSuccess(ctx, nil)
}

func (m *MerchantNode) createWallet(ctx *owtp.Context) {

	log.Printf("Merchat Call: createWallet \n")
	log.Printf("params: %v\n", ctx.Params())

	/*
	| 参数名称     | 类型   | 是否可空 | 描述                    |
	|--------------|--------|----------|-------------------------|
	| coin         | string | 否       | 币种标识                |
	| alias        | string | 否       | 钱包别名                |
	| passwordType | uint   | 否       | 0：自定义密码，1：协商密码 |
	| password     | string | 是       | 自定义密码              |
	| authKey    | string | 是       | 授权公钥                |
	*/

	coin := ctx.Params().Get("coin").String()
	alias := ctx.Params().Get("alias").String()
	//passwordType := ctx.Params().Get("passwordType").Uint()
	password := ctx.Params().Get("password").String()

	//导入到每个币种的数据库
	am := assets.GetMerchantAssets(coin)
	if am == nil {
		responseError(ctx, errors.New("Assets manager no find!"))
		return
	}
	wallet, err := am.CreateMerchantWallet(alias, password)
	if err != nil {
		responseError(ctx, err)
		return
	}

	//生成随机的钱包ID
	wallet.WalletID = openwallet.NewWalletID().String()

	db, err := m.OpenDB()
	if err != nil {
		responseError(ctx, err)
		return
	}
	defer db.Close()

	db.Save(wallet)

	log.Printf("walletID = %s \n", wallet.WalletID)

	responseSuccess(ctx, wallet)
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
	ctx.Response(result, owtp.StatusSuccess, "success")
}

//responseError 失败结果响应
func responseError(ctx *owtp.Context, err error) {
	ctx.Response(nil, owtp.ErrBadRequest, err.Error())
}
