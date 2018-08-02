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

package workflow

// import (
// 	"bufio"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/astaxie/beego/config"
// 	// "github.com/tidwall/gjson"
// 	// "github.com/blocktree/OpenWallet/common"
// 	// "github.com/blocktree/OpenWallet/common/file"
// 	"github.com/blocktree/OpenWallet/keystore"
// 	"github.com/bndr/gotabulate"
// 	// "github.com/btcsuite/btcd/chaincfg"
// 	// "github.com/btcsuite/btcutil"
// 	// "github.com/btcsuite/btcutil/hdkeychain"
// 	// "github.com/codeskyblue/go-sh"
// 	"github.com/pkg/errors"
// 	"github.com/shopspring/decimal"
// 	"io/ioutil"
// 	"log"
// 	// "math"
// 	"os"
// 	"path/filepath"
// 	// "sort"
// 	"strings"
// 	//"time"
// )

//ListUnspent 获取未花记录
// func ListUnspent(min uint64) ([]*Unspent, error) {
//
// 	var (
// 		utxos = make([]*Unspent, 0)
// 	)
//
// 	request := []interface{}{
// 		min,
// 	}
//
// 	result, err := client.Call("listunspent", "GET", request)
// 	if err != nil {
// 		return nil, err
// 	}
// 	fmt.Println("Result=", result)
//
// 	// array := result.Array()
// 	// for _, a := range array {
// 	// 	utxos = append(utxos, NewUnspent(&a))
// 	// }
//
// 	return utxos, nil
//
// }

//
// //RebuildWalletUnspent 批量插入未花记录到本地
// func RebuildWalletUnspent(walletID string) error {
//
// 	wallet, err := GetWalletInfo(walletID)
// 	if err != nil {
// 		return err
// 	}
//
// 	//查找核心钱包确认数大于1的
// 	utxos, err := ListUnspent(1)
// 	if err != nil {
// 		return err
// 	}
//
// 	db, err := wallet.OpenDB()
// 	if err != nil {
// 		return err
// 	}
// 	defer db.Close()
//
// 	//清空历史的UTXO
// 	db.Drop("Unspent")
//
// 	//开始事务
// 	tx, err := db.Begin(true)
// 	if err != nil {
// 		return err
// 	}
// 	defer tx.Rollback()
//
// 	//批量插入到本地数据库
// 	//设置utxo的钱包账户
// 	for _, utxo := range utxos {
// 		var addr Address
// 		err = db.One("Address", utxo.Address, &addr)
// 		utxo.Account = addr.Account
// 		utxo.HDAddress = addr
// 		key := common.NewString(fmt.Sprintf("%s_%d_%s", utxo.TxID, utxo.Vout, utxo.Address)).SHA256()
// 		utxo.Key = key
//
// 		err = tx.Save(utxo)
// 		if err != nil {
// 			return err
// 		}
// 	}
//
// 	return tx.Commit()
// }
//
//ListUnspentFromLocalDB 查询本地数据库的未花记录
// func ListUnspentFromLocalDB(walletID string) ([]*Unspent, error) {
// 	var (
// 		wallet *Wallet
// 	)
//
// 	wallets, err := GetWalletKeys(keyDir)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	//获取钱包余额
// 	for _, w := range wallets {
// 		if w.WalletID == walletID {
// 			wallet = w
// 			break
// 		}
// 	}
//
// 	if wallet == nil {
// 		return nil, errors.New("The wallet that your given name is not exist!")
// 	}
//
// 	db, err := wallet.OpenDB()
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer db.Close()
//
// 	var utxos []*Unspent
// 	err = db.Find("Account", walletID, &utxos)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	return utxos, nil
// }

// //BuildTransaction 构建交易单
// func BuildTransaction(utxos []*Unspent, to, change string, amount, fees decimal.Decimal) (string, decimal.Decimal, error) {
//
// 	var (
// 		inputs      = make([]interface{}, 0)
// 		outputs     = make(map[string]interface{})
// 		totalAmount = decimal.New(0, 0)
// 	)
//
// 	for _, u := range utxos {
//
// 		if u.Spendable {
// 			ua, _ := decimal.NewFromString(u.Amount)
// 			totalAmount = totalAmount.Add(ua)
//
// 			inputs = append(inputs, map[string]interface{}{
// 				"txid": u.TxID,
// 				"vout": u.Vout,
// 			})
// 		}
//
// 	}
//
// 	if totalAmount.LessThan(amount) {
// 		return "", decimal.New(0, 0), errors.New("The balance is not enough!")
// 	}
//
// 	changeAmount := totalAmount.Sub(amount).Sub(fees)
// 	if changeAmount.GreaterThan(decimal.New(0, 0)) {
// 		//ca, _ := changeAmount.Float64()
// 		outputs[change] = changeAmount.StringFixed(8)
//
// 		fmt.Printf("Create change address for receiving %s coin.\n", outputs[change])
// 	}
//
// 	//ta, _ := amount.Float64()
// 	outputs[to] = amount.StringFixed(8)
//
// 	request := []interface{}{
// 		inputs,
// 		outputs,
// 	}
//
// 	rawTx, err := client.Call("createrawtransaction", request)
// 	if err != nil {
// 		return "", decimal.New(0, 0), err
// 	}
//
// 	return rawTx.String(), changeAmount, nil
// }
//
// //SignRawTransaction 钱包交易单
// func SignRawTransaction(txHex, walletID string, key *keystore.HDKey, utxos []*Unspent) (string, error) {
//
// 	var (
// 		wifs = make([]string, 0)
// 	)
//
// 	//查找未花签名需要的私钥
// 	for _, u := range utxos {
//
// 		childKey, err := key.DerivedKeyWithPath(u.HDAddress.HDPath)
//
// 		privateKey, err := childKey.ECPrivKey()
// 		if err != nil {
// 			return "", err
// 		}
//
// 		cfg := chaincfg.MainNetParams
// 		if isTestNet {
// 			cfg = chaincfg.TestNet3Params
// 		}
//
// 		wif, err := btcutil.NewWIF(privateKey, &cfg, true)
// 		if err != nil {
// 			return "", err
// 		}
//
// 		wifs = append(wifs, wif.String())
//
// 	}
//
// 	request := []interface{}{
// 		txHex,
// 		utxos,
// 		wifs,
// 	}
//
// 	result, err := client.Call("signrawtransaction", request)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return result.Get("hex").String(), nil
//
// }
//
// //SendRawTransaction 广播交易
// func SendRawTransaction(txHex string) (string, error) {
//
// 	request := []interface{}{
// 		txHex,
// 	}
//
// 	result, err := client.Call("sendrawtransaction", request)
// 	if err != nil {
// 		return "", err
// 	}
//
// 	return result.String(), nil
//
// }
//
// //SendTransaction 发送交易
// func SendTransaction(walletID, to string, amount decimal.Decimal, password string, feesInSender bool) ([]string, error) {
//
// 	var (
// 		usedUTXO   []*Unspent
// 		balance    = decimal.New(0, 0)
// 		totalSend  = amount
// 		actualFees = decimal.New(0, 0)
// 		sendTime   = 1
// 		txIDs      = make([]string, 0)
// 	)
//
// 	utxos, err := ListUnspentFromLocalDB(walletID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	//获取utxo，按小到大排序
// 	sort.Sort(UnspentSort{utxos, func(a, b *Unspent) int {
//
// 		if a.Amount > b.Amount {
// 			return 1
// 		} else {
// 			return -1
// 		}
// 	}})
//
// 	//读取钱包
// 	w, err := GetWalletInfo(walletID)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	totalBalance, _ := decimal.NewFromString(w.Balance)
// 	if totalBalance.LessThanOrEqual(amount) && feesInSender {
// 		return nil, errors.New("The wallet's balance is not enough!")
// 	} else if totalBalance.LessThan(amount) && !feesInSender {
// 		return nil, errors.New("The wallet's balance is not enough!")
// 	}
//
// 	//加载钱包
// 	key, err := w.HDKey(password)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	//解锁钱包
// 	err = UnlockWallet(password, 120)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	//创建找零地址
// 	changeAddr, err := CreateChangeAddress(walletID, key)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	feesRate, err := EstimateFeeRate()
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	fmt.Printf("Calculating wallet unspent record to build transaction...\n")
//
// 	//循环的计算余额是否足够支付发送数额+手续费
// 	for {
//
// 		usedUTXO = make([]*Unspent, 0)
// 		balance = decimal.New(0, 0)
//
// 		//计算一个可用于支付的余额
// 		for _, u := range utxos {
//
// 			if u.Spendable {
// 				ua, _ := decimal.NewFromString(u.Amount)
// 				balance = balance.Add(ua)
// 				usedUTXO = append(usedUTXO, u)
// 				if balance.GreaterThanOrEqual(totalSend) {
// 					break
// 				}
// 			}
// 		}
//
// 		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
// 		fees, err := EstimateFee(int64(len(usedUTXO)), 2, feesRate)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		//如果要手续费有发送支付，得计算加入手续费后，计算余额是否足够
// 		if feesInSender {
// 			//总共要发送的
// 			totalSend = amount.Add(fees)
// 			if totalSend.GreaterThan(balance) {
// 				continue
// 			}
// 			totalSend = amount
// 		} else {
//
// 			if fees.GreaterThanOrEqual(amount) {
// 				return nil, errors.New("The sent amount is not enough for fees!")
// 			}
//
// 			totalSend = amount.Sub(fees)
// 		}
//
// 		actualFees = fees
//
// 		break
//
// 	}
//
// 	changeAmount := balance.Sub(totalSend).Sub(actualFees)
//
// 	fmt.Printf("-----------------------------------------------\n")
// 	fmt.Printf("From WalletID: %s\n", walletID)
// 	fmt.Printf("To Address: %s\n", to)
// 	fmt.Printf("Use: %v\n", balance.StringFixed(8))
// 	fmt.Printf("Fees: %v\n", actualFees.StringFixed(8))
// 	fmt.Printf("Receive: %v\n", totalSend.StringFixed(8))
// 	fmt.Printf("Change: %v\n", changeAmount.StringFixed(8))
// 	fmt.Printf("-----------------------------------------------\n")
//
// 	//UTXO如果大于设定限制，则分拆成多笔交易单发送
// 	if len(usedUTXO) > maxTxInputs {
// 		sendTime = int(math.Ceil(float64(len(usedUTXO)) / float64(maxTxInputs)))
// 	}
//
// 	for i := 0; i < sendTime; i++ {
//
// 		var sendUxto []*Unspent
// 		var pieceOfSend = decimal.New(0, 0)
//
// 		s := i * maxTxInputs
//
// 		//最后一个，计算余数
// 		if i == sendTime-1 {
// 			sendUxto = usedUTXO[s:]
//
// 			pieceOfSend = totalSend
// 		} else {
// 			sendUxto = usedUTXO[s : s+maxTxInputs]
//
// 			for _, u := range sendUxto {
// 				ua, _ := decimal.NewFromString(u.Amount)
// 				pieceOfSend = pieceOfSend.Add(ua)
// 			}
//
// 		}
//
// 		//计算手续费，找零地址有2个，一个是发送，一个是新创建的
// 		piecefees, err := EstimateFee(int64(len(sendUxto)), 2, feesRate)
// 		if piecefees.LessThan(decimal.NewFromFloat(0.00001)) {
// 			piecefees = decimal.NewFromFloat(0.00001)
// 		}
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		//解锁钱包
// 		err = UnlockWallet(password, 120)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		//log.Printf("pieceOfSend = %s \n", pieceOfSend.StringFixed(8))
// 		//log.Printf("piecefees = %s \n", piecefees.StringFixed(8))
// 		//log.Printf("feesRate = %s \n", feesRate.StringFixed(8))
//
// 		//创建交易
// 		txRaw, _, err := BuildTransaction(sendUxto, to, changeAddr.Address, pieceOfSend, piecefees)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		fmt.Printf("Build Transaction Successfully\n")
//
// 		//签名交易
// 		signedHex, err := SignRawTransaction(txRaw, walletID, key, sendUxto)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		fmt.Printf("Sign Transaction Successfully\n")
//
// 		txid, err := SendRawTransaction(signedHex)
// 		if err != nil {
// 			return nil, err
// 		}
//
// 		fmt.Printf("Submit Transaction Successfully\n")
//
// 		txIDs = append(txIDs, txid)
// 		//txIDs = append(txIDs, signedHex)
//
// 		//减去已发送的
// 		totalSend = totalSend.Sub(pieceOfSend)
//
// 	}
//
// 	//发送成功后，删除已使用的UTXO
// 	clearUnspends(usedUTXO, w)
//
// 	LockWallet()
//
// 	return txIDs, nil
// }
// //clearUnspends 清楚已使用的UTXO
// func clearUnspends(utxos []*Unspent, wallet *Wallet) {
// 	db, err := wallet.OpenDB()
// 	if err != nil {
// 		return
// 	}
// 	defer db.Close()
//
// 	//开始事务
// 	tx, err := db.Begin(true)
// 	if err != nil {
// 		return
// 	}
// 	defer tx.Rollback()
//
// 	for _, utxo := range utxos {
// 		tx.DeleteStruct(utxo)
// 	}
//
// 	tx.Commit()
// }
//
