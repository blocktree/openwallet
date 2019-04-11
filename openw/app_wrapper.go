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

package openw

import (
	"github.com/asdine/storm"
	"github.com/blocktree/openwallet/openwallet"
)

type AppID string

//AppWrapper 应用装器，扩展应用功能
type AppWrapper struct {
	*Wrapper
	appID string
}

func NewAppWrapper(args ...interface{}) *AppWrapper {

	wrapper := NewWrapper(args...)

	appWrapper := AppWrapper{Wrapper: wrapper}

	for _, arg := range args {
		switch obj := arg.(type) {
		case AppID:
			appWrapper.appID = string(obj)
		case *Wrapper:
			appWrapper.Wrapper = obj
		}
	}

	return &appWrapper
}

// GetWalletInfo
func (wrapper *AppWrapper) GetWalletInfo(walletID string) (*openwallet.Wallet, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var wallet openwallet.Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// GetWalletList
func (wrapper *AppWrapper) GetWalletList(offset, limit int) ([]*openwallet.Wallet, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var wallets []*openwallet.Wallet
	err = db.All(&wallets, storm.Limit(limit), storm.Skip(offset))
	if err != nil {
		return nil, err
	}

	return wallets, nil
}
