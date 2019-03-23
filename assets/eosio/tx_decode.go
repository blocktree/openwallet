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

package eosio

import (
	"encoding/hex"
	"fmt"
	"github.com/blocktree/openwallet/log"
	"github.com/blocktree/openwallet/openwallet"
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/btcsuite/btcd/btcec"
	"github.com/eoscanada/eos-go/ecc"
	"github.com/eoscanada/eos-go/token"
	"github.com/shopspring/decimal"
	"time"
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

//CreateRawTransaction 创建交易单
func (decoder *TransactionDecoder) CreateRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	var (
		accountTotalSent = decimal.Zero
		txFrom           = make([]string, 0)
		txTo             = make([]string, 0)
		keySignList      = make([]*openwallet.KeySignature, 0)
		accountID        = rawTx.Account.AccountID
		accountBalance   eos.Asset
		amountStr        string
		to               eos.AccountName
	)

	codeAccount := rawTx.Coin.Contract.Address
	tokenCoin := rawTx.Coin.Contract.Token
	//tokenDecimals := rawTx.Coin.Contract.Decimals

	//获取wallet
	account, err := wrapper.GetAssetsAccountInfo(accountID)
	if err != nil {
		return err
	}

	if account.Alias == "" {
		return fmt.Errorf("[%s] have not been created", accountID)
	}

	from := eos.AccountName(eos.AccountName(account.Alias))

	//accountResp, err := decoder.wm.Api.GetAccount(eos.AccountName(account.Alias))
	//if err != nil {
	//	return fmt.Errorf("eos account not found on chain")
	//}

	accountAssets, err := decoder.wm.Api.GetCurrencyBalance(eos.AccountName(account.Alias), tokenCoin, eos.AccountName(codeAccount))
	if len(accountAssets) == 0 {
		return fmt.Errorf("eos account balance is not enough")
	}

	accountBalance = accountAssets[0]

	for k, v := range rawTx.To {
		to = eos.AccountName(k)
		amountStr = v
		break
	}

	accountBalanceDec := decimal.New(int64(accountBalance.Amount), -int32(accountBalance.Precision))
	amountDec, _ := decimal.NewFromString(amountStr)

	if accountBalanceDec.LessThan(amountDec) {
		return fmt.Errorf("the balance: %s is not enough", amountStr)
	}

	amountInt64 := amountDec.Shift(int32(accountBalance.Precision)).IntPart()
	quantity := eos.Asset{Amount: eos.Int64(amountInt64), Symbol: accountBalance.Symbol}
	memo := rawTx.ExtParam

	txOpts := &eos.TxOptions{}
	if err := txOpts.FillFromChain(decoder.wm.Api); err != nil {
		return fmt.Errorf("filling tx opts: %s", err)
	}
	tx := eos.NewTransaction([]*eos.Action{token.NewTransfer(from, to, quantity, memo)}, txOpts)
	stx := eos.NewSignedTransaction(tx)
	txdata, cfd, err := stx.PackedTransactionAndCFD()
	if err != nil {
		return err
	}
	//交易哈希
	sigDigest := eos.SigDigest(txOpts.ChainID, txdata, cfd)

	addresses, err := wrapper.GetAddressList(0, -1,
		"AccountID", accountID)
	if err != nil {
		return err
	}

	if len(addresses) == 0 {
		return fmt.Errorf("[%s] have not EOS public key", accountID)
	}

	for _, addr := range addresses {
		signature := openwallet.KeySignature{
			EccType: decoder.wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: hex.EncodeToString(sigDigest),
		}
		keySignList = append(keySignList, &signature)
	}

	//计算账户的实际转账amount
	accountTotalSentAddresses, findErr := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID, "Address", to)
	if findErr != nil || len(accountTotalSentAddresses) == 0 {
		amountDec, _ := decimal.NewFromString(amountStr)
		accountTotalSent = accountTotalSent.Add(amountDec)
	}
	accountTotalSent = decimal.Zero.Sub(accountTotalSent)

	txFrom = []string{fmt.Sprintf("%s:%s", from, amountStr)}
	txTo = []string{fmt.Sprintf("%s:%s", to, amountStr)}

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]*openwallet.KeySignature)
	}

	rawTx.RawHex = hex.EncodeToString(txdata)
	rawTx.Signatures[rawTx.Account.AccountID] = keySignList
	rawTx.FeeRate = ""
	rawTx.Fees = ""
	rawTx.IsBuilt = true
	rawTx.TxAmount = accountTotalSent.String()
	rawTx.TxFrom = txFrom
	rawTx.TxTo = txTo

	return nil

}

//SignRawTransaction 签名交易单
func (decoder *TransactionDecoder) SignRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	key, err := wrapper.HDKey()
	if err != nil {
		return err
	}

	keySignatures := rawTx.Signatures[rawTx.Account.AccountID]
	if keySignatures != nil {
		for _, keySignature := range keySignatures {

			childKey, err := key.DerivedKeyWithPath(keySignature.Address.HDPath, keySignature.EccType)
			keyBytes, err := childKey.GetPrivateKeyBytes()
			if err != nil {
				return err
			}
			decoder.wm.Log.Debug("privateKey:", hex.EncodeToString(keyBytes))

			privKey, _ := btcec.PrivKeyFromBytes(btcec.S256(), keyBytes)
			hash, err := hex.DecodeString(keySignature.Message)
			if err != nil {
				return fmt.Errorf("decoder transaction hash failed, unexpected err: %v", err)
			}

			decoder.wm.Log.Debug("hash:", hash)

			sig, err := privKey.SignCanonical(btcec.S256(), hash)
			if err != nil {
				return fmt.Errorf("sign transaction hash failed, unexpected err: %v", err)
			}

			keySignature.Signature = hex.EncodeToString(sig)
		}
	}

	decoder.wm.Log.Info("transaction hash sign success")

	rawTx.Signatures[rawTx.Account.AccountID] = keySignatures

	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (decoder *TransactionDecoder) VerifyRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) error {

	if rawTx.Signatures == nil || len(rawTx.Signatures) == 0 {
		//this.wm.Log.Std.Error("len of signatures error. ")
		return fmt.Errorf("transaction signature is empty")
	}

	var tx eos.Transaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}
	err = eos.UnmarshalBinary(txHex, &tx)
	if err != nil {
		return fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}

	stx := eos.NewSignedTransaction(&tx)

	//支持多重签名
	for accountID, keySignatures := range rawTx.Signatures {
		decoder.wm.Log.Debug("accountID Signatures:", accountID)
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)

			stx.Signatures = append(
				stx.Signatures,
				ecc.Signature{Curve: ecc.CurveK1, Content: signature},
			)

			decoder.wm.Log.Debug("Signature:", keySignature.Signature)
			decoder.wm.Log.Debug("PublicKey:", keySignature.Address.PublicKey)
		}
	}

	//packed, err := stx.Pack(eos.CompressionNone)
	//if err != nil {
	//	return err
	//}

	//TODO:验证签名是否通过

	bin, err := eos.MarshalBinary(stx)
	if err != nil {
		return fmt.Errorf("signed transaction encode failed, unexpected error: %v", err)
	}

	rawTx.IsCompleted = true
	rawTx.RawHex = hex.EncodeToString(bin)

	return nil
}

//SendRawTransaction 广播交易单
func (decoder *TransactionDecoder) SubmitRawTransaction(wrapper openwallet.WalletDAI, rawTx *openwallet.RawTransaction) (*openwallet.Transaction, error) {

	var stx eos.SignedTransaction
	txHex, err := hex.DecodeString(rawTx.RawHex)
	if err != nil {
		return nil, fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}
	err = eos.UnmarshalBinary(txHex, &stx)
	if err != nil {
		return nil, fmt.Errorf("transaction decode failed, unexpected error: %v", err)
	}

	packedTx, err := stx.Pack(eos.CompressionNone)
	if err != nil {
		return nil, err
	}

	response, err := decoder.wm.Api.PushTransaction(packedTx)
	if err != nil {
		return nil, fmt.Errorf("push transaction: %s", err)
	}

	log.Infof("Transaction [%s] submitted to the network successfully.", hex.EncodeToString(response.Processed.ID))

	rawTx.TxID = hex.EncodeToString(response.Processed.ID)
	rawTx.IsSubmit = true

	decimals := int32(rawTx.Coin.Contract.Decimals)
	fees := "0"

	//记录一个交易单
	tx := &openwallet.Transaction{
		From:       rawTx.TxFrom,
		To:         rawTx.TxTo,
		Amount:     rawTx.TxAmount,
		Coin:       rawTx.Coin,
		TxID:       rawTx.TxID,
		Decimal:    decimals,
		AccountID:  rawTx.Account.AccountID,
		Fees:       fees,
		SubmitTime: time.Now().Unix(),
	}

	tx.WxID = openwallet.GenTransactionWxID(tx)

	return tx, nil
}

//GetRawTransactionFeeRate 获取交易单的费率
func (decoder *TransactionDecoder) GetRawTransactionFeeRate() (feeRate string, unit string, err error) {
	return "", "", fmt.Errorf("not implement")
}

//CreateSummaryRawTransaction 创建汇总交易
func (decoder *TransactionDecoder) CreateSummaryRawTransaction(wrapper openwallet.WalletDAI, sumRawTx *openwallet.SummaryRawTransaction) ([]*openwallet.RawTransaction, error) {
	return nil, fmt.Errorf("CreateSummaryRawTransaction not implement")
}
