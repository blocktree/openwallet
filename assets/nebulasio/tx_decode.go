/*
 * Copyright 2018 The OpenWallet Authors
 * decoder file is part of the OpenWallet library.
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
package nebulasio

import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"sort"
	"strconv"

	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-owcrypt"
	"github.com/bytom/common"
	"github.com/shopspring/decimal"
)

type TransactionDecoder struct {
	openwallet.TransactionDecoderBase
	wm *WalletManager //钱包管理者
}

//NewTransactionDecoder 交易单解析器
func NewTransactionDecoder(wm *WalletManager) *TransactionDecoder {
	decoder := TransactionDecoder{}
	decoder.wm = wm
	return &decoder
}

func CheckRawTransaction(rawTx *openwallet.RawTransaction) error {
	//账户模型原始账单只有一个To
	if len(rawTx.To) != 1 {
		openwLogger.Log.Errorf("noly one To address can be set.")
		return errors.New("noly one to address can be set.")
	}
	return nil
}

type AddrBalance struct {
	Address      string
	Balance      *big.Int
	TokenBalance *big.Int
	Index        int
}

type txFeeInfo struct {
	GasUse   *big.Int
	GasPrice *big.Int
	Fee      *big.Int
}

func (decoder *TransactionDecoder) CreateSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
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
		balance, err := ConvertNasStringToWei(addr.Balance) //ConvertToBigInt(addr.Balance, 16)
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

	amount, err := ConvertNasStringToWei(amountStr)
	if err != nil {
		openwLogger.Log.Errorf("convert tx amount to big.int failed, err=%v", err)
		return err
	}

	var estimatefee *txFeeInfo
	//	var data string
	for i, _ := range addrsBalanceList {
		totalAmount := new(big.Int)
		if addrsBalanceList[i].Balance.Cmp(amount) > 0 {
			estimatefee, err = decoder.wm.Getestimatefee(addrsBalanceList[i].Address, to, amount)
			if err != nil {
				openwLogger.Log.Errorf("GetTransactionFeeEstimated from[%v] -> to[%v] failed, err=%v", addrsBalanceList[i].Address, to, err)
				return err
			}

			if rawTx.FeeRate != "" {
				estimatefee.GasPrice, err = ConvertNasStringToWei(rawTx.FeeRate) //ConvertToBigInt(rawTx.FeeRate, 16)
				if err != nil {
					openwLogger.Log.Errorf("fee rate passed through error, err=%v", err)
					return err
				}
				estimatefee.CalcFee()
			}

			totalAmount.Add(totalAmount, estimatefee.Fee)

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
				break
			}
		}
	}

	if len(keySignList) != 1 {
		return errors.New("no enough balance address found in wallet. ")
	}

	//最终费率
	gasprice, err := ConverWeiStringToNasDecimal(estimatefee.GasPrice.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}
	//最终手续费
	fee, err := ConverWeiStringToNasDecimal(estimatefee.Fee.String())
	if err != nil {
		log.Error("convert wei string to gas price failed, err=", err)
		return err
	}

	var nonce uint64
	//获取db记录的nonce并确认nonce值
	nonce_db, err := wrapper.GetAddressExtParam(keySignList[0].Address.Address, decoder.wm.FullName())
	if err != nil {
		return err
	}
	//判断nonce_db是否为空,为空则说明当前nonce是0
	if nonce_db == nil {
		nonce = 1
	} else {
		nonce = decoder.wm.ConfirmTxdecodeNonce(keySignList[0].Address.Address, nonce_db.(string))
	}
	//构建交易单
	var TX *SubmitTransaction
	TX, err = decoder.wm.CreateRawTransaction(keySignList[0].Address.Address, to, Gaslimit, gasprice.Mul(coinDecimal).String(), amount.String(), nonce)
	if err != nil {
		return err
	}

	rawHex, err := EncodeToTransactionRawHex(TX)
	if err != nil {
		return err
	}

	keySignList[0].Nonce = strconv.FormatUint(nonce, 10)
	keySignList[0].Message = hex.EncodeToString(TX.Hash[:])
	signatureMap[rawTx.Account.AccountID] = keySignList

	rawTx.RawHex = rawHex
	rawTx.Signatures = signatureMap
	rawTx.FeeRate = gasprice.String()
	rawTx.Fees = fee.StringFixed(decoder.wm.Decimal())
	rawTx.IsBuilt = true

	return nil
}

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.CreateSimpleRawTransaction(wrapper, rawTx)
	}
	//wjq return decoder.CreateNRC20TokenRawTransaction(wrapper, rawTx)
	return nil
}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
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

	keySignature := rawTx.Signatures[rawTx.Account.AccountID][0]
	fromAddr := keySignature.Address

	childKey, _ := key.DerivedKeyWithPath(fromAddr.HDPath, decoder.wm.Config.CurveType)
	PrivateKey, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		log.Error("get private key bytes, err=", err)
		return err
	}

	tx_hash := common.FromHex(keySignature.Message) //TX.Hash

	signed, err := SignRawTransaction(PrivateKey, tx_hash)
	if err != nil {
		log.Error("signature error !")
		return nil
	}

	//TX.Sign = signed

	keySignature.Signature = hex.EncodeToString(signed)

	log.Debug("** pri:", hex.EncodeToString(PrivateKey))
	log.Debug("** tx_hash:", keySignature.Message)
	log.Debug("** Signature:", keySignature.Signature)

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
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

	message := common.FromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Message)     //TX.Hash
	signature := common.FromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Signature) //TX.Sign
	publicKey := common.FromHex(rawTx.Signatures[rawTx.Account.AccountID][0].Address.PublicKey)
	//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
	verify_result := VerifyRawTransaction(publicKey, message, signature)
	if verify_result == owcrypt.SUCCESS {
		log.Debug("transaction verify passed")
		rawTx.IsCompleted = true

	} else {
		log.Debug("transaction verify failed")
		rawTx.IsCompleted = false
	}

	return nil
}

func (decoder *TransactionDecoder) SubmitSimpleRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	//check交易交易单基本字段
	err := CheckRawTransaction(rawTx)
	if err != nil {
		openwLogger.Log.Errorf("Verify raw tx failed, err=%v", err)
		return err
	}
	if len(rawTx.Signatures) != 1 {
		openwLogger.Log.Errorf("len of signatures error. ")
		return errors.New("len of signatures error. ")
	}

	accSignatures, exist := rawTx.Signatures[rawTx.Account.AccountID]
	if !exist {
		openwLogger.Log.Errorf("wallet[%v] signature not found ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(accSignatures) == 0 {
		openwLogger.Log.Errorf("wallet[%v] signature is empty ", rawTx.Account.AccountID)
		return errors.New("wallet signature not found ")
	}

	if len(rawTx.RawHex) == 0 {
		return fmt.Errorf("transaction hex is empty")
	}

	if !rawTx.IsCompleted {
		return fmt.Errorf("transaction is not completed validation")
	}

	keySignature := accSignatures[0]

	tx, err := DecodeRawHexToTransaction(rawTx.RawHex)
	if err != nil {
		return err
	}

	tx.Sign = common.FromHex(keySignature.Signature)

	submitRawHex, err := EncodeTransaction(tx)
	if err != nil {
		return err
	}

	txid, err := decoder.wm.SubmitRawTransaction(submitRawHex)
	if err != nil {
		return err
	} else {
		//广播成功后记录nonce值到本地
		//fmt.Printf("Submit Success , Save nonce To AddressExtParam!\n")
		wrapper.SetAddressExtParam(rawTx.Signatures[rawTx.Account.AccountID][0].Address.Address, decoder.wm.FullName(), rawTx.Signatures[rawTx.Account.AccountID][0].Nonce)
	}
	rawTx.TxID = txid
	rawTx.IsSubmit = true

	//fmt.Printf("rawTx=%+v\n", rawTx)

	return nil
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {
	if !rawTx.Coin.IsContract {
		return decoder.SubmitSimpleRawTransaction(wrapper, rawTx)
	}

	return nil
	//wjq return decoder.SubmitErc20TokenRawTransaction(wrapper, rawTx)
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {

	rate := decoder.wm.EstimateFeeRate()
	rate_decimal := decimal.RequireFromString(rate).Div(coinDecimal)

	return rate_decimal.StringFixed(decoder.wm.Decimal()), "NAS", nil
}
