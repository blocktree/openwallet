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
package ethereum

import (
	"encoding/json"
	"errors"
	"fmt"

	//"log"
	"math/big"
	"sort"
	"strconv"
	"sync"
	"time"

	//	"github.com/blocktree/OpenWallet/assets/ethereum"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCrypt"
	"github.com/bytom/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

type EthTxExtPara struct {
	Data     string `json:"data"`
	GasLimit string `json:"gasLimit"`
}

/*func (this *EthTxExtPara) GetGasLimit() (uint64, error) {
	gasLimit, err := strconv.ParseUint(removeOxFromHex(this.GasLimit), 16, 64)
	if err != nil {
		openwLogger.Log.Errorf("parse gas limit to uint64 failed, err=%v", err)
		return 0, err
	}
	return gasLimit, nil
}*/

const (
	ADRESS_STATIS_OVERDATED_TIME = 30
)

type AddressTxStatistic struct {
	Address          string
	TransactionCount *uint64
	LastModifiedTime *time.Time
	Valid            *int //如果valid指针指向的整形为0, 说明该地址已经被清理线程清理
	AddressLocker    *sync.Mutex
	//1. 地址级别, 不可并发广播交易, 造成nonce混乱
	//2. 地址级别, 不可广播, 读取nonce同时进行, 会造成nonce混乱
}

func (this *AddressTxStatistic) UpdateTime() {
	now := time.Now()
	this.LastModifiedTime = &now
}

type EthTransactionDecoder struct {
	openwallet.TransactionDecoderBase
	AddrTxStatisMap *sync.Map
	//	DecoderLocker *sync.Mutex    //保护一些全局不可并发的操作, 如对AddrTxStatisMap的初始化
	wm *WalletManager //钱包管理者
}

func (this *EthTransactionDecoder) GetTransactionCount2(address string) (*AddressTxStatistic, uint64, error) {
	now := time.Now()
	valid := 1
	t := AddressTxStatistic{
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
		TransactionCount: new(uint64),
	}

	v, loaded := this.AddrTxStatisMap.LoadOrStore(address, t)
	//LoadOrStore返回后, AddressLocker加锁前, map中的nonce可能已经被清理了, 需要检查valid是否为1
	txStatis := v.(AddressTxStatistic)
	txStatis.AddressLocker.Lock()
	txStatis.AddressLocker.Unlock()
	if loaded {
		if *txStatis.Valid == 0 {
			return nil, 0, errors.New("the node is busy, try it again later. ")
		}
		txStatis.UpdateTime()
		return &txStatis, *txStatis.TransactionCount, nil
	}
	nonce, err := this.wm.GetNonceForAddress2(appendOxToAddress(address))
	if err != nil {
		openwLogger.Log.Errorf("get nonce for address via rpc failed, err=%v", err)
		return nil, 0, err
	}
	*txStatis.TransactionCount = nonce
	return &txStatis, *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) GetTransactionCount(address string) (uint64, error) {
	if this.AddrTxStatisMap == nil {
		return 0, errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return 0, errors.New("no records found to the key passed through.")
	}

	txStatis := v.(AddressTxStatistic)
	return *txStatis.TransactionCount, nil
}

func (this *EthTransactionDecoder) SetTransactionCount(address string, transactionCount uint64) error {
	if this.AddrTxStatisMap == nil {
		return errors.New("map should be initialized before using.")
	}

	v, exist := this.AddrTxStatisMap.Load(address)
	if !exist {
		return errors.New("no records found to the key passed through.")
	}

	now := time.Now()
	valid := 1
	txStatis := AddressTxStatistic{
		TransactionCount: &transactionCount,
		LastModifiedTime: &now,
		AddressLocker:    new(sync.Mutex),
		Valid:            &valid,
		Address:          address,
	}

	if exist {
		txStatis.AddressLocker = v.(AddressTxStatistic).AddressLocker
	} else {
		txStatis.AddressLocker = &sync.Mutex{}
	}

	this.AddrTxStatisMap.Store(address, txStatis)
	return nil
}

func (this *EthTransactionDecoder) RemoveOutdatedAddrStatic() {
	addrStatisList := make([]AddressTxStatistic, 0)
	this.AddrTxStatisMap.Range(func(k, v interface{}) bool {
		addrStatis := v.(AddressTxStatistic)
		if addrStatis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			addrStatisList = append(addrStatisList, addrStatis)
		}
		return true
	})

	clear := func(statis *AddressTxStatistic) {
		statis.AddressLocker.Lock()
		defer statis.AddressLocker.Unlock()
		if statis.LastModifiedTime.Before(time.Now().Add(-1 * (ADRESS_STATIS_OVERDATED_TIME * time.Minute))) {
			*statis.Valid = 0
			this.AddrTxStatisMap.Delete(statis)
		}
	}

	for i, _ := range addrStatisList {
		clear(&addrStatisList[i])
	}
}

func (this *EthTransactionDecoder) RunClearAddrStatic() {
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			this.RemoveOutdatedAddrStatic()
		}
	}()
}

func VerifyRawTransaction(rawTx *openwallet.RawTransaction) error {
	if len(rawTx.To) != 1 {
		openwLogger.Log.Errorf("noly one to address can be set.")
		return errors.New("noly one to address can be set.")
	}

	return nil
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *EthTransactionDecoder {
	decoder := EthTransactionDecoder{}
	//	decoder.DecoderLocker = new(sync.Mutex)
	decoder.wm = wm
	decoder.AddrTxStatisMap = new(sync.Map)
	decoder.RunClearAddrStatic()
	return &decoder
}

func (this *EthTransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		log.Error("get address list failed, err=", err)
		return err
	}
	if len(addresses) == 0 {
		openwLogger.Log.Errorf("no addresses found in wallet[%v]", rawTx.Account.AccountID)
		return errors.New("no addresses found in wallet.")
	}

	addrsBalanceList := make([]AddrBalance, 0, len(addresses))
	for i, addr := range addresses {
		balance, err := ConvertEthStringToWei(addr.Balance) //ConvertToBigInt(addr.Balance, 16)
		if err != nil {
			openwLogger.Log.Errorf("convert address [%v] balance [%v] to big.int failed, err = %v ", addr.Address, addr.Balance, err)
			return err
		}
		addrsBalanceList = append(addrsBalanceList, AddrBalance{
			Address: addr.Address,
			Balance: balance,
			Index:   i,
		})
	}

	sort.Slice(addrsBalanceList, func(i int, j int) bool {
		if addrsBalanceList[i].Balance.Cmp(addrsBalanceList[j].Balance) < 0 {
			return true
		}
		return false
	})

	signatureMap := make(map[string][]*openwallet.KeySignature)
	keySignList := make([]*openwallet.KeySignature, 0, 1)
	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, err := ConvertEthStringToWei(amountStr) //ConvertToBigInt(amountStr, 16)
	if err != nil {
		openwLogger.Log.Errorf("convert tx amount to big.int failed, err=%v", err)
		return err
	}

	var fee *txFeeInfo
	totalFee := big.NewInt(0)
	//	var data string
	for i, _ := range addrsBalanceList {
		totalAmount := new(big.Int)
		if addrsBalanceList[i].Balance.Cmp(amount) > 0 {
			fee, err = this.wm.GetTransactionFeeEstimated(addrsBalanceList[i].Address, to, amount, "")
			if err != nil {
				openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrsBalanceList[i].Address, to, err)
				return err
			}

			if rawTx.FeeRate != "" {
				fee.GasPrice, err = ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
				if err != nil {
					openwLogger.Log.Errorf("fee rate passed through error, err=%v", err)
					return err
				}
				fee.CalcFee()
			}

			totalAmount.Add(totalAmount, fee.Fee)
			if addrsBalanceList[i].Balance.Cmp(totalAmount) > 0 {
				//fromAddr = addrsBalanceList[i].address
				fromAddr := &openwallet.Address{
					AccountID:   addresses[addrsBalanceList[i].Index].AccountID,
					Address:     addresses[addrsBalanceList[i].Index].Address,
					PublicKey:   addresses[addrsBalanceList[i].Index].PublicKey,
					Alias:       addresses[addrsBalanceList[i].Index].Alias,
					Tag:         addresses[addrsBalanceList[i].Index].Tag,
					Index:       addresses[addrsBalanceList[i].Index].Index,
					HDPath:      addresses[addrsBalanceList[i].Index].HDPath,
					WatchOnly:   addresses[addrsBalanceList[i].Index].WatchOnly,
					Symbol:      addresses[addrsBalanceList[i].Index].Symbol,
					Balance:     addresses[addrsBalanceList[i].Index].Balance,
					IsMemo:      addresses[addrsBalanceList[i].Index].IsMemo,
					Memo:        addresses[addrsBalanceList[i].Index].Memo,
					CreatedTime: addresses[addrsBalanceList[i].Index].CreatedTime,
				}

				keySignList = append(keySignList, &openwallet.KeySignature{
					Address: fromAddr,
				})
				totalFee.Add(totalFee, fee.Fee)
				break
			}
		}
	}

	totalFeeDecimal, err := ConverWeiStringToEthDecimal(totalFee.String())
	if err != nil {
		log.Errorf("convert total fee from wei string to eth decimal failed, err=%v", err)
		return err
	}
	rawTx.Fees = totalFeeDecimal.String()

	if len(keySignList) != 1 {
		return errors.New("no enough balance address found in wallet. ")
	}

	_, nonce, err := this.GetTransactionCount2(keySignList[0].Address.Address)
	if err != nil {
		openwLogger.Log.Errorf("GetTransactionCount2 failed, err=%v", err)
		return err
	}
	log.Debug("chainID:", this.wm.GetConfig().ChainID)
	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))
	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(to),
		amount, fee.GasLimit.Uint64(), fee.GasPrice, []byte(""))

	txstr, _ := json.MarshalIndent(tx, "", " ")
	log.Debug("**txStr:", string(txstr))
	msg := signer.Hash(tx)

	gasLimitStr, err := ConverWeiStringToEthDecimal(fee.GasLimit.String())
	if err != nil {
		log.Error("ConverWeiStringToEthDecimal failed, err=", err)
		return err
	}

	extpara := EthTxExtPara{
		GasLimit: gasLimitStr.String(), //"0x" + fee.GasLimit.Text(16),
	}
	extparastr, _ := json.Marshal(extpara)
	rawTx.ExtParam = string(extparastr)
	keySignList[0].Nonce = "0x" + strconv.FormatUint(nonce, 16)
	keySignList[0].Message = common.ToHex(msg[:])
	signatureMap[rawTx.Account.AccountID] = keySignList
	rawTx.Signatures = signatureMap
	gasprice, err := ConverWeiStringToEthDecimal(fee.GasPrice.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}
	rawTx.FeeRate = gasprice.String()
	rawTx.IsBuilt = true

	return nil
}

func (this *EthTransactionDecoder) CreateErc20TokenRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	//获取wallet
	addresses, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID) //wrapper.GetWallet().GetAddressesByAccount(rawTx.Account.AccountID)
	if err != nil {
		log.Error("get address list failed, err=", err)
		return err
	}
	if len(addresses) == 0 {
		openwLogger.Log.Errorf("no addresses found in wallet[%v]", rawTx.Account.AccountID)
		return errors.New("no addresses found in wallet.")
	}

	addrsBalanceList := make([]*AddrBalance, 0, len(addresses))
	addrsBalanceIfList := make([]AddrBalanceInf, 0, len(addresses))
	for i, addr := range addresses {
		balance, err := ConvertEthStringToWei(addr.Balance) //ConvertToBigInt(addr.Balance, 16)
		if err != nil {
			openwLogger.Log.Errorf("convert address [%v] balance [%v] to big.int failed, err = %v ", addr.Address, addr.Balance, err)
			return err
		}

		addBalance := &AddrBalance{
			Address: addr.Address,
			Balance: balance,
			Index:   i,
		}
		addrsBalanceList = append(addrsBalanceList, addBalance)
		addrsBalanceIfList = append(addrsBalanceIfList, addBalance)
	}

	err = this.wm.GetTokenBalanceByAddress(rawTx.Coin.Contract.Address, addrsBalanceIfList...)
	if err != nil {
		log.Errorf("get token balance failed, err=%v", err)
		return err
	}

	sort.Slice(addrsBalanceList, func(i int, j int) bool {
		if addrsBalanceList[i].TokenBalance.Cmp(addrsBalanceList[j].TokenBalance) < 0 {
			return true
		}
		return false
	})

	signatureMap := make(map[string][]*openwallet.KeySignature)
	keySignList := make([]*openwallet.KeySignature, 0, 1)
	var amountStr, to string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}

	amount, err := ConvertFloatStringToBigInt(amountStr, int(rawTx.Coin.Contract.Decimals)) //ConvertToBigInt(amountStr, 10)
	if err != nil {
		openwLogger.Log.Errorf("convert tx amount to big.int failed, err=%v", err)
		return err
	}

	var fee *txFeeInfo
	var data string
	totalFee := big.NewInt(0)
	for i, _ := range addrsBalanceList {
		//		totalAmount := new(big.Int)
		if addrsBalanceList[i].TokenBalance.Cmp(amount) > 0 {
			data, err = makeERC20TokenTransData(rawTx.Coin.Contract.Address, to, amount)
			if err != nil {
				openwLogger.Log.Errorf("make token transaction data failed, err=%v", err)
				return err
			}

			fee, err = this.wm.GetTransactionFeeEstimated(addrsBalanceList[i].Address, rawTx.Coin.Contract.Address, nil, data)
			if err != nil {
				openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrsBalanceList[i].Address, to, err)
				return err
			}

			if rawTx.FeeRate != "" {
				fee.GasPrice, err = ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
				if err != nil {
					openwLogger.Log.Errorf("fee rate passed through error, err=%v", err)
					return err
				}
				fee.CalcFee()
			}

			if addrsBalanceList[i].Balance.Cmp(fee.Fee) > 0 {
				fromAddr := &openwallet.Address{
					AccountID:   addresses[addrsBalanceList[i].Index].AccountID,
					Address:     addresses[addrsBalanceList[i].Index].Address,
					PublicKey:   addresses[addrsBalanceList[i].Index].PublicKey,
					Alias:       addresses[addrsBalanceList[i].Index].Alias,
					Tag:         addresses[addrsBalanceList[i].Index].Tag,
					Index:       addresses[addrsBalanceList[i].Index].Index,
					HDPath:      addresses[addrsBalanceList[i].Index].HDPath,
					WatchOnly:   addresses[addrsBalanceList[i].Index].WatchOnly,
					Symbol:      addresses[addrsBalanceList[i].Index].Symbol,
					Balance:     addresses[addrsBalanceList[i].Index].Balance,
					IsMemo:      addresses[addrsBalanceList[i].Index].IsMemo,
					Memo:        addresses[addrsBalanceList[i].Index].Memo,
					CreatedTime: addresses[addrsBalanceList[i].Index].CreatedTime,
				}

				keySignList = append(keySignList, &openwallet.KeySignature{
					Address: fromAddr,
				})

				totalFee.Add(totalFee, fee.Fee)
				break
			}
		}
	}

	if len(keySignList) != 1 {
		return errors.New("no enough balance address found in wallet. ")
	}

	totalFeeDecimal, err := ConverWeiStringToEthDecimal(totalFee.String())
	if err != nil {
		log.Errorf("convert fee from wei string to eth decimal failed, err=%v", err)
		return err
	}

	rawTx.Fees = totalFeeDecimal.String()

	_, nonce, err := this.GetTransactionCount2(keySignList[0].Address.Address)
	if err != nil {
		openwLogger.Log.Errorf("GetTransactionCount2 failed, err=%v", err)
		return err
	}

	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))
	tx := types.NewTransaction(nonce, ethcommon.HexToAddress(rawTx.Coin.Contract.Address),
		big.NewInt(0), fee.GasLimit.Uint64(), fee.GasPrice, common.FromHex(data))

	txstr, _ := json.MarshalIndent(tx, "", " ")
	log.Debug("**txStr:", string(txstr))
	msg := signer.Hash(tx)

	gasLimitStr, err := ConverWeiStringToEthDecimal(fee.GasLimit.String())
	if err != nil {
		log.Error("ConverWeiStringToEthDecimal failed, err=", err)
		return err
	}

	extpara := EthTxExtPara{
		Data:     data,
		GasLimit: gasLimitStr.String(), //"0x" + fee.GasLimit.Text(16),
	}
	extparastr, _ := json.Marshal(extpara)
	rawTx.ExtParam = string(extparastr)

	keySignList[0].Nonce = "0x" + strconv.FormatUint(nonce, 16)
	keySignList[0].Message = common.ToHex(msg[:])

	signatureMap[rawTx.Account.AccountID] = keySignList
	rawTx.Signatures = signatureMap
	gasprice, err := ConverWeiStringToEthDecimal(fee.GasPrice.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}
	rawTx.FeeRate = gasprice.String()
	rawTx.IsBuilt = true
	return nil
}

//CreateRawTransaction 创建交易单
func (this *EthTransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return this.CreateSimpleRawTransaction(wrapper, rawTx)
	}
	return this.CreateErc20TokenRawTransaction(wrapper, rawTx)
}

//SignRawTransaction 签名交易单
func (this *EthTransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	if len(rawTx.Signatures) != 1 {
		log.Error("raw tx signature error")
		return errors.New("raw tx signature error")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		log.Error("get HDKey from wallet wrapper failed, err=%v", err)
		return err
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(rawTx.Signatures[rawTx.Account.AccountID]) != 1 {
		log.Error("signature failed in account[%v].", rawTx.Account.AccountID)
		return errors.New("signature failed in account.")
	}

	signnode := rawTx.Signatures[rawTx.Account.AccountID][0]
	fromAddr := signnode.Address

	childKey, _ := key.DerivedKeyWithPath(fromAddr.HDPath, owcrypt.ECC_CURVE_SECP256K1)
	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		log.Error("get private key bytes, err=", err)
		return err
	}
	//prikeyStr := common.ToHex(keyBytes)
	//log.Debugf("pri:%v", common.ToHex(keyBytes))

	message := common.FromHex(signnode.Message)
	//seckey := math.PaddedBigBytes(key.PrivateKey.D, key.PrivateKey.Params().BitSize/8)

	/*sig, err := secp256k1.Sign(message, keyBytes)
	if err != nil {
		log.Error("secp256k1.Sign failed, err=", err)
		return err
	}*/
	sig, ret := owcrypt.ETHsignature(keyBytes, message)
	if ret != owcrypt.SUCCESS {
		errdesc := fmt.Sprintln("signature error, ret:", "0x"+strconv.FormatUint(uint64(ret), 16))
		log.Error(errdesc)
		return errors.New(errdesc)
	}

	signnode.Signature = common.ToHex(sig)

	log.Debug("** pri:", common.ToHex(keyBytes))
	log.Debug("** message:", signnode.Message)
	log.Debug("** Signature:", signnode.Signature)

	return nil
}

func (this *EthTransactionDecoder) SubmitSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature

	var extPara EthTxExtPara
	log.Debug("rawTx.ExtParam:", rawTx.ExtParam)
	err = json.Unmarshal([]byte(rawTx.ExtParam), &extPara)
	if err != nil {
		openwLogger.Log.Errorf("decode json from extpara failed, err=%v", err)
		return err
	}

	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))

	var to, amountStr string
	for k, v := range rawTx.To {
		to = k
		amountStr = v
		break
	}
	amount, err := ConvertEthStringToWei(amountStr) //ConvertToBigInt(amountStr, 10)
	if err != nil {
		openwLogger.Log.Errorf("amount convert to big int failed, err=%v", err)
		return err
	}

	txStatis, _, err := this.GetTransactionCount2(from)
	if err != nil {
		openwLogger.Log.Errorf("get transaction count2 failed, err=%v", err)
		return errors.New("get transaction count2 faile")
	}

	log.Debug("extPara.GasLimit:", extPara.GasLimit)
	gaslimit, err := ConvertEthStringToWei(extPara.GasLimit) //extPara.GetGasLimit()
	if err != nil {
		openwLogger.Log.Errorf("get gas limit failed, err=%v", err)
		return errors.New("get gas limit failed")
	}

	gasPrice, err := ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
	if err != nil {
		openwLogger.Log.Errorf("get gas price failed, err=%v", err)
		return errors.New("get gas price failed")
	}

	err = func() error {
		txStatis.AddressLocker.Lock()
		defer txStatis.AddressLocker.Unlock()
		nonceSigned, err := strconv.ParseUint(removeOxFromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Nonce),
			16, 64)
		if err != nil {
			openwLogger.Log.Errorf("parse nonce from rawTx failed, err=%v", err)
			return errors.New("parse nonce from rawTx failed. ")
		}
		if nonceSigned != *txStatis.TransactionCount {
			openwLogger.Log.Errorf("nonce out of dated, please try to start ur tx once again. ")
			return errors.New("nonce out of dated, please try to start ur tx once again. ")
		}

		tx := types.NewTransaction(nonceSigned, ethcommon.HexToAddress(to),
			amount, gaslimit.Uint64(), gasPrice, nil)
		tx, err = tx.WithSignature(signer, common.FromHex(sig))
		if err != nil {
			openwLogger.Log.Errorf("tx with signature failed, err=%v ", err)
			return errors.New("tx with signature failed. ")
		}

		txstr, _ := json.MarshalIndent(tx, "", " ")
		log.Debug("**after signed txStr:", string(txstr))

		rawTxPara, err := rlp.EncodeToBytes(tx)
		if err != nil {
			openwLogger.Log.Errorf("encode tx to rlp failed, err=%v ", err)
			return errors.New("encode tx to rlp failed. ")
		}

		txid, err := this.wm.WalletClient.ethSendRawTransaction(common.ToHex(rawTxPara))
		if err != nil {
			openwLogger.Log.Errorf("sent raw tx faield, err=%v", err)
			return errors.New("sent raw tx faield. ")
		}

		rawTx.TxID = txid
		rawTx.IsSubmit = true
		txStatis.UpdateTime()
		(*txStatis.TransactionCount)++

		log.Debug("transaction[", txid, "] has been sent out.")
		return nil
	}()

	if err != nil {
		log.Errorf("send raw transaction failed, err= %v", err)
		return err
	}

	return nil
}

func (this *EthTransactionDecoder) SubmitErc20TokenRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	from := rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature

	var extPara EthTxExtPara
	log.Debug("rawTx.ExtParam:", rawTx.ExtParam)
	err = json.Unmarshal([]byte(rawTx.ExtParam), &extPara)
	if err != nil {
		openwLogger.Log.Errorf("decode json from extpara failed, err=%v", err)
		return err
	}

	data := extPara.Data
	log.Debug("extPara.GasLimit:", extPara.GasLimit)
	gaslimit, err := ConvertEthStringToWei(extPara.GasLimit) //extPara.GetGasLimit()
	if err != nil {
		openwLogger.Log.Errorf("get gas limit failed, err=%v", err)
		return errors.New("get gas limit failed")
	}

	signer := types.NewEIP155Signer(big.NewInt(int64(this.wm.GetConfig().ChainID)))

	txStatis, _, err := this.GetTransactionCount2(from)
	if err != nil {
		openwLogger.Log.Errorf("get transaction count2 failed, err=%v", err)
		return errors.New("get transaction count2 faile")
	}

	gasPrice, err := ConvertEthStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
	if err != nil {
		openwLogger.Log.Errorf("get gas price failed, err=%v", err)
		return errors.New("get gas price failed")
	}

	err = func() error {
		txStatis.AddressLocker.Lock()
		defer txStatis.AddressLocker.Unlock()

		nonceSigned, err := strconv.ParseUint(removeOxFromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Nonce),
			16, 64)
		if err != nil {
			openwLogger.Log.Errorf("parse nonce from rawTx failed, err=%v", err)
			return errors.New("parse nonce from rawTx failed. ")
		}

		if nonceSigned != *txStatis.TransactionCount {
			openwLogger.Log.Errorf("nonce out of dated, please try to start ur tx once again. ")
			return errors.New("nonce out of dated, please try to start ur tx once again. ")
		}

		tx := types.NewTransaction(nonceSigned, ethcommon.HexToAddress(rawTx.Coin.Contract.Address),
			big.NewInt(0), gaslimit.Uint64(), gasPrice, common.FromHex(data))
		tx, err = tx.WithSignature(signer, common.FromHex(sig))
		if err != nil {
			openwLogger.Log.Errorf("tx with signature failed, err=%v ", err)
			return errors.New("tx with signature failed. ")
		}

		txstr, _ := json.MarshalIndent(tx, "", " ")
		log.Debug("**after signed txStr:", string(txstr))

		rawTxPara, err := rlp.EncodeToBytes(tx)
		if err != nil {
			openwLogger.Log.Errorf("encode tx to rlp failed, err=%v ", err)
			return errors.New("encode tx to rlp failed. ")
		}

		txid, err := this.wm.WalletClient.ethSendRawTransaction(common.ToHex(rawTxPara))
		if err != nil {
			openwLogger.Log.Errorf("sent raw tx faield, err=%v", err)
			return errors.New("sent raw tx faield. ")
		}

		rawTx.TxID = txid
		rawTx.IsSubmit = true
		txStatis.UpdateTime()
		(*txStatis.TransactionCount)++

		log.Debug("transaction[", txid, "] has been sent out.")
		return nil
	}()

	if err != nil {
		log.Errorf("send raw transaction failed, err= %v", err)
		return err
	}
	return nil
}

//SendRawTransaction 广播交易单
func (this *EthTransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return this.SubmitSimpleRawTransaction(wrapper, rawTx)
	}
	return this.SubmitErc20TokenRawTransaction(wrapper, rawTx)
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (this *EthTransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := VerifyRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}

	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	if _, exist := rawTx.Signatures[rawTx.Account.AccountID]; !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}
	sig := rawTx.Signatures[rawTx.Account.AccountID][0].Signature
	msg := rawTx.Signatures[rawTx.Account.AccountID][0].Message
	pubkey := rawTx.Signatures[rawTx.Account.AccountID][0].Address.PublicKey
	//curveType := rawTx.Signatures[rawTx.Account.AccountID][0].EccType

	log.Debug("-- pubkey:", pubkey)
	log.Debug("-- message:", msg)
	log.Debug("-- Signature:", sig)
	signature := common.FromHex(sig)
	publickKey := owcrypt.PointDecompress(common.FromHex(pubkey), owcrypt.ECC_CURVE_SECP256K1)
	publickKey = publickKey[1:len(publickKey)]
	ret := owcrypt.Verify(publickKey, nil, 0, common.FromHex(msg), 32, signature[0:len(signature)-1],
		owcrypt.ECC_CURVE_SECP256K1|owcrypt.HASH_OUTSIDE_FLAG)
	if ret != owcrypt.SUCCESS {
		errinfo := fmt.Sprintf("verify error, ret:%v\n", "0x"+strconv.FormatUint(uint64(ret), 16))
		fmt.Println(errinfo)
		return errors.New(errinfo)
	}

	return nil
}
