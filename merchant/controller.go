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
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/common/file"
	"github.com/blocktree/OpenWallet/owtp"
	"log"
	"path/filepath"
	"time"
)

var (
	//商户节点
	merchantNode *owtp.OWTPNode
	//连接状态通道
	reconnect chan bool
	//断开状态通道
	disconnected chan struct{}
	//是否重连
	isReconnect bool
	//重连时的等待时间
	reconnectWait time.Duration = 10
)

func init() {
	isReconnect = true
}

//访问数据库
func openDB() (*storm.DB, error) {
	file.MkdirAll(merchantDir)
	return storm.Open(filepath.Join(merchantDir, cacheFile))
}

//run 运行商户节点管理
func run() error {

	var (
		err error
	)

	defer func() {
		close(reconnect)
		close(disconnected)
	}()

	reconnect = make(chan bool, 1)
	disconnected = make(chan struct{}, 1)

	//启动连接
	reconnect <- true

	log.Printf("Merchant node running now... \n")

	//节点运行时
	for {
		select {
		case <-reconnect:
			//重新连接
			log.Printf("Connecting to %s\n", merchantNodeURL)
			err = merchantNode.Connect()
			if err != nil {
				log.Printf("Connect merchant node faild unexpected error: %v. \n", err)
				disconnected <- struct{}{}
			} else {
				log.Printf("Connect merchant node successfully. \n")
			}
		case <-disconnected:
			if isReconnect {
				//重新连接，前等待
				log.Printf("Reconnect after %d seconds... \n", reconnectWait)
				time.Sleep(reconnectWait * time.Second)
				reconnect <- true
			} else {
				//退出
				break
			}
		}
	}

	return nil
}

/********** 商户服务相关方法【主动】 **********/

func GetChargeAddress() {

	var (
		completed bool
	)

	if merchantNode == nil {
		return
	}

	for {
		//获取地址
		merchantNode.Call(
			"getChargeAddress",
			nil,
			true,
			func(resp owtp.Response) {
				if resp.Status == 0 {
					completed = true
				}
			})

		if completed == true {
			break
		}
	}
}

/********** 钱包管理相关方法【被动】 **********/

//setupRouter 配置路由
func setupRouter(node *owtp.OWTPNode) {

	node.HandleFunc("subscribe", subscribe)
	node.HandleFunc("configWallet", configWallet)
	node.HandleFunc("getWalletInfo", getWalletInfo)
	node.HandleFunc("submitTrasaction", submitTrasaction)
}

//subscribe 订阅方法
func subscribe(ctx *owtp.Context) {

	log.Printf("Merchat Call: subscribe \n")
	log.Printf("params: %v\n", ctx.Params())

	var (
		subscriptions []*Subscription
	)

	db, err := openDB()
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

func configWallet(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func getWalletInfo(ctx *owtp.Context) {

	responseSuccess(ctx, nil)
}

func submitTrasaction(ctx *owtp.Context) {

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
