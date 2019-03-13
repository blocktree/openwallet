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

package openwallet

import "github.com/asdine/storm"

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
func (wrapper *AppWrapper) GetWalletInfo(walletID string) (*Wallet, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var wallet Wallet
	err = db.One("WalletID", walletID, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

// GetWalletList
func (wrapper *AppWrapper) GetWalletList(offset, limit int) ([]*Wallet, error) {

	//打开数据库
	db, err := wrapper.OpenStormDB()
	if err != nil {
		return nil, err
	}
	defer wrapper.CloseDB()

	var wallets []*Wallet
	err = db.All(&wallets, storm.Limit(limit), storm.Skip(offset))
	if err != nil {
		return nil, err
	}

	return wallets, nil
}
