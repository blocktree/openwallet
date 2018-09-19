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
	"github.com/blocktree/OpenWallet/hdkeystore"
	"github.com/blocktree/OpenWallet/openwallet"
	// "path/filepath"
	// "github.com/astaxie/beego/config"
	// "github.com/pkg/errors"
	// "github.com/shopspring/decimal"
)

type WalletManager struct {
	Config         *WalletConfig                 //钱包管理配置
	Storage        *hdkeystore.HDKeystore        //秘钥存取
	Blockscanner   *TronBlockScanner             //区块扫描器
	FullnodeClient *Client                       // 全节点客户端
	WalletClient   *Client                       // 节点客户端
	WalletsInSum   map[string]*openwallet.Wallet //参与汇总的钱包
}

func NewWalletManager() *WalletManager {
	wm := WalletManager{}
	wm.Config = NewConfig(Symbol, MasterKey)
	wm.Storage = hdkeystore.NewHDKeystore(wm.Config.keyDir, hdkeystore.StandardScryptN, hdkeystore.StandardScryptP)
	//参与汇总的钱包
	wm.WalletsInSum = make(map[string]*openwallet.Wallet)
	//区块扫描器
	wm.Blockscanner = NewTronBlockScanner(&wm)
	// wm.Decoder = &addressDecoder{}
	// wm.TxDecoder = NewTransactionDecoder(&wm)
	return &wm
}

// wallet/votewitnessaccount Function：Vote on the super representative demo：curl -X POSThttp://127.0.0.1:8090/wallet/votewitnessaccount -d ‘{ “owner_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”, “votes”: [{“vote_address”: “41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”, “vote_count”: 5}] }’ Parameters：Owner_address is the voter address, converted to a hex string; votes.vote_address is the address of the super delegate being voted, converted to a hex string; vote_count is the number of votes
// wallet/createassetissue Function：Issue Token demo：curl -X POSThttp://127.0.0.1:8090/wallet/createassetissue -d ‘{ “owner_address”:””, “name”:”{{assetIssueName}}”, “abbr”: “{{abbrName}}”, “total_supply” :4321, “trx_num”:1, “num”:1, “start_time” :{{startTime}}, “end_time”:{{endTime}}, “vote_score”:2, “description”:”007570646174654e616d6531353330363038383733343633”, “url”:”007570646174654e616d6531353330363038383733343633”, “free_asset_net_limit”:10000, “public_free_asset_net_limit”:10000, “frozen_supply”:{“frozen_amount”:1, “frozen_days”:2} }’
// wallet/createwitness Function：Apply to become a super representative demo：curl -X POSThttp://127.0.0.1:8090/wallet/createwitness -d ‘{“owner_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”, “url”: “007570646174654e616d6531353330363038383733343633”}’ Parameters：owner_address is the account address of the applicant，converted to a hex string；url is the official website address，converted to a hex string Return value：Super Representative application Transaction raw data
// wallet/transferasset Function：Transfer Token demo：curl -X POSThttp://127.0.0.1:8090/wallet/transferasset -d ‘{“owner_address”:”41d1e7a6bc354106cb410e65ff8b181c600ff14292”, “to_address”: “41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”, “asset_name”: “0x6173736574497373756531353330383934333132313538”, “amount”: 100}’ Parameters：Owner_address is the address of the withdrawal account, converted to a hex string；To_address is the recipient address，converted to a hex string；asset_name is the token contract address，converted to a hex string；Amount is the amount of token to transfer Return value：Token transfer Transaction raw data
// wallet/easytransfer Function: Easily transfer from an address using the password string. Only works with accounts created from createAddress Demo: curl -X POST http://127.0.0.1:8090/wallet/easytransfer -d ‘{“passPhrase”: “7465737470617373776f7264”,”toAddress”: “41D1E7A6BC354106CB410E65FF8B181C600FF14292”, “amount”:10}’ Parameters: passPhrase is the password, converted from ascii to hex. toAddress is the recipient address, converted into a hex string; amount is the amount of TRX ‘to transfer expressed in SUN. Warning: Please control risks when using this API. To ensure environmental security, please do not invoke APIs provided by other or invoke this very API on a public network.
// wallet/easytransferbyprivate Function：Easily transfer from an address using the private key. demo: curl -X POST http://127.0.0.1:8090/wallet/easytransferbyprivate -d ‘{“privateKey”: “D95611A9AF2A2A45359106222ED1AFED48853D9A44DEFF8DC7913F5CBA727366”, “toAddress”:”4112E621D5577311998708F4D7B9F71F86DAE138B5”,”amount”:10000}’ Parameters：passPhrase is the private key in hex string format. toAddress is the recipient address, converted into a hex string; amount is the amount of TRX to transfer in SUN. Return value： transaction, including execution results. Warning: Please control risks when using this API. To ensure environmental security, please do not invoke APIs provided by other or invoke this very API on a public network.

// wallet/participateassetissue Function：Create a new Token demo：curl -X POST http://127.0.0.1:8090/wallet/participateassetissue -d ‘{ “to_address”: “41e552f6487585c2b58bc2c9bb4492bc1f17132cd0”, “owner_address”:”41e472f387585c2b58bc2c9bb4492bc1f17342cd1”, “amount”:100, “asset_name”:”3230313271756265696a696e67” }’ Parameters： to_address is the address of the Token issuer，converted to a hex string owner_address is the address of the Token owner，converted to a hex string amount is the number of tokens created asset_name is the name of the token，converted to a hex string Return value：Token creation Transaction raw data
// wallet/freezebalance Function：Freezes an amount of TRX. Will give bandwidth and TRON Power(voting rights) to the owner of the frozen tokens. demo：curl -X POST http://127.0.0.1:8090/wallet/freezebalance -d ‘{ “owner_address”:”41D1E7A6BC354106CB410E65FF8B181C600FF14294”, “frozen_balance”: 10000, “frozen_duration”: 3 }’ Parameters： owner_address is the address that is freezing trx account，converted to a hex string frozen_balance is the number of frozen TRX frozen_duration is the duration in days to be frozen Return value：Freeze trx transaction raw data
// wallet/unfreezebalance Function：Unfreeze TRX that has passed the minimum freeze duration. Unfreezing will remove bandwidth and TRON Power. demo：curl -X POST http://127.0.0.1:8090/wallet/unfreezebalance -d ‘{ “owner_address”:”41e472f387585c2b58bc2c9bb4492bc1f17342cd1”, }’ Parameters： owner_address address to unfreeze TRX for，converted to a hex string Return value：Unfreeze TRX transaction raw data
// wallet/unfreezeasset Function：Unfreeze a token that has passed the minimum freeze duration. demo：curl -X POST http://127.0.0.1:8090/wallet/unfreezeasset -d ‘{ “owner_address”:”41e472f387585c2b58bc2c9bb4492bc1f17342cd1”, }’ Parameters： owner_address address to unfreeze Tokens for，converted to a hex string Return value：Unfreeze Token transaction raw data
// wallet/withdrawbalance Function：Withdraw Super Representative rewards, useable every 24 hours. demo：curl -X POST http://127.0.0.1:8090/wallet/withdrawbalance -d ‘{ “owner_address”:”41e472f387585c2b58bc2c9bb4492bc1f17342cd1”, }’ Parameters： owner_address is the address to withdraw from，converted to a hex string Return value：Withdraw TRX transaction raw data
// wallet/updateasset Function：Update a Token’s information demo：curl -X POST http://127.0.0.1:8090/wallet/updateasset -d ‘{ “owner_address”:”41e472f387585c2b58bc2c9bb4492bc1f17342cd1”, “description”: “”， “url”: “”, “new_limit” : 1000000, “new_public_limit” : 100 }’ Parameters： Owner_address is the address of the token issuer, converted to a hex string Description is a description of the token, converted to a hex string Url is the official website address of the token issuer, converted to a hex string New_limit is the free bandwidth that each token can use for each holder. New_public_limit is the free bandwidth of the token Return value: Token update transaction raw data
// wallet/listnodes Function：List the nodes which the api fullnode is connecting on the network demo: curl -X POST http://127.0.0.1:8090/wallet/listnodes Parameters：None Return value：List of nodes
// wallet/getassetissuebyaccount Function：List the tokens issued by an account. demo: curl -X POSThttp://127.0.0.1:8090/wallet/getassetissuebyaccount -d ‘{“address”: “41F9395ED64A6E1D4ED37CD17C75A1D247223CAF2D”}’ Parameters：Token issuer account address，converted to a hex string Return value：List of tokens issued by the account

// wallet/getassetissuebyname Function：Query token by name. demo: curl -X POSThttp://127.0.0.1:8090/wallet/getassetissuebyname -d ‘{“value”: “44756354616E”}’ Parameters：The name of the token, converted to a hex string Return value：token.

// wallet/listwitnesses Function：Query the list of Super Representatives demo: curl -X POSThttp://127.0.0.1:8090/wallet/listwitnesses Parameters：None Return value：List of all Super Representatives
// wallet/getassetissuelist Function：Query the list of Tokens demo: curl -X POSThttp://127.0.0.1:8090/wallet/getassetissuelist Parameters：None Return value：List of all Tokens
// wallet/getpaginatedassetissuelist Function：Query the list of Tokens with pagination demo: curl -X POSThttp://127.0.0.1:8090/wallet/getpaginatedassetissuelist -d ‘{“offset”: 0, “limit”: 10}’ Parameters：Offset is the index of the starting Token, and limit is the number of Tokens expected to be returned. Return value：List of Tokens

// wallet/getnextmaintenancetime Function：Get the time of the next Super Representative vote demo: curl -X POST http://127.0.0.1:8090/wallet/getnextmaintenancetime Parameters：None Return value: number of milliseconds until the next voting time.
