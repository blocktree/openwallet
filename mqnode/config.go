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

package mqnode

import (
	"errors"
	"fmt"
	"github.com/astaxie/beego/config"
	"github.com/blocktree/OpenWallet/common/file"
	"path/filepath"
	"github.com/blocktree/OpenWallet/manager"
	"github.com/blocktree/OpenWallet/owtp"
	"strings"
)

var (
	//商户资料文件夹
	merchantDir = filepath.Join("ow_data")
	//商户资料缓存
	cacheFile = "openw_cache.db"
	//配置文件名
	configFileName = "openw_ini.ini"
	//默认配置内容
	defaultConfig = `

# openw node ID
server_node_id = "testnode"

# local node publicKey 
publickey = ""

# local node privateKey 
privatekey = ""

# mq url
merchant_node_url = ""

# mq exchange
exchange = "DEFAULT_EXCHANGE"

# mq sendQueueName
send_queue_name = "Test"

# mq receiveQueueName
receive_queue_name = "Test"

# mq account
account = ""

# mq password
password = ""

# manage key_dir 
key_dir = "key"

# manage db_path 
db_path = "db"

# manage backup file 
backup = "backup"

# manage supportAssets( Split by ",")
support_assets = "BTC"

# manage enable_block_scan
enable_block_scan = true

# manage is_test_net
is_test_net = true
`
)

//printConfig Print config information
func printConfig() error {

	initConfig()
	//读取配置
	absFile := filepath.Join(merchantDir, configFileName)
	fmt.Printf("-----------------------------------------------------------\n")
	file.PrintFile(absFile)
	fmt.Printf("-----------------------------------------------------------\n")

	return nil

}

//initConfig 初始化配置文件
func initConfig() {

	//读取配置
	absFile := filepath.Join(merchantDir, configFileName)
	if !file.Exists(absFile) {
		file.MkdirAll(merchantDir)
		file.WriteFile(absFile, []byte(defaultConfig), false)
	}

}

//loadConfig 读取配置
func loadConfig() (NodeConfig, error) {

	var (
		c       config.Configer
		err     error
		configs = NodeConfig{}
	)

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return configs, errors.New("Config is not setup. Please run ' openw server config -i ' ")
	}
	//节点id
	configs.MerchantNodeID = c.String("server_node_id")
	//本地公钥
	configs.LocalPublicKey = c.String("publickey")
	//本地私钥
	configs.LocalPrivateKey = c.String("privatekey")
	//缓存文件
	configs.CacheFile = filepath.Join(merchantDir, cacheFile)
	//mq地址
	configs.MerchantNodeURL = c.String("merchant_node_url")
	//链接方式
	configs.ConnectType = owtp.MQ
	//exchange
	configs.Exchange = c.String("exchange")
	//发送队列名
	configs.QueueName = c.String("send_queue_name")
	//接收队列名
	configs.ReceiveQueueName = c.String("receive_queue_name")
	//mq账户
	configs.Account = c.String("account")
	//mq密码
	configs.Password = c.String("password")
	return configs, nil
}

//loadConfig 读取配置
func loadManagerConfig() (*manager.Config, error) {

	var (
		c       config.Configer
		err     error
		configs = manager.Config{}
	)

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err = config.NewConfig("ini", absFile)
	if err != nil {
		return nil, errors.New("Config is not setup. Please run ' openw server config -i ' ")
	}

	defaultDataDir := filepath.Join(".", "openw_data")

	//钥匙备份路径
	configs.KeyDir = filepath.Join(defaultDataDir, c.String("key_dir"))
	//本地数据库文件路径
	configs.DBPath = filepath.Join(defaultDataDir, c.String("db_path"))
	//备份路径
	configs.BackupDir = filepath.Join(defaultDataDir, c.String("backup"))

	//支持资产用户","分割
	supportAssetsStr := c.String("support_assets")
	supportAssets := strings.Split(supportAssetsStr, ",")
	if supportAssets != nil && len(supportAssets) > 0 {
		configs.SupportAssets = supportAssets
	}

	//开启区块扫描
	configs.EnableBlockScan, err = c.Bool("enable_block_scan")
	//测试网
	configs.IsTestnet, err = c.Bool("is_test_net")
	return &configs, nil
}
