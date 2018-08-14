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

package merchant

import (
	"github.com/blocktree/OpenWallet/owtp"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"path/filepath"
	"github.com/astaxie/beego/config"
	"errors"
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
		return errors.New("Config is not setup. Please run 'wmd merchant config -i ' ")
	}

	priv := c.String("privatekey")
	pub := c.String("publickey")

	printKeychain(priv, pub)

	return nil
}

func InitMerchantKeychainFlow() error {

	//随机创建证书
	cert := owtp.NewRandomCertificate("")
	if len(cert.PrivateKeyBytes()) == 0 {
		log.Error("Create keychain failed!")
		return fmt.Errorf("Create keychain failed ")
	}

	priv, pub := cert.KeyPair()

	//写入到配置文件
	initConfig()
	absFile := filepath.Join(merchantDir, configFileName)
	c, err := config.NewConfig("ini", absFile)
	if err != nil {
		return errors.New("Config is not setup. Please run 'wmd merchant config -i ' ")
	}

	c.Set("privatekey", priv)
	c.Set("publickey", pub)
	c.SaveConfigFile(absFile)

	log.Info("Create keychain successfully.")

	printKeychain(priv, pub)

	return nil
}

//JoinMerchantNodeFlow 连接商户服务节点
func JoinMerchantNodeFlow() error {

	var (
		err error
		c   NodeConfig
	)

	c, err = loadConfig()
	if err != nil {
		return err
	}

	merchantNode, err = NewMerchantNode(c)
	if err != nil {
		return err
	}

	merchantNode.Run()

	return nil
}

func ConfigMerchantFlow() error {
	initConfig()
	return nil
}

func ShowMechantConfig() error {
	printConfig()
	return nil
}

//printKeychain 打印证书钥匙串
func printKeychain(priv string, pub string) {

	if len(priv) == 0 {
		priv = "nothing"
	}

	if len(pub) == 0 {
		pub = "nothing"
	}

	//打印证书信息
	log.Notice("--------------- PRIVATE KEY ---------------")
	log.Notice(priv)
	fmt.Println()
	log.Notice("--------------- PUBLIC KEY ---------------")
	log.Notice(pub)
	fmt.Println()
}