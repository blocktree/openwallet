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

package manager

import "path/filepath"

var (
	defaultDataDir = filepath.Join(".", "openw_data")
)

type Config struct {
	keyDir    string //钥匙备份路径
	dbPath    string //本地数据库文件路径
	backupDir string //备份路径
	isTestnet bool   //是否测试网
	supportAssets []string //支持的资产类型
}

func NewConfig() *Config {

	c := Config{}

	//钥匙备份路径
	c.keyDir = filepath.Join(defaultDataDir, "key")
	//本地数据库文件路径
	c.dbPath = filepath.Join(defaultDataDir, "db")
	//备份路径
	c.backupDir = filepath.Join(defaultDataDir, "backup")

	c.supportAssets = []string{"BTC"}

	return &c
}

//loadConfig 加载配置文件
//@param path 配置文件路径
func loadConfig(path string) *Config {
	return nil
}
