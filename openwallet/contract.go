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
	abi        string
}

// GetABI
func (contract *SmartContract) GetABI() string {
	return contract.abi
}

// SetABI
func (contract *SmartContract) SetABI(abiJson string) {
	contract.abi = abiJson
}

//GenContractID 合约ID
func GenContractID(symbol, address string) string {
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return base64.StdEncoding.EncodeToString(crypto.SHA256([]byte(fmt.Sprintf("%v_%v", symbol, address))))
}

const (
	// 0：hex字符串，1：json字符串，2：base64字符串
	TxRawTypeHex    = 0
	TxRawTypeJSON   = 1
	TxRawTypeBase64 = 2
)

// SmartContractRawTransaction 智能合约原始交易单
type SmartContractRawTransaction struct {
	Coin        Coin                       `json:"coin"`       //@required 区块链类型标识
	TxID        string                     `json:"txID"`       //交易单ID，广播后会生成
	Sid         string                     `json:"sid"`        //@required 业务订单号，保证业务不重复交易而用
	Account     *AssetsAccount             `json:"account"`    //@required 创建交易单的账户
	Signatures  map[string][]*KeySignature `json:"signatures"` //拥有者accountID: []未花签名
	IsBuilt     bool                       `json:"isBuilt"`    //是否完成构建建议单
	IsCompleted bool                       `json:"isComplete"` //是否完成所有签名
	IsSubmit    bool                       `json:"isSubmit"`   //是否已广播
	Raw         string                     `json:"Raw"`        //交易单调用参数，根据RawType填充数据
	RawType     uint64                     `json:"rawType"`    // 0：hex字符串，1：json字符串，2：base64字符串
	ABIParam    []string                   `json:"abiParam"`   //abi调用参数，[method, arg1, arg2, args...]
	Value       string                     `json:"value"`      //主币数量
	FeeRate     string                     `json:"feeRate"`    //自定义费率
	Fees        string                     `json:"fees"`       //手续费
	TxFrom      string                     `json:"txFrom"`     //调用地址
	TxTo        string                     `json:"txTo"`       //调用地址，与合约地址一致
}

type SmartContractReceipt struct {
	Coin        Coin                  `json:"coin"`            //@required 区块链类型标识
	WxID        string                `json:"wxid" storm:"id"` //@required 通过GenTransactionWxID计算
	TxID        string                `json:"txid"`            //@required
	From        string                `json:"from"`            //@required 调用者
	To          string                `json:"to"`              //@required 调用地址，与合约地址一致
	Value       string                `json:"value"`           //主币数量
	Fees        string                `json:"fees"`            //手续费
	RawReceipt  string                `json:"rawReceipt"`      //@required 原始交易回执，一般为json
	Events      []*SmartContractEvent `json:"actions"`         //@required 执行事件, 例如：event Transfer
	BlockHash   string                `json:"blockHash"`       //@required
	BlockHeight uint64                `json:"blockHeight"`     //@required
	ConfirmTime int64                 `json:"confirmTime"`     //@required
	Status      string                `json:"status"`          //@required 链上状态，0：失败，1：成功
	Reason      string                `json:"reason"`          //失败原因，失败状态码
	ExtParam    string                `json:"extParam"`        //扩展参数，用于调用智能合约，json结构

	/*
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

func (tx *SmartContractReceipt) GenWxID() {
	tx.WxID = GenTransactionWxID2(tx.TxID, tx.Coin.Symbol, tx.Coin.ContractID)
}

// SmartContractEvent 事件记录
type SmartContractEvent struct {
	Contract *SmartContract `json:"contract"` //合约
	Event    string         `json:"event"`    //记录事件
	Value    string         `json:"Value"`    //结果参数，json字符串
}

const (
	SmartContractCallResultStatusFail    = 0
	SmartContractCallResultStatusSuccess = 1
)

// SmartContractCallResult 调用结果，不产生交易
type SmartContractCallResult struct {
	Method    string `json:"method"`    //调用方法
	Value     string `json:"Value"`     //json结果
	RawHex    string `json:"rawHex"`    //16进制字符串结果
	Status    uint64 `json:"status"`    //0：成功，1：失败
	Exception string `json:"exception"` //异常错误
}

//SmartContractDecoder 智能合约解析器
type SmartContractDecoder interface {
	ABIDAI

	//GetTokenBalanceByAddress 查询地址token余额列表
	GetTokenBalanceByAddress(contract SmartContract, address ...string) ([]*TokenBalance, error)
	//调用合约ABI方法
	CallSmartContractABI(wrapper WalletDAI, rawTx *SmartContractRawTransaction) (*SmartContractCallResult, *Error)
	//创建原始交易单
	CreateSmartContractRawTransaction(wrapper WalletDAI, rawTx *SmartContractRawTransaction) *Error
	//SubmitRawTransaction 广播交易单
	SubmitSmartContractRawTransaction(wrapper WalletDAI, rawTx *SmartContractRawTransaction) (*SmartContractReceipt, *Error)
}

type SmartContractDecoderBase struct {
}

func (decoder *SmartContractDecoderBase) GetTokenBalanceByAddress(contract SmartContract, address ...string) ([]*TokenBalance, error) {
	return nil, fmt.Errorf("GetTokenBalanceByAddress not implement")
}

//调用合约ABI方法
func (decoder *SmartContractDecoderBase) CallSmartContractABI(wrapper WalletDAI, rawTx *SmartContractRawTransaction) (*SmartContractCallResult, *Error) {
	return nil, Errorf(ErrSystemException, "CallSmartContractABI not implement")
}

//创建原始交易单
func (decoder *SmartContractDecoderBase) CreateSmartContractRawTransaction(wrapper WalletDAI, rawTx *SmartContractRawTransaction) *Error {
	return Errorf(ErrSystemException, "CreateSmartContractRawTransaction not implement")
}

//SubmitRawTransaction 广播交易单
func (decoder *SmartContractDecoderBase) SubmitSmartContractRawTransaction(wrapper WalletDAI, rawTx *SmartContractRawTransaction) (*SmartContractReceipt, *Error) {
	return nil, Errorf(ErrSystemException, "SubmitSmartContractRawTransaction not implement")
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
