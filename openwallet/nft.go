/*
 * Copyright 2022 The OpenWallet Authors
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

package openwallet

type NFT struct {
	Symbol   string `json:"symbol"` //@required 主币的symbol
	Address  string `json:"address"`
	Token    string `json:"token"` //@required NFT的symbol
	Protocol string `json:"protocol"`
	Name     string `json:"name"`
	TokenID  string `json:"tokenID"`
}

type NFTBalance struct {
	NFT     *NFT
	Balance string
}

type NFTOwner struct {
	NFT   *NFT
	Owner string
}

type NFTMetaData struct {
	NFT *NFT
	URI string
}

// NFTEventType
const (
	NFTEventTypeTransferred = 0 //转账
	NFTEventTypeMinted      = 1 //铸造
	NFTEventTypeBurned      = 2 //销毁
)

type NFTTransfer struct {
	Tokens    []NFT    `json:"tokens"`    //@required nft
	Operator  string   `json:"operator"`  //required 被授权转账的操作者
	From      string   `json:"from"`      //@required 发送者
	To        string   `json:"to"`        //@required 接受者
	Amounts   []string `json:"amounts"`   //@required erc1155 token有数量
	EventType uint64   `json:"eventType"` //@required
}

type NFTApproval struct {
	Token    *NFT   `json:"token"`    //@required nft
	Owner    string `json:"owner"`    //@required nft拥有者
	Operator string `json:"operator"` //required 被授权的操作者
	IsAll    bool   `json:"isAll"`    //@required 是否全部NFT，true：token = null，false，token = *NFT
}

//NFTContractDecoder NFT智能合约解析器
type NFTContractDecoder interface {
	//GetNFTBalanceByAddress 查询地址NFT余额列表
	GetNFTBalanceByAddress(nft *NFT, owner string) (*NFTBalance, *Error)
	// GetNFTBalanceByAddressBatch 查询地址NFT余额列表
	GetNFTBalanceByAddressBatch(nft []*NFT, owner []string) ([]*NFTBalance, *Error)
	//GetNFTOwnerByTokenID 查询地址token的拥有者
	GetNFTOwnerByTokenID(nft *NFT) (*NFTOwner, *Error)
	//GetMetaDataOfNFT 查询NFT的MetaData
	GetMetaDataOfNFT(nft *NFT) (*NFTMetaData, *Error)
	//GetNFTTransfer 从event解析NFT转账信息
	GetNFTTransfer(event *SmartContractEvent) (*NFTTransfer, *Error)
}

type NFTContractDecoderBase struct {
}

// GetNFTBalanceByAddress 查询地址NFT余额列表
// NFT.TokenID为空则查询合约下拥有者所NFT数量。
func (decoder *NFTContractDecoderBase) GetNFTBalanceByAddress(nft *NFT, owner string) (*NFTBalance, *Error) {
	return nil, Errorf(ErrSystemException, "GetNFTBalanceByAddress not implement")
}

// GetNFTBalanceByAddressBatch 查询地址NFT余额列表
func (decoder *NFTContractDecoderBase) GetNFTBalanceByAddressBatch(nft []*NFT, owner []string) ([]*NFTBalance, *Error) {
	return nil, Errorf(ErrSystemException, "GetNFTBalanceByAddressBatch not implement")
}

//GetNFTOwnerByTokenID 查询地址token的拥有者
func (decoder *NFTContractDecoderBase) GetNFTOwnerByTokenID(nft *NFT) (*NFTOwner, *Error) {
	return nil, Errorf(ErrSystemException, "GetNFTOwnerByTokenID not implement")
}

//GetMetaDataOfNFT 查询NFT的MetaData
func (decoder *NFTContractDecoderBase) GetMetaDataOfNFT(nft *NFT) (*NFTMetaData, *Error) {
	return nil, Errorf(ErrSystemException, "GetMetaDataOfNFT not implement")
}

//GetNFTTransfer 从event解析NFT转账信息
func (decoder *NFTContractDecoderBase) GetNFTTransfer(event *SmartContractEvent) (*NFTTransfer, *Error) {
	return nil, Errorf(ErrSystemException, "GetNFTTransfer not implement")
}
