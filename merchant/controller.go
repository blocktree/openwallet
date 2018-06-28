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
	"log"
)

var (
	merchantNode *owtp.OWTPNode
)

func init() {

	//TODO:读取配置文件，加载商户连接，商户公钥，本地私钥

}

/********** 商户服务相关方法【主动】 **********/





/********** 钱包管理相关方法【被动】 **********/

//setupRouter 配置路由
func setupRouter(node *owtp.OWTPNode) {

	node.HandleFunc("subscribe", subscribe)
	node.HandleFunc("configWallet", configWallet)
	node.HandleFunc("getWalletInfo", getWalletInfo)
	node.HandleFunc("submitTrasaction", submitTrasaction)
}


func subscribe(ctx *owtp.Context) {

	log.Printf("call subscribe \n")
	log.Printf("params: %v\n", ctx.Params)


	var (
		status uint64 = 0
		msg string = "success"
		result  = make(map[string]interface{})
	)

	result["good"] = "hello"

	resp := owtp.Response{
		Status: status,
		Msg:    msg,
		Result: result,
	}

	ctx.Resp = resp
}

func configWallet(ctx *owtp.Context) {

}

func getWalletInfo(ctx *owtp.Context) {

}

func submitTrasaction(ctx *owtp.Context) {

}
