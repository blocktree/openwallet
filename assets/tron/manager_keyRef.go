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

package tron

import (
	"encoding/hex"
	"fmt"
	"log"
	// "github.com/blocktree/OpenWallet/log"
)

//GetPrivateKeyRef 找到地址索引对应的私钥，返回私钥hex格式字符串
func (wm *WalletManager) GetPrivateKeyRef(walletID, password string, index uint64, serializes uint32) (string, error) {

	//读取钱包文件
	w, err := wm.GetWalletInfo(walletID)
	if err != nil {
		log.Println(err)
		return "", err
	}

	//解密钱包文件并加载钱包KEY
	k, err := w.HDKey(password)
	if err != nil {
		log.Println(err)
		return "", err
	}

	derivedPath := fmt.Sprintf("%s/%d", k.RootPath, index)
	key, err := k.DerivedKeyWithPath(derivedPath, wm.Config.CurveType)
	if err != nil {
		log.Println(err)
		return "", err
	}

	childKey, err := key.GenPrivateChild(uint32(serializes))
	if err != nil {
		return "", err
	}

	keyBytes, err := childKey.GetPrivateKeyBytes()
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(keyBytes), nil
}
