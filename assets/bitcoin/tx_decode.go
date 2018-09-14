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

package bitcoin

import (
	"errors"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/go-OWCBasedFuncs/btcLikeTxDriver"
	"github.com/shopspring/decimal"
	"sort"
	"strings"
	"encoding/hex"
)

//CreateRawTransaction 创建交易单
func (wm *WalletManager) CreateRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {

	var (
		vins      = make([]btcLikeTxDriver.Vin, 0)
		vouts     = make([]btcLikeTxDriver.Vout, 0)
		txUnlocks = make([]btcLikeTxDriver.TxUnlock, 0)

		usedUTXO []*Unspent
		balance  = decimal.New(0, 0)
		//totalSend  = amounts
		totalSend    = decimal.New(0, 0)
		actualFees   = decimal.New(0, 0)
		feesRate     = decimal.New(0, 0)
		accountID    = rawTx.Account.AccountID
		destinations = make([]string, 0)
	)

	address, err := wrapper.GetAddressList(0, -1, "AccountID", rawTx.Account.AccountID)
	if err != nil {
		return err
	}

	searchAddrs := make([]string, 0)
	for _, address := range address {
		searchAddrs = append(searchAddrs, address.Address)
	}

	//查找账户的utxo
	unspents, err := wm.ListUnspent(0, searchAddrs...)
	if err != nil {
		return err
	}

	if len(rawTx.To) == 0 {
		return errors.New("Receiver addresses is empty!")
	}

	//计算总发送金额
	for addr, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		totalSend = totalSend.Add(deamount)
		destinations = append(destinations, addr)
	}

	//获取utxo，按小到大排序
	sort.Sort(UnspentSort{unspents, func(a, b *Unspent) int {

		if a.Amount > b.Amount {
			return 1
		} else {
			return -1
		}
	}})

	//totalBalance, _ := decimal.NewFromString(wm.GetWalletBalance(w.WalletID))
	//if totalBalance.LessThanOrEqual(totalSend) {
	//	return "", errors.New("The wallet's balance is not enough!")
	//}

	//创建找零地址
	//changeAddrs, err := wrapper.CreateAddress(accountID, 1, wm.Decoder, true, wm.Config.IsTestNet)
	////changeAddr, err := wm.CreateChangeAddress(walletID, key)
	//if err != nil {
	//	return err
	//}
	//
	//changeAddress := changeAddrs[0]

	//取账户最后一个地址
	changeAddress := address[len(address)-1]

	if len(rawTx.FeeRate) == 0 {
		feesRate, err = wm.EstimateFeeRate()
		if err != nil {
			return err
		}
	} else {
		feesRate, _ = decimal.NewFromString(rawTx.FeeRate)
	}

	log.Info("Calculating wallet unspent record to build transaction...")
	computeTotalSend := totalSend
	//循环的计算余额是否足够支付发送数额+手续费
	for {

		usedUTXO = make([]*Unspent, 0)
		balance = decimal.New(0, 0)

		//计算一个可用于支付的余额
		for _, u := range unspents {

			if u.Spendable {
				ua, _ := decimal.NewFromString(u.Amount)
				balance = balance.Add(ua)
				usedUTXO = append(usedUTXO, u)
				if balance.GreaterThanOrEqual(computeTotalSend) {
					break
				}
			}
		}

		if balance.LessThan(computeTotalSend) {
			return errors.New("The balance is not enough!")
		}

		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
		fees, err := wm.EstimateFee(int64(len(usedUTXO)), int64(len(destinations)+1), feesRate)
		if err != nil {
			return err
		}

		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
		//总共要发送的
		computeTotalSend = totalSend.Add(fees)
		if computeTotalSend.GreaterThan(balance) {
			continue
		}
		computeTotalSend = totalSend

		actualFees = fees

		break

	}

	//UTXO如果大于设定限制，则分拆成多笔交易单发送
	if len(usedUTXO) > wm.Config.MaxTxInputs {
		errStr := fmt.Sprintf("The transaction is use max inputs over: %d", wm.Config.MaxTxInputs)
		return errors.New(errStr)
	}

	changeAmount := balance.Sub(balance).Sub(actualFees)

	log.Std.Notice("-----------------------------------------------\n")
	log.Std.Notice("From Account: %s\n", accountID)
	log.Std.Notice("To Address: %s\n", strings.Join(destinations, ", "))
	log.Std.Notice("Use: %v\n", balance.StringFixed(8))
	log.Std.Notice("Fees: %v\n", actualFees.StringFixed(8))
	log.Std.Notice("Receive: %v\n", computeTotalSend.StringFixed(8))
	log.Std.Notice("Change: %v\n", changeAmount.StringFixed(8))
	log.Std.Notice("-----------------------------------------------\n")

	//装配输入
	for _, utxo := range usedUTXO {
		in := btcLikeTxDriver.Vin{utxo.TxID, uint32(utxo.Vout)}
		vins = append(vins, in)

		txUnlock := btcLikeTxDriver.TxUnlock{LockScript: utxo.ScriptPubKey}
		txUnlocks = append(txUnlocks, txUnlock)
	}

	//装配输入
	for to, amount := range rawTx.To {
		deamount, _ := decimal.NewFromString(amount)
		deamount = deamount.Mul(wm.Config.CoinDecimal)
		out := btcLikeTxDriver.Vout{to, uint64(deamount.IntPart())}
		vouts = append(vouts, out)
	}

	//changeAmount := balance.Sub(totalSend).Sub(actualFees)
	if changeAmount.GreaterThan(decimal.New(0, 0)) {
		deamount := changeAmount.Mul(wm.Config.CoinDecimal)
		out := btcLikeTxDriver.Vout{changeAddress.Address, uint64(deamount.IntPart())}
		vouts = append(vouts, out)

		//fmt.Printf("Create change address for receiving %s coin.\n", outputs[change])
	}

	//锁定时间
	lockTime := uint32(0)

	//追加手续费支持
	replaceable := false

	/////////构建空交易单
	emptyTrans, err := btcLikeTxDriver.CreateEmptyRawTransaction(vins, vouts, lockTime, replaceable)

	if err != nil {
		log.Error("构建空交易单失败")
	} else {
		fmt.Println("空交易单：")
		fmt.Println(emptyTrans)
	}

	////////构建用于签名的交易单哈希
	transHash, err := btcLikeTxDriver.CreateRawTransactionHashForSig(emptyTrans, txUnlocks)
	if err != nil {
		log.Error("获取待签名交易单哈希失败")
	} else {
		for i, t := range transHash {
			fmt.Println("第", i+1, "个交易单哈希值为")
			fmt.Println(t)
		}
	}

	rawTx.RawHex = emptyTrans

	if rawTx.Signatures == nil {
		rawTx.Signatures = make(map[string][]openwallet.KeySignature)
	}

	//装配签名

	keySigs := make([]openwallet.KeySignature, 0)

	for i, unlock := range txUnlocks {

		beSignHex := transHash[i]

		addr, err := wrapper.GetAddress(unlock.Address)
		if err != nil {
			return err
		}

		signature := openwallet.KeySignature{
			EccType: wm.Config.CurveType,
			Nonce:   "",
			Address: addr,
			Message: beSignHex,
		}

		keySigs = append(keySigs, signature)

	}

	rawTx.Signatures[rawTx.Account.AccountID] = keySigs

	return nil
}

//SignRawTransaction 签名交易单
func (wm *WalletManager) SignRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	return nil
}

//SendRawTransaction 广播交易单
func (wm *WalletManager) SubmitRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {
	return nil
}

//VerifyRawTransaction 验证交易单，验证交易单并返回加入签名后的交易单
func (wm *WalletManager) VerifyRawTransaction(wrapper *openwallet.WalletWrapper, rawTx *openwallet.RawTransaction) error {

	var (
		txUnlocks = make([]btcLikeTxDriver.TxUnlock, 0)
		emptyTrans = rawTx.RawHex
		sigPub = make([]btcLikeTxDriver.SignaturePubkey, 0)
	)


	for _, keySignatures := range rawTx.Signatures {
		for _, keySignature := range keySignatures {

			signature, _ := hex.DecodeString(keySignature.Signature)
			pubkey, _ := hex.DecodeString(keySignature.Address.PublicKey)

			signaturePubkey := btcLikeTxDriver.SignaturePubkey {
				Signature: signature,
				Pubkey: pubkey,
			}

			sigPub = append(sigPub, signaturePubkey)
		}
	}

	txBytes, err := hex.DecodeString(emptyTrans)
	if err != nil {
		return errors.New("Invalid transaction hex data!")
	}

	trx := btcLikeTxDriver.DecodeRawTransaction(txBytes)


	for _, vin := range trx.Vins {

		utxo, err := wm.GetTxOut(vin.GetTxID(), uint64(vin.GetVout()))
		if err != nil {
			return err
		}

		txUnlock := btcLikeTxDriver.TxUnlock{LockScript: utxo.Get("scriptPubKey.hex").String()}
		txUnlocks = append(txUnlocks, txUnlock)

	}

	////////填充签名结果到空交易单
	//  传入TxUnlock结构体的原因是： 解锁向脚本支付的UTXO时需要对应地址的赎回脚本， 当前案例的对应字段置为 "" 即可
	signedTrans, err := btcLikeTxDriver.InsertSignatureIntoEmptyTransaction(emptyTrans, sigPub, txUnlocks)
	if err != nil {
		log.Error("交易单拼接失败")
	} else {
		fmt.Println("拼接后的交易单")
		fmt.Println(signedTrans)
	}

	/////////验证交易单
	//验证时，对于公钥哈希地址，需要将对应的锁定脚本传入TxUnlock结构体
	pass := btcLikeTxDriver.VerifyRawTransaction(signedTrans, txUnlocks)
	if pass {
		fmt.Println("验证通过")
	} else {
		log.Error("验证失败")
	}

	return nil
}
