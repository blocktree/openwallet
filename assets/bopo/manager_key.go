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

package bopo

// import (
// 	"bufio"
// 	"encoding/json"
// 	"fmt"
// 	"github.com/astaxie/beego/config"
// 	// "github.com/tidwall/gjson"
// 	// "github.com/blocktree/openwallet/common"
// 	// "github.com/blocktree/openwallet/common/file"
// 	"github.com/blocktree/openwallet/keystore"
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

//
// //ImportPrivKey 导入私钥
// func ImportPrivKey(wif, walletID string) error {
//
// 	request := []interface{}{
// 		wif,
// 		walletID,
// 		false,
// 	}
//
// 	_, err := client.Call("importprivkey", request)
//
// 	if err != nil {
// 		return err
// 	}
//
// 	return err
//
// }
//
// //ImportMulti 批量导入地址和私钥
// func ImportMulti(addresses []*Address, keys []string, walletID string, watchOnly bool) ([]int, error) {
//
// 	/*
// 		[
// 		{
// 			"scriptPubKey" : { "address": "1NL9w5fP9kX2D9ToNZPxaiwFJCngNYEYJo" },
// 			"timestamp" : 0,
// 			"label" : "Personal"
// 		},
// 		{
// 			"scriptPubKey" : "76a9149e857da0a5b397559c78c98c9d3f7f655d19c68688ac",
// 			"timestamp" : 1493912405,
// 			"label" : "TestFailure"
// 		}
// 		]' '{ "rescan": true }'
// 	*/
//
// 	var (
// 		request     []interface{}
// 		imports     = make([]interface{}, 0)
// 		failedIndex = make([]int, 0)
// 	)
//
// 	if len(addresses) != len(keys) {
// 		return nil, errors.New("Import addresses is not equal keys count!")
// 	}
//
// 	for i, a := range addresses {
// 		k := keys[i]
// 		obj := map[string]interface{}{
// 			"scriptPubKey": map[string]interface{}{
// 				"address": a.Address,
// 			},
// 			"label":     walletID,
// 			"timestamp": "now",
// 			"watchonly": watchOnly,
// 		}
//
// 		if !watchOnly {
// 			obj["keys"] = []string{k}
// 		}
//
// 		imports = append(imports, obj)
// 	}
//
// 	request = []interface{}{
// 		imports,
// 		map[string]interface{}{
// 			"rescan": false,
// 		},
// 	}
//
// 	result, err := client.Call("importmulti", request)
// 	if err != nil {
// 		return nil, err
// 	}
//
// 	for i, r := range result.Array() {
// 		if !r.Get("success").Bool() {
// 			failedIndex = append(failedIndex, i)
// 		}
// 	}
//
// 	return failedIndex, err
//
// }
//
// //KeyPoolRefill 重新填充私钥池
// func KeyPoolRefill(keyPoolSize uint64) error {
//
// 	request := []interface{}{
// 		keyPoolSize,
// 	}
//
// 	_, err := client.Call("keypoolrefill", request)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
// }
//

//
// //CreateNewPrivateKey 创建私钥，返回私钥wif格式字符串
// func CreateNewPrivateKey(key *keystore.HDKey, start, index uint64) (string, *Address, error) {
//
// 	derivedPath := fmt.Sprintf("%s/%d/%d", key.RootPath, start, index)
// 	//fmt.Printf("derivedPath = %s\n", derivedPath)
// 	childKey, err := key.DerivedKeyWithPath(derivedPath)
//
// 	privateKey, err := childKey.ECPrivKey()
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	cfg := chaincfg.MainNetParams
// 	if isTestNet {
// 		cfg = chaincfg.TestNet3Params
// 	}
//
// 	wif, err := btcutil.NewWIF(privateKey, &cfg, true)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	address, err := childKey.Address(&cfg)
// 	if err != nil {
// 		return "", nil, err
// 	}
//
// 	addr := Address{
// 		Address:   address.String(),
// 		Account:   key.RootId,
// 		HDPath:    derivedPath,
// 		CreatedAt: time.Now(),
// 	}
//
// 	return wif.String(), &addr, err
// }
//
// //CreateBatchPrivateKey
// func CreateBatchPrivateKey(key *keystore.HDKey, count uint64) ([]string, error) {
//
// 	var (
// 		wifs = make([]string, 0)
// 	)
//
// 	start := time.Now().Unix()
// 	for i := uint64(0); i < count; i++ {
// 		wif, _, err := CreateNewPrivateKey(key, uint64(start), i)
// 		if err != nil {
// 			continue
// 		}
// 		wifs = append(wifs, wif)
// 	}
//
// 	return wifs, nil
//
// }
//
//
// //DumpWallet 导出钱包所有私钥文件
// func DumpWallet(filename string) error {
//
// 	request := []interface{}{
// 		filename,
// 	}
//
// 	_, err := client.Call("dumpwallet", request)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
//
// }
//
// //ImportWallet 导入钱包私钥文件
// func ImportWallet(filename string) error {
//
// 	request := []interface{}{
// 		filename,
// 	}
//
// 	_, err := client.Call("importwallet", request)
// 	if err != nil {
// 		return err
// 	}
//
// 	return nil
//
// }
