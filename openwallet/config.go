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

package openwallet

import "github.com/astaxie/beego/config"

//AssetsConfig 用于给AssetsAdapter调用者初始化、加载外部配置的接口
type AssetsConfig interface {

	//LoadExternalConfig 加载外部配置
	LoadAssetsConfig(c config.Configer) error

	//InitDefaultConfig 初始化默认配置
	InitAssetsConfig() (config.Configer, error)
}

type AssetsConfigBase struct{}

//LoadAssetsConfig 加载外部配置
func (as *AssetsConfigBase) LoadAssetsConfig(c config.Configer) error {
	return nil
}

//InitAssetsConfig 初始化默认配置
func (as *AssetsConfigBase) InitAssetsConfig() (config.Configer, error) {
	return nil, nil
}
