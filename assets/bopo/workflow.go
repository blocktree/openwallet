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

package bopo

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
//
// const (
// 	maxAddresNum = 10000
// )

//
// //EstimateFee 预估手续费
// func EstimateFee(inputs, outputs int64, feeRate decimal.Decimal) (decimal.Decimal, error) {
//
// 	var piece int64 = 1
//
// 	//UTXO如果大于设定限制，则分拆成多笔交易单发送
// 	if inputs > int64(maxTxInputs) {
// 		piece = int64(math.Ceil(float64(inputs) / float64(maxTxInputs)))
// 	}
//
// 	//计算公式如下：148 * 输入数额 + 34 * 输出数额 + 10
// 	trx_bytes := decimal.New(inputs*148+outputs*34+piece*10, 0)
// 	trx_fee := trx_bytes.Div(decimal.New(1000, 0)).Mul(feeRate)
//
// 	return trx_fee, nil
// }
//
// //EstimateFeeRate 预估的没KB手续费率
// func EstimateFeeRate() (decimal.Decimal, error) {
//
// 	defaultRate, _ := decimal.NewFromString("0.0001")
//
// 	//估算交易大小 手续费
// 	request := []interface{}{
// 		2,
// 	}
//
// 	result, err := client.Call("estimatefee", request)
// 	if err != nil {
// 		return decimal.New(0, 0), err
// 	}
//
// 	feeRate, _ := decimal.NewFromString(result.String())
//
// 	if feeRate.LessThan(defaultRate) {
// 		feeRate = defaultRate
// 	}
//
// 	return feeRate, nil
// }
//
// skipKeyFile ignores editor backups, hidden files and folders/symlinks.
// func skipKeyFile(fi os.FileInfo) bool {
// 	// Skip editor backups and UNIX-style hidden files.
// 	if strings.HasSuffix(fi.Name(), "~") || strings.HasPrefix(fi.Name(), ".") {
// 		return true
// 	}
// 	// Skip misc special files, directories (yes, symlinks too).
// 	if fi.IsDir() || fi.Mode()&os.ModeType != 0 {
// 		return true
// 	}
//
// 	return false
// }
