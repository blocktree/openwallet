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
	"fmt"
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
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
	m.Node.HandleFunc("createWallet", m.createWallet)
	m.Node.HandleFunc("createAddress", m.createAddress)
	m.Node.HandleFunc("getAddressList", m.getAddressList)
	m.Node.HandleFunc("configWallet", m.configWallet)
	m.Node.HandleFunc("getWalletInfo", m.getWalletInfo)
	m.Node.HandleFunc("submitTransaction", m.submitTransaction)
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
		if s.WalletID == "" {
			continue
		}
		subscriptions = append(subscriptions, s)
		//log.Printf("s = %v\n", s)
		//添加订阅钱包
		wallet := openwallet.NewWatchOnlyWallet(s.WalletID, s.Coin)
		err = db.Save(wallet)
		if err != nil {
			responseError(ctx, err)
			return
		}
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
		errorMsg := fmt.Sprintf("%s assets manager not found!", coin)
		responseError(ctx, errors.New(errorMsg))
		return
	}

	wallet := openwallet.Wallet{
		WalletID: openwallet.NewWalletID().String(),
		Alias:    alias,
		Password: password,
	}

	err := am.CreateMerchantWallet(&wallet)
	if err != nil {
		responseError(ctx, err)
		return
	}

	db, err := m.OpenDB()
	if err != nil {
		responseError(ctx, err)
		return
	}
	defer db.Close()

	db.Save(&wallet)

	log.Printf("walletID = %s \n", wallet.WalletID)

	responseSuccess(ctx, wallet)
}

func (m *MerchantNode) configWallet(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func (m *MerchantNode) getWalletInfo(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func (m *MerchantNode) createAddress(ctx *owtp.Context) {

	log.Printf("Merchat Call: createAddress \n")
	log.Printf("params: %v\n", ctx.Params())

	/*
		| 参数名称 | 类型   | 是否可空 | 描述         |
		|----------|--------|----------|--------------|
		| coin     | string | 否       | 币种标识     |
		| walletID | string | 否       | 钱包ID       |
		| count    | uint   | 否       | 条数         |
		| password | string | 是       | 钱包解锁密码 |
	*/

	coin := ctx.Params().Get("coin").String()
	walletID := ctx.Params().Get("walletID").String()
	count := ctx.Params().Get("count").Uint()
	password := ctx.Params().Get("password").String()

	//导入到每个币种的数据库
	am := assets.GetMerchantAssets(coin)
	if am == nil {
		errorMsg := fmt.Sprintf("%s assets manager not found!", coin)
		responseError(ctx, errors.New(errorMsg))
		return
	}

	//提交给资产管理包转账
	wallet, err := m.GetMerchantWalletByID(walletID)
	if err != nil {
		responseError(ctx, err)
		return
	}
	wallet.Password = password

	//导入到每个币种的数据库
	mer := assets.GetMerchantAssets(coin)
	newAddrs, err := mer.CreateMerchantAddress(wallet, count)

	if err != nil {
		responseError(ctx, err)
		return
	}

	result := map[string]interface{}{
		"addresses": newAddrs,
	}

	responseSuccess(ctx, result)
}

func (m *MerchantNode) getAddressList(ctx *owtp.Context) {

	log.Printf("Merchat Call: getAddressList \n")
	log.Printf("params: %v\n", ctx.Params())

	/*
	| 参数名称 | 类型   | 是否可空 | 描述     |
	|----------|--------|----------|----------|
	| coin     | string | 否       | 币种标识 |
	| walletID | string | 否       | 钱包ID   |
	| offset    | uint   | 是       | 从0开始     |
	| limit    | uint   | 是       | 查询条数     |
	*/

	coin := ctx.Params().Get("coin").String()
	walletID := ctx.Params().Get("walletID").String()
	offset := ctx.Params().Get("offset").Uint()
	limit := ctx.Params().Get("limit").Uint()

	//导入到每个币种的数据库
	am := assets.GetMerchantAssets(coin)
	if am == nil {
		errorMsg := fmt.Sprintf("%s assets manager not found!", coin)
		responseError(ctx, errors.New(errorMsg))
		return
	}

	//提交给资产管理包转账
	wallet, err := m.GetMerchantWalletByID(walletID)
	if err != nil {
		responseError(ctx, err)
		return
	}

	//导入到每个币种的数据库
	mer := assets.GetMerchantAssets(coin)
	addrs, err := mer.GetMerchantAddressList(wallet, offset, limit)

	if err != nil {
		responseError(ctx, err)
		return
	}

	result := map[string]interface{}{
		"addresses": addrs,
	}

	responseSuccess(ctx, result)
}

func (m *MerchantNode) submitTransaction(ctx *owtp.Context) {

	log.Printf("Merchat Call: submitTransaction \n")
	log.Printf("params: %v\n", ctx.Params())

	var (
		//withdraws = make([]*openwallet.Withdraw, 0)
		wallets  = make(map[string][]*openwallet.Withdraw)
		tmpArray []*openwallet.Withdraw
	)

	db, err := m.OpenDB()
	if err != nil {
		responseError(ctx, err)
		return
	}
	defer db.Close()

	for _, p := range ctx.Params().Get("withdraws").Array() {
		s := openwallet.NewWithdraw(p)
		//withdraws = append(withdraws, s)
		db.Save(s)

		tmpArray = wallets[s.WalletID]
		if tmpArray == nil {
			tmpArray = make([]*openwallet.Withdraw, 0)
		}

		tmpArray = append(tmpArray, s)
		wallets[s.WalletID] = tmpArray

	}

	for wid, withs := range wallets {
		if len(withs) > 0 {
			//提交给资产管理包转账
			wallet, err := m.GetMerchantWalletByID(wid)
			if err != nil {
				responseError(ctx, err)
				return
			}
			wallet.Password = withs[0].Password
			//导入到每个币种的数据库
			mer := assets.GetMerchantAssets(withs[0].Symbol)
			mer.SubmitTransactions(wallet, withs)
		}
	}

	responseSuccess(ctx, nil)
}

//responseSuccess 成功结果响应
func responseSuccess(ctx *owtp.Context, result interface{}) {
	ctx.Response(result, owtp.StatusSuccess, "success")
}

//responseError 失败结果响应
func responseError(ctx *owtp.Context, err error) {
	ctx.Response(nil, owtp.ErrCustomError, err.Error())
}
