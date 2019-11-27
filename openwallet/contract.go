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

package openwallet

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/blocktree/openwallet/crypto"
)

type SmartContract struct {
	ContractID string `json:"contractID" storm:"id"` //计算ID：base64(sha256({symbol}_{address})) 主链symbol
	Symbol     string `json:"symbol"`                //主币的symbol
	Address    string `json:"address"`
	Token      string `json:"token"` //合约的symbol
	Protocol   string `json:"protocol"`
	Name       string `json:"name"`
	Decimals   uint64 `json:"decimals"`
}

//GenContractID 合约ID
func GenContractID(symbol, address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("%v_%v", symbol, address))))
}

// SmartContractRawTransaction 智能合约原始交易单
type SmartContractRawTransaction struct {
	Symbol      string                     `json:"symbol"`     //@required 区块链类型标识
	TxID        string                     `json:"txID"`       //交易单ID，广播后会生成
	Sid         string                     `json:"sid"`        //@required 业务订单号，保证业务不重复交易而用
	RawHex      string                     `json:"rawHex"`     //@required 区块链协议构造的交易原生数据
	Account     *AssetsAccount             `json:"account"`    //@required 创建交易单的账户
	Signatures  map[string][]*KeySignature `json:"signatures"` //拥有者accountID: []未花签名
	IsBuilt     bool                       `json:"isBuilt"`    //是否完成构建建议单
	IsCompleted bool                       `json:"isComplete"` //是否完成所有签名
	IsSubmit    bool                       `json:"isSubmit"`   //是否已广播
	CallParam   string                     `json:"callParam"`  //调用参数，用于调用智能合约，json结构

	/*
		ExtParam 根据不同区块链协议，对智能合约的调用，提供灵活可变的参数。
		例如：ETH 智能合约调用参数
		{
			"gasPrice": "0.000002",  						//自定义费率
			"gasLimit": "50000000",  						//自定义燃料上限
			"gasUsed": "32234",  							//实际使用燃料数
			"senderAddress": "0x1234567abcdeffdcba4321", 	//支付交易单的地址
			"contractAddress": "0xdeffdcba43211234567abc", 	//合约地址
			"amount": "0.001", 								//转入合约主币数量
			"callData": "deffdcba43211234567abc", 			//调用方法的ABI编码
			"nonce": 1,  									//地址账户交易序号
		}
	*/
}

type SmartContractTransaction struct {
	WxID        string                 `json:"wxid" storm:"id"` //@required 通过GenTransactionWxID计算
	TxID        string                 `json:"txid"`            //@required
	AccountID   string                 `json:"accountID"`       //@required 创建交易单的账户
	Actions     []*SmartContractAction `json:"actions"`         //@required 执行事件, 例如：合约的Transfer事件
	BlockHash   string                 `json:"blockHash"`       //@required
	BlockHeight uint64                 `json:"blockHeight"`     //@required
	SubmitTime  int64                  `json:"submitTime"`      //@required
	ConfirmTime int64                  `json:"confirmTime"`     //@required
	Status      string                 `json:"status"`          //链上状态，0：失败，1：成功
	Reason      string                 `json:"reason"`          //失败原因，失败状态码
	CallParam   string                 `json:"callParam"`       //调用参数，用于调用智能合约，json结构
}

type SmartContractAction struct {
	Method string `json:"method"` //执行方法
	Raw    string `json:"raw"`    //结果参数
}

//SmartContractDecoder 智能合约解析器
type SmartContractDecoder interface {
	ABIDAI

	//GetTokenBalanceByAddress 查询地址token余额列表
	GetTokenBalanceByAddress(contract SmartContract, address ...string) ([]*TokenBalance, error)

	//GetSmartContractInfo 获取智能合约信息
	//GetSmartContractInfo(contractAddress string) (*SmartContract, error)

}

type SmartContractDecoderBase struct {
}

func (decoder *SmartContractDecoderBase) GetTokenBalanceByAddress(contract SmartContract, address ...string) ([]*TokenBalance, error) {
	return nil, fmt.Errorf("GetTokenBalanceByAddress not implement")
}

// GetABIInfo get abi
func (decoder *SmartContractDecoderBase) GetABIInfo(address string) (*ABIInfo, error) {
	return nil, fmt.Errorf("GetABIInfo not implement")
}

// SetABIInfo set abi
func (decoder *SmartContractDecoderBase) SetABIInfo(address string, abi ABIInfo) error {
	return fmt.Errorf("GetABIInfo not implement")
}

// ABIDAI abi data access interface
type ABIDAI interface {
	//@require
	GetABIInfo(address string) (*ABIInfo, error)
	//@require
	SetABIInfo(address string, abi ABIInfo) error
}

// ABIInfo abi model
type ABIInfo struct {
	Address string      `json:"address"`
	ABI     interface{} `json:"abi"`
}
