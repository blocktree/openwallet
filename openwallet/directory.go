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

import (
	"path/filepath"
	"strings"
)

//GetDataDir 钱包数据目录
func GetDataDir(symbol string) string {
	return filepath.Join("data", strings.ToLower(symbol))
}

//GetKeyDir 密钥目录
func GetKeyDir(symbol string) string {
	return filepath.Join(GetDataDir(symbol), "key")
}

//GetDBDir 钱包数据库目录
func GetDBDir(symbol string) string {
	return filepath.Join(GetDataDir(symbol), "db")
}

//GetBackupDir 钱包备份目录
func GetBackupDir(symbol string) string {
	return filepath.Join(GetDataDir(symbol), "backup")
}

//GetExportAddressDir 导出地址目录
func GetExportAddressDir(symbol string) string {
	return filepath.Join(GetDataDir(symbol), "address")
}