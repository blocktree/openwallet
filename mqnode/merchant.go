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
	"github.com/blocktree/OpenWallet/owtp"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"path/filepath"
	"github.com/astaxie/beego/config"
	"errors"
	"github.com/blocktree/OpenWallet/manager"
)

func init() {
	owtp.Debug = true
}

func GetMerchantKeychain() error {

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	priv := c.String("privatekey")
	pub := c.String("publickey")
	id := c.String("server_node_id")
	sendQueueName := c.String("send_queue_name")
	receiveQueueName := c.String("receive_queue_name")
	printKeychain(id,priv, pub,sendQueueName,receiveQueueName)

	return nil
}

func InitServiceKeychain() error {

	//随机创建证书
	cert := owtp.NewRandomCertificate()
	if len(cert.PrivateKeyBytes()) == 0 {
		log.Error("Create keychain failed!")
		return fmt.Errorf("Create keychain failed ")
	}

	priv, pub := cert.KeyPair()
	id := cert.ID()
	//写入到配置文件
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	//创建收发通道
	sendQueueName := "OW_RPC_GO"
	receiveQueueName :=  id+"_JAVA"

	c.Set("server_node_id", id)
	c.Set("privatekey", priv)
	c.Set("publickey", pub)
	c.Set("send_queue_name",sendQueueName )
	c.Set("receive_queue_name",receiveQueueName )
	c.SaveConfigFile(absFile)

	log.Info("Create keychain successfully.")

	printKeychain(id,priv, pub,sendQueueName,receiveQueueName)

	return nil
}

//RunServer 连接服务服务节点
func RunServer() error {

	var (
		err error
		c   NodeConfig
	)

	c, err = loadConfig()
	if err != nil {
		return err
	}

	merchantNode, err = NewBitNodeNode(c)
	if err != nil {
		return err
	}
	config,err := loadManagerConfig()
	if err != nil {
		return err
	}
	merchantNode.manager = openw.NewWalletManager(config)
	merchantNode.Run()

	return nil
}

func ConfigService() error {
	initConfig()
	return nil
}

func ShowServiceConfig() error {
	printConfig()
	return nil
}

//printKeychain 打印证书钥匙串
func printKeychain(id ,priv,pub,sendQueueName,receiveQueueName string) {

	if len(priv) == 0 {
		priv = "nothing"
	}

	if len(pub) == 0 {
		pub = "nothing"
	}

	//打印证书信息
	log.Notice("--------------- ID ---------------")
	log.Notice(id)
	fmt.Println()
	log.Notice("--------------- PRIVATE KEY ---------------")
	log.Notice(priv)
	fmt.Println()
	log.Notice("--------------- PUBLIC KEY ---------------")
	log.Notice(pub)
	fmt.Println()
	log.Notice("--------------- SEND QUEUE NAME ---------------")
	log.Notice(sendQueueName)
	fmt.Println()
	log.Notice("--------------- RECEIVE QUEUE NAME ---------------")
	log.Notice(receiveQueueName)
	fmt.Println()
}

func SetSymbolAssests(s string) error {

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	c.Set("support_assets",s)
	c.SaveConfigFile(absFile)

	return nil
}

func SetMQUrl(url string) error {

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	c.Set("merchant_node_url",url)
	c.SaveConfigFile(absFile)

	return nil
}

func SetMQAccount(account string) error {

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	c.Set("account",account)
	c.SaveConfigFile(absFile)

	return nil
}

func SetMQPassword(account string) error {

	//读取配置
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'openw node config -i ' ")
	}

	c.Set("account",account)
	c.SaveConfigFile(absFile)

	return nil
}