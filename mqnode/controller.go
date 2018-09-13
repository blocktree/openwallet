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

package mqnode

import (
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/owtp"
	"github.com/pkg/errors"
	"github.com/blocktree/OpenWallet/manager"
	"strconv"
	"encoding/json"
)

const (
	//订阅类型，1：钱包余额，2：充值记录，3：汇总日志
	SubscribeTypeBalance    = 1
	SubscribeTypeCharge     = 2
	SubscribeTypeSummaryLog = 3
)

var (
	//商户节点
	merchantNode *BitBankNode

	/* 异常错误 */

	//节点断开
	ErrMerchantNodeDisconnected = errors.New("Merchant node is not connected!")
)

/********** 钱包管理相关方法【被动】 **********/

//setupRouter 配置路由
func (m *BitBankNode) setupRouter() {

	m.Node.HandleFunc("subscribe", m.subscribe)
	m.Node.HandleFunc("createWallet", m.createWallet)
	m.Node.HandleFunc("createAddress", m.createAddress)
	m.Node.HandleFunc("getAddressList", m.getAddressList)
	m.Node.HandleFunc("createAssetsAccount", m.createAssetsAccount)
	m.Node.HandleFunc("getWalletList", m.getWalletList)
	m.Node.HandleFunc("submitTransaction", m.submitTransaction)
	m.Node.HandleFunc("getNewHeight", m.getNewHeight)
	m.Node.HandleFunc("getBalanceByAddress", m.getBalanceByAddress)
	m.Node.HandleFunc("getWalletBalance", m.getWalletBalance)
	m.Node.HandleFunc("rescanBlockHeight", m.rescanBlockHeight)

}

//subscribe 订阅方法
/**
 * appID
 */
func (m *BitBankNode) subscribe(ctx *owtp.Context) {

	log.Info("Merchant handle: subscribe")
	log.Info("params:", ctx.Params())

	/*
	| 参数名称 | 类型   | 是否可空 | 描述                                                                         |
	|----------|--------|----------|------------------------------------------------------------------------------|
	| type     | int    | 否       | 订阅类型，1：钱包余额，2：交易记录，3：充值记录(未花交易)，4：提现记录（未花消费记录） |
	| symbol   | string | 否       | 订阅的币种钱包类型                                                           |
	| appID    | string | 否       | 钱包应用id                                                                   |
	| walletID | string | 否       | 钱包id                                                                       |
	*/

	var (
		subscriptions []*Subscription
	)

	for _, p := range ctx.Params().Get("subscriptions").Array() {
		s := NewSubscription(p)
		subscriptions = append(subscriptions, s)
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)
	err := Subscribe(subscriptions)
	if err != nil {
		responseError(ctx, errors.New("subscriptions error"))
		return
	}
	responseSuccess(ctx, nil)
}

//importWatchOnlyAddress 导入资产账户订阅地址
func (m *BitBankNode) importWatchOnlyAddress(ctx *owtp.Context) {

	log.Info("Merchant handle: importWatchOnlyAddress")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型     | 是否可空 | 描述           |
		|-----------|----------|----------|----------------|
		| appID     | string   | 否       | 钱包应用id     |
		| walletID  | string   | 否       | 钱包id         |
		| accountID | string   | 否       | 资产账户id     |
		| addresses | [string] | 否       | 导入的地址数组 |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	addresses := ctx.Params().Get("addresses").Array()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}

	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}

	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}

	if len(addresses) == 0 {
		responseError(ctx, errors.New("addresses is empty"))
		return
	}
	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)
	err := ow.ImportWatchOnlyAddress(appID, walletID, accountID , nil)
	if err != nil {
		responseError(ctx, errors.New("getWatchOnlyAddressInfo error"))
		return
	}
	responseSuccess(ctx, nil)
}

//getWatchOnlyAddressInfo 获取资产账户已导入的地址统计信息
func (m *BitBankNode) getWatchOnlyAddressInfo(ctx *owtp.Context) {

	log.Info("Merchant handle: getWatchOnlyAddressInfo")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称     | 类型         | 是否可空 | 描述               |
		|--------------|--------------|----------|--------------------|
		| appID     | string       | 否       | 钱包应用id         |
		| walletID  | string       | 否       | 钱包id             |
		| accountID | string       | 否       | 资产账户id         |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}

	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}

	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}

//	config := manager.NewConfig()
//	ow := manager.NewWalletManager(config)
//	err := ow.ImportWatchOnlyAddress(appID, walletID, accountID , nil)
//	if err != nil {
//		responseError(ctx, errors.New("getWatchOnlyAddressInfo error"))
//		return
//	}
//	responseSuccess(ctx, creationWallet)
}


//### 3.4 创建钱包 `createWallet`
func (m *BitBankNode) createWallet(ctx *owtp.Context) {

	log.Info("Merchant handle: createWallet")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称                       | 类型   | 是否可空 | 描述                                          |
	|--------------------------------|--------|----------|-----------------------------------------------|
	| appID                          | string | 否       | 钱包应用id                                    |
	| alias                          | string | 否       | 钱包别名                                      |
	| isTrust                      | int    | 否       | 是否托管密钥，0：否，1：是               |
	| **isTrust = 1，以下字段必填** |        |          |                                               |
	| passwordType                   | int    | 否       | 0：自定义密码，1：协商密码                       |
	| password                       | string | 是       | 自定义密码                                    |
	| authKey                        | string | 是       | 授权公钥                                      |
	| **isTrust = 0，以下字段必填** |        |          |                                               |
	| walletID                       | string | 是       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码， |
	| rootPath                       | string | 是       | 钱包HD账户根路径，如：m/44'/88'                 |
	*/

	appID := ctx.Params().Get("appID").String()
	alias := ctx.Params().Get("alias").String()
	isTrust := ctx.Params().Get("isTrust").Int()
	//passwordType := ctx.Params().Get("passwordType").Int()
	password := ctx.Params().Get("password").String()
	authKey := ctx.Params().Get("authKey").String()
	walletID := ctx.Params().Get("walletID").String()
	rootPath := ctx.Params().Get("rootPath").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID  is empty"))
		return
	}

	if len(alias) == 0 {
		responseError(ctx, errors.New("alias  is empty"))
		return
	}

	if len(ctx.Params().Get("isTrust").String()) == 0 {
		responseError(ctx, errors.New("isTrust  is empty"))
		return
	}

	if isTrust == 0 {
		if len(walletID) == 0 {
			responseError(ctx, errors.New("walletID is empty"))
			return
		}
		if len(rootPath) == 0 {
			responseError(ctx, errors.New("rootPath is empty"))
			return
		}
	}else if isTrust == 1{
		if len(ctx.Params().Get("passwordType").String()) == 0  {
			responseError(ctx, errors.New("passwordType is empty"))
			return
		}
		if len(password) == 0 {
			responseError(ctx, errors.New("password is empty"))
			return
		}
		if len(authKey) == 0 {
			responseError(ctx, errors.New("authKey is empty"))
			return
		}
	}

	var isTrustBool bool
	if isTrust == 1{
		isTrustBool = true
	}else{
		isTrustBool = false
	}

	wallet := &openwallet.Wallet{
		AppID      : appID,
		WalletID    :  walletID,
		Alias       : alias,
		Password    : password,
		RootPath    : rootPath,
		IsTrust     : isTrustBool,
	}


	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)
	ow.Init()
	//执行创建方法
	creationWallet,keystore,err := ow.CreateWallet(appID,wallet)

	if err != nil {
		responseError(ctx, errors.New("createWallet error"))
		return
	}
	//h := &hdkeystore.HDKey{
	//
	//}

	result := map[string]interface{}{
		"wallet": creationWallet,
		"keystore":keystore,
	}
	responseSuccess(ctx, result)

}

//### 3.5 获取钱包信息 `getWalletInfo`
func (m *BitBankNode) getWalletInfo(ctx *owtp.Context) {

	log.Info("Merchant handle: getWalletInfo")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称 | 类型   | 是否可空 | 描述                                         |
		|----------|--------|----------|----------------------------------------------|
		| appID    | string | 否       | 钱包应用id                                   |
		| walletID | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码 |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID  is empty"))
		return
	}

	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID  is empty"))
		return
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)
	wallet,err := ow.GetWalletInfo(appID,walletID)

	if err != nil {
		responseError(ctx, errors.New("getWalletInfo error"))
		return
	}
	result := map[string]interface{}{
		"wallet": wallet,
	}
	responseSuccess(ctx, result)
}

//### 3.6 创建资产账户 `createAssetsAccount`
func (m *BitBankNode) createAssetsAccount(ctx *owtp.Context) {

	log.Info("Merchant handle: createAssetsAccount")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称                     | 类型   | 是否可空 | 描述                                                        |
		|------------------------------|--------|----------|-------------------------------------------------------------|
		| appID                        | string | 否       | 钱包应用id                                                  |
		| alias                        | string | 否       | 账户别名                                                    |
		| walletID                     | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码                |
		| symbol                       | string | 否       | 币种类型                                                    |
		| otherOwnerKeys               | string | 是       | 其他资产账户拥有者公钥数组，默认为空                         |
		| reqSigs                      | int    | 是       | 必要签名数，必须大于0，默认为1                                |
		| isTrust                      | int    | 否       | 是否托管密钥，0：否，1：是                                      |
		| **isTrust = 1，以下字段必填** |        |          |                                                             |
		| password                     | string | 否       | 解锁钱包密码                                                |
		| isTrust = 0，以下字段必填 |        |          |                                                             |
		| publicKey                    | string | 否       | 钱包分配的资产账户公钥，不可与已存在的重复，公钥地址编码，唯一 |
		| accountIndex                 | int    | 否       | 账户索引数，必须大于等于0，不可与已存在的重复                 |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	alias := ctx.Params().Get("alias").String()
	symbol := ctx.Params().Get("symbol").String()
	isTrustStr := ctx.Params().Get("isTrust").String()
	otherOwnerKeys := ctx.Params().Get("otherOwnerKeys").Array()
	//reqSigs := ctx.Params().Get("reqSigs").Int()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(alias) == 0 {
		responseError(ctx, errors.New("alias is empty"))
		return
	}
	if len(symbol) == 0 {
		responseError(ctx, errors.New("symbol is empty"))
		return
	}
	if len(isTrustStr) == 0 {
		responseError(ctx, errors.New("isTrust is empty"))
		return
	}

	isTrust,error := strconv.Atoi(isTrustStr)
	if error != nil{
		responseError(ctx, errors.New("isTrust must be a number"))
		return
	}


	password := ""
	publicKey := ""
	var accountIndex uint64

	if isTrust == 1{
		password = ctx.Params().Get("password").String()
		if len(password) == 0 {
			responseError(ctx, errors.New("password is empty"))
			return
		}
	}else if isTrust == 0{
		publicKey = ctx.Params().Get("publicKey").String()
		if len(publicKey) == 0 {
			responseError(ctx, errors.New("publicKey is empty"))
			return
		}
		accountIndex = ctx.Params().Get("accountIndex").Uint()
		if accountIndex == 0 {
			responseError(ctx, errors.New("accountIndex is empty"))
			return
		}
	}else{
		responseError(ctx, errors.New("isTrust must be a 1 or 0"))
		return
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)
	ow.Init()
	//创建 assetsAccount
	var assetsAccount *openwallet.AssetsAccount
	err := json.Unmarshal([]byte(ctx.Params().Raw), &assetsAccount)
	if err != nil {
		responseError(ctx, errors.New("can't unmarshal to AssetsAccount"))
		return
	}

	//创建 otherOwnerKeysList
	otherOwnerKeysList := make([]string, 0)
	for _,v := range otherOwnerKeys{
		otherOwnerKeysList = append(otherOwnerKeysList, v.Str)
	}

	newAssetsAccount,err := ow.CreateAssetsAccount(appID, walletID , assetsAccount,otherOwnerKeysList )

	if err != nil {
		responseError(ctx, errors.New("createAssetsAccount error"))
		return
	}
	//result := map[string]interface{}{
	//	"wallet": newAssetsAccount,
	//}

	responseSuccess(ctx, newAssetsAccount)
}


// ### 3.7 获取资产账户 `getAssetsAccountInfo`
func (m *BitBankNode) getAssetsAccountInfo(ctx *owtp.Context) {

	log.Info("Merchant handle: getWalletList")
	log.Info("params:", ctx.Params())

	/*
	| 参数名称  | 类型   | 是否可空 | 描述                                         |
	|-----------|--------|----------|----------------------------------------------|
	| appID     | string | 否       | 钱包应用id                                   |
	| walletID  | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码 |
	| accountID | string | 否       | 资产账户ID                                   |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	account,err := ow.GetAssetsAccountInfo(appID,walletID,accountID)

	if err != nil{
		responseError(ctx, errors.New("getAssetsAccountInfo error"))
		return
	}

	//result := map[string]interface{}{
	//	"wallets": walletsMaps,
	//}

	responseSuccess(ctx, account)
}

//### 3.8 创建地址 `createAddress`
func (m *BitBankNode) createAddress(ctx *owtp.Context) {

	log.Info("Merchant handle: createAddress")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string | 否       | 钱包应用id                                   |
		| walletID  | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码 |
		| accountID | string | 否       | 资产账户ID                                   |
		| count     | string | 否       | 创建数量                                     |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	countStr := ctx.Params().Get("count").String()
	count := ctx.Params().Get("count").Uint()
	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}
	if len(countStr) == 0 {
		responseError(ctx, errors.New("countStr is empty"))
		return
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	addressList,err := ow.CreateAddress(appID,walletID,accountID,count)
	if err != nil{
		responseError(ctx, errors.New("createAddress error"))
		return
	}

	result := map[string]interface{}{
		"addresses": addressList,
	}

	responseSuccess(ctx, result)
}


//### 3.9 获取地址列表 `getAddressList`
func (m *BitBankNode) getAddressList(ctx *owtp.Context) {

	log.Info("Merchant handle: getAddressList")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string | 否       | 钱包应用id                                   |
		| walletID  | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码 |
		| accountID | string | 否       | 资产账户ID                                   |
		| offset    | int    | 是       | 从0开始                                      |
		| limit     | int    | 是       | 查询条数                                     |
		| watchOnly | int    | 否       | 观察类型，只做订阅使用，0：否，1：是              |
	*/
	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	offset := ctx.Params().Get("offset").Int()
	limit := ctx.Params().Get("limit").Int()
	watchOnly := ctx.Params().Get("watchOnly").Int()
	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}
	if len(ctx.Params().Get("watchOnly").String()) == 0 {
		responseError(ctx, errors.New("watchOnly is empty"))
		return
	}
	watchOnlyBool := false
	if watchOnly == 1{
		watchOnlyBool = true
	}else if watchOnly == 0{
		watchOnlyBool = false
	}else{
		responseError(ctx, errors.New("watchOnly must be 0 or 1"))
		return
	}
	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	addresses,err := ow.GetAddressList(appID,walletID,accountID,int(offset),int(limit),watchOnlyBool)
	if err != nil{
		responseError(ctx, errors.New("getAddressList error"))
		return
	}

	result := map[string]interface{}{
		"addresses": addresses,
	}

	responseSuccess(ctx, result)
}

//3.10 获取钱包列表 `getWalletList`
func (m *BitBankNode) getWalletList(ctx *owtp.Context) {

	log.Info("Merchant handle: getWalletList")
	log.Info("params:", ctx.Params())

	/*
	| 参数名称  | 类型   | 是否可空 | 描述                                         |
	|-----------|--------|----------|----------------------------------------------|
	| appID    | string | 否       | 钱包应用id |
	| offset   | int    | 是       | 从0开始    |
	| limit    | int    | 是       | 查询条数   |
	*/


	appID := ctx.Params().Get("appID").String()
	offset := ctx.Params().Get("walletID").Int()
	limit := ctx.Params().Get("accountID").Int()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}

	if offset == 0{
		offset = 0
	}
	if limit == 0{
		limit = 20
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	walletList ,err := ow.GetWalletList(appID,int(offset),int(limit))
	if err != nil{
		responseError(ctx, errors.New("getWalletList error"))
		return
	}

	result := map[string]interface{}{
		"wallets": walletList,
	}

	responseSuccess(ctx, result)
}


//### 3.11 获取资产账户列表 `getAssetsAccountList`
func (m *BitBankNode) getAssetsAccountList(ctx *owtp.Context) {

	log.Info("Merchant handle: getAssetsAccountList")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID    | string | 否       | 钱包应用id                                   |
		| walletID | string | 否       | 钱包ID，由钱包种子哈希，openwallet钱包地址编码 |
		| offset   | int    | 是       | 从0开始                                      |
		| limit    | int    | 是       | 查询条数                                     |
	*/
	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	offset := ctx.Params().Get("walletID").Int()
	limit := ctx.Params().Get("accountID").Int()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}

	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if offset == 0{
		offset = 0
	}
	if limit == 0{
		limit = 20
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	walletList ,err := ow.GetAssetsAccountList(appID,walletID,int(offset),int(limit))
	if err != nil{
		responseError(ctx, errors.New("getAssetsAccountList error"))
		return
	}

	result := map[string]interface{}{
		"accounts": walletList,
	}

	responseSuccess(ctx, result)
}

//### 3.12 创建转账交易 `createTransaction`
func (m *BitBankNode) createTransaction(ctx *owtp.Context) {

	log.Info("Merchant handle: createTransaction")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string     | 否       | 钱包应用id，bitbank是一个App     |
		| coin      | CoinInfo | 否       | 转账币种信息                    |
		| accountID | string     | 否       | 资产账户ID                      |
		| amount    | string     | 否       | 转账数量                        |
		| address   | string     | 否       | 地址                            |
		| feeRate   | string     | 否       | 自定服务费率， fees/K            |
		| memo      | string     | 是       | 备注                            |
	*/

	appID := ctx.Params().Get("appID").String()
	coin := ctx.Params().Get("coin").Map()
	accountID := ctx.Params().Get("accountID").String()
	amount := ctx.Params().Get("amount").String()
	address := ctx.Params().Get("address").String()
	feeRate := ctx.Params().Get("feeRate").String()
	memo := ctx.Params().Get("memo").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}
	if len(amount) == 0 {
		responseError(ctx, errors.New("amount is empty"))
		return
	}
	if len(address) == 0 {
		responseError(ctx, errors.New("address is empty"))
		return
	}
	if len(feeRate) == 0 {
		responseError(ctx, errors.New("feeRate is empty"))
		return
	}
	if len(memo) == 0 {
		responseError(ctx, errors.New("memo is empty"))
		return
	}
	if coin == nil ||  len(coin) == 0 {
		responseError(ctx, errors.New("coin is empty"))
		return
	}
	if _,ok := coin["symbol"]; !ok {
		responseError(ctx, errors.New("symbol is empty"))
		return
	}
	if _,ok := coin["isContract"]; !ok {
		responseError(ctx, errors.New("isContract is empty"))
		return
	}
	if _,ok := coin["contractID"]; !ok {
		responseError(ctx, errors.New("contractID is empty"))
		return
	}

	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	rawTransaction ,err := ow.CreateTransaction(appID, accountID, amount, address, feeRate, memo)
	if err != nil{
		responseError(ctx, errors.New("createTransaction error"))
		return
	}

	result := map[string]interface{}{
		"rawTx": rawTransaction,
	}

	responseSuccess(ctx, result)
}

//### 3.13 广播转账交易 `submitTransaction`
func (m *BitBankNode) submitTransaction(ctx *owtp.Context) {

	log.Info("Merchant handle: submitTransaction")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string     | 否       | 钱包应用id，bitbank是一个App     |
		| rawTx     | RawTransaction | 否       | 未完成的交易单              |
	*/

	appID := ctx.Params().Get("appID").String()
	rawTx := ctx.Params().Get("rawTx").Raw

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	var raw openwallet.RawTransaction

	err := json.Unmarshal([]byte(rawTx), &raw)
	if err != nil {
		responseError(ctx, errors.New("can't unmarshal to RawTransaction"))
		return
	}


	config := manager.NewConfig()
	ow := manager.NewWalletManager(config)

	transaction ,err := ow.SubmitTransaction(appID,raw.Account.AccountID,raw)
	if err != nil{
		responseError(ctx, errors.New("submitTransaction error"))
		return
	}

	result := map[string]interface{}{
		"tx": transaction,
	}

	responseSuccess(ctx, result)
}


//### 3.14 提交转账交易 `sendTransaction`
func (m *BitBankNode) sendTransaction(ctx *owtp.Context) {

	log.Info("Merchant handle: sendTransaction")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID           | string               | 否       | 钱包应用id，bitbank是一个App       |
		| walletID        | string               | 否       | 钱包id，必须是托管密码的钱包       |
		| withdraws       | map[string: Object]] | 否       | 所需提币条目 （每个为一个withdraw） |
		| -> address      | string               | 否       | 目标地址                       |
		| -> withdraw     | Object               | 否       | 所需提币条目 （每个为一个withdraw） |
		| -> -> accountID | string               | 否       | 资产账户ID                        |
		| -> -> sid       | string               | 否       | 安全id（防止重复提交）              |
		| -> -> isMemo    | int                  | 否       | 1为memo，0为address                |
		| -> -> address   | string               | 否       | 目标地址                        |
		| -> -> amount    | string               | 否       | 提笔金额                          |
		| -> -> memo      | string               | 是       | 备注                              |
		| -> -> password  | string               | 是       | 解锁钱包密码                      |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	withdraws := ctx.Params().Get("withdraws").Map()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if withdraws == nil || len(withdraws) == 0 {
		responseError(ctx, errors.New("withdraws is empty"))
		return
	}

	withdrawList := make([]*openwallet.Withdraw,0)

	for _,v := range withdraws{
		var w *openwallet.Withdraw
		err := json.Unmarshal([]byte(v.Raw), &w)
		if err != nil {
			continue
		}
		withdrawList = append(withdrawList,w)
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)

	//if err != nil{
	//	responseError(ctx, errors.New("submitTransaction error"))
	//	return
	//}
	//
	//result := map[string]interface{}{
	//	"tx": transaction,
	//}

	responseSuccess(ctx, withdrawList)
}

//### 3.15 获取资产账户余额 `getAssetsAccountBalance`
func (m *BitBankNode) getAssetsAccountBalance(ctx *owtp.Context) {

	log.Info("Merchant handle: getAssetsAccountBalance")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string | 否       | 钱包应用id，bitbank是一个App |
		| walletID  | string  | 否       | 钱包ID                      |
		| accountID | string | 否       | 账户ID                      |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}


	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)
	//ow.GetAssetsAccountInfo()
	//if err != nil{
	//	responseError(ctx, errors.New("getAssetsAccountBalance error"))
	//	return
	//}
	//
	//result := map[string]interface{}{
	//	"tx": transaction,
	//}

	responseSuccess(ctx, nil)
}

//### 3.16 通过地址获取余额 `getAddressBalance`
func (m *BitBankNode) getAddressBalance(ctx *owtp.Context) {

	log.Info("Merchant handle: getAddressBalance")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string  | 否       | 钱包应用id，bitbank是一个App |
		| walletID  | string  | 否       | 钱包ID                      |
		| accountID | string  | 否       | 账户ID                      |
		| address   | Balance | 否       | 地址                        |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	address := ctx.Params().Get("address").String()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}
	if len(address) == 0 {
		responseError(ctx, errors.New("address is empty"))
		return
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)


	//ow.GetAssetsAccountInfo()
	//if err != nil{
	//	responseError(ctx, errors.New("getAssetsAccountBalance error"))
	//	return
	//}
	//
	//result := map[string]interface{}{
	//	"tx": transaction,
	//}

	responseSuccess(ctx, nil)
}

//### 3.17 获取交易记录 `getTransactions`
func (m *BitBankNode) getTransactions(ctx *owtp.Context) {

	log.Info("Merchant handle: getTransactions")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string | 否       | 钱包应用id，bitbank是一个App |
		| walletID | string | 否       | 钱包ID                      |
		| accountID | string | 否       | 账户ID                      |
		| offset    | int    | 是       | 从0开始                     |
		| limit     | int    | 是       | 查询条数                    |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	//offset := ctx.Params().Get("offset").Int()
	//limit := ctx.Params().Get("limit").Int()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)

	responseSuccess(ctx, nil)
}

//### 3.18 推送订阅数据 `pushNotifications`
func (m *BitBankNode) pushNotifications(ctx *owtp.Context) {

	log.Info("Merchant handle: pushNotifications")
	log.Info("params:", ctx.Params())

	/*
| 参数名称  | 类型                            | 是否可空 | 描述                                      |
|-----------|---------------------------------|----------|-------------------------------------------|
| appID     | string                          | 否       | 钱包应用id，bitbank是一个App               |
| walletID  | string                          | 否       | 钱包ID                                    |
| accountID | string                          | 否       | 账户ID                                    |
| dataType  | int                             | 否       | 数据类型：1：钱包余额，2：交易记录，3：充值记录(未花交易)，4：提现记录（未花消费记录） |
| content   | [Transaction]/Balance/[Unspent] | 否       | 根据数据类型，返回数据主体                 |

#### 未花交易 `Unspent`

| 参数名称      | 类型   | 是否可空 | 描述                      |
|---------------|--------|----------|---------------------------|
| txid          | string | 否       | 唯一交易单号              |
| symbol        | string | 否       | 币种类型                  |
| vout          | int    | 否       | 输出位置                  |
| accountID     | string | 否       | 资产账户ID                |
| address       | string | 否       | 地址                      |
| amount        | string | 否       | 未花数量                  |
| confirmations | int    | 否       | 确认次数                  |
| spendable     | int    | 否       | 是否可用， 0：不可用，1：可用 |
| confirmTime   | int    | 否       | 确认时间                  |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	//dataType := ctx.Params().Get("dataType").Int()
	//content := ctx.Params().Get("content").Array()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}
	if len(ctx.Params().Get("dataType").String()) == 0 {
		responseError(ctx, errors.New("dataType is empty"))
		return
	}
	if len(ctx.Params().Get("content").String()) == 0 {
		responseError(ctx, errors.New("content is empty"))
		return
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)

	responseSuccess(ctx, nil)
}

//### 3.19 获取资产账户代币余额 `getAssetsAccountTokens`
func (m *BitBankNode) getAssetsAccountTokens(ctx *owtp.Context) {

	log.Info("Merchant handle: getAssetsAccountTokens")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称  | 类型   | 是否可空 | 描述                                         |
		|-----------|--------|----------|----------------------------------------------|
		| appID     | string | 否       | 钱包应用id，bitbank是一个App |
		| walletID  | string  | 否       | 钱包ID                      |
		| accountID | string | 否       | 账户ID                      |
		| offset    | int    | 是       | 从0开始                     |
		| limit     | int    | 是       | 查询条数                    |
	*/

	appID := ctx.Params().Get("appID").String()
	walletID := ctx.Params().Get("walletID").String()
	accountID := ctx.Params().Get("accountID").String()
	//offset := ctx.Params().Get("offset").Int()
	//limit := ctx.Params().Get("limit").Int()

	if len(appID) == 0 {
		responseError(ctx, errors.New("appID is empty"))
		return
	}
	if len(walletID) == 0 {
		responseError(ctx, errors.New("walletID is empty"))
		return
	}
	if len(accountID) == 0 {
		responseError(ctx, errors.New("accountID is empty"))
		return
	}

	//config := manager.NewConfig()
	//ow := manager.NewWalletManager(config)

	responseSuccess(ctx, nil)
}



func (m *BitBankNode) getNewHeight(ctx *owtp.Context) {

	log.Info("Merchant handle: getNewHeight")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称 | 类型   | 是否可空 | 描述     |
		|----------|--------|----------|----------|
		| coin     | string | 否       | 币种标识 |
	*/

	coin := ctx.Params().Get("coin").String()
	//walletID := ctx.Params().Get("walletID").String()

	am := assets.GetMerchantAssets(coin)
	blockchain, err := am.GetBlockchainInfo()
	if err != nil {
		responseError(ctx, err)
		return
	}

	result := map[string]interface{}{
		"coin":      coin,
		"cmdHeight": blockchain.ScanHeight,
		"height":    blockchain.Blocks,
	}

	responseSuccess(ctx, result)
}

func (m *BitBankNode) getWalletBalance(ctx *owtp.Context) {

	log.Info("Merchant handle: getWalletBalance")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称     | 类型   | 是否可空 | 描述   |
		|--------------|--------|----------|--------|
		| coin         | string | 否       | 币名   |
		| walletID     | string | 否       | 钱包ID |
	*/

	coin := ctx.Params().Get("coin").String()
	walletID := ctx.Params().Get("walletID").String()

	am := assets.GetMerchantAssets(coin)
	balance, err := am.GetMerchantWalletBalance(walletID)
	if err != nil {
		responseError(ctx, err)
		return
	}

	result := map[string]interface{}{
		"balance": balance,
	}

	responseSuccess(ctx, result)
}

func (m *BitBankNode) getBalanceByAddress(ctx *owtp.Context) {

	log.Info("Merchant handle: getBalanceByAddress")
	log.Info("params:", ctx.Params())

	/*
		| 参数名称     | 类型   | 是否可空 | 描述   |
		|--------------|--------|----------|--------|
		| coin         | string | 否       | 币名   |
		| walletID     | string | 否       | 钱包ID |
		| address      | string  | 否        | 地址 |
	*/

	coin := ctx.Params().Get("coin").String()
	walletID := ctx.Params().Get("walletID").String()
	address := ctx.Params().Get("address").String()

	am := assets.GetMerchantAssets(coin)
	balance, err := am.GetMerchantAddressBalance(walletID, address)
	if err != nil {
		responseError(ctx, err)
		return
	}

	result := map[string]interface{}{
		"balance": balance,
	}

	responseSuccess(ctx, result)
}

func (m *BitBankNode) rescanBlockHeight(ctx *owtp.Context) {
	log.Info("Merchant handle: rescanBlockHeight")
	log.Info("params:", ctx.Params())

	/*
	| 参数名称    | 类型   | 是否可空 | 描述                   |
	|-------------|--------|----------|------------------------|
	| coin        | string | 否       | 币名                   |
	| startHeight | string | 否       | 起始高度               |
	| endHeight   | string | 否       | 结束高度，0代表最新高度 |
	*/

	var (
		err error
	)

	coin := ctx.Params().Get("coin").String()
	//walletID := ctx.Params().Get("walletID").String()
	startHeight := ctx.Params().Get("startHeight").Uint()
	endHeight := ctx.Params().Get("endHeight").Uint()

	am := assets.GetMerchantAssets(coin)
	if endHeight == 0 {
		err = am.SetMerchantRescanBlockHeight(startHeight)
	} else {
		err = am.MerchantRescanBlockHeight(startHeight, endHeight)
	}

	if err != nil {
		responseError(ctx, err)
		return
	}

	responseSuccess(ctx, nil)
}

//responseSuccess 成功结果响应
func responseSuccess(ctx *owtp.Context, result interface{}) {
	ctx.Response(result, owtp.StatusSuccess, "success")
	log.Info(ctx.Method, ":", "Response:", ctx.Resp.JsonData())
}

//responseError 失败结果响应
func responseError(ctx *owtp.Context, err error) {
	ctx.Response(nil, owtp.ErrCustomError, err.Error())
	log.Info(ctx.Method, ":", "Response:", ctx.Resp.JsonData())
}
