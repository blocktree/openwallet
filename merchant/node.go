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
	"errors"
	"fmt"
	"github.com/asdine/storm"
	"github.com/blocktree/OpenWallet/owtp"
	"log"
	"time"
)

//商户节点
type MerchantNode struct {

	//节点配置
	Config NodeConfig
	//商户节点
	Node *owtp.OWTPNode
	//连接状态通道
	reconnect chan bool
	//断开状态通道
	disconnected chan struct{}
	//是否重连
	isReconnect bool
	//重连时的等待时间
	ReconnectWait time.Duration
}

func NewMerchantNode(config NodeConfig) (*MerchantNode, error) {

	m := MerchantNode{}

	if len(config.MerchantNodeURL) == 0 {
		return nil, errors.New("merchant node url is not configed!")
	}

	//授权配置
	auth, err := owtp.NewOWTPAuth(
		config.NodeKey,
		config.PublicKey,
		config.PrivateKey,
		true,
		config.CacheFile,
	)

	if err != nil {
		return nil, err
	}

	//创建节点，连接商户
	node := owtp.NewOWTPNode(config.NodeID, config.MerchantNodeURL, auth)

	m.Node = node
	m.Config = config

	//断开连接后，重新连接
	m.Node.SetCloseHandler(func(n *owtp.OWTPNode) {
		log.Printf("merchantNode disconnect. \n")
		m.disconnected <- struct{}{}
	})

	m.isReconnect = true
	m.ReconnectWait = 10

	//设置路由
	m.setupRouter()

	return &m, nil
}

//OpenDB 访问数据库
func (m *MerchantNode) OpenDB() (*storm.DB, error) {
	return storm.Open(m.Config.CacheFile)
}

//Run 运行商户节点管理
func (m *MerchantNode) Run() error {

	var (
		err error
	)

	defer func() {
		close(m.reconnect)
		close(m.disconnected)
	}()

	m.reconnect = make(chan bool, 1)
	m.disconnected = make(chan struct{}, 1)

	//启动连接
	m.reconnect <- true

	log.Printf("Merchant node running now... \n")

	//节点运行时
	for {
		select {
		case <-m.reconnect:
			//重新连接
			log.Printf("Connecting to %s\n", m.Node.URL)
			err = m.Node.Connect()
			if err != nil {
				log.Printf("Connect merchant node faild unexpected error: %v. \n", err)
				m.disconnected <- struct{}{}
			} else {
				log.Printf("Connect merchant node successfully. \n")
			}
		case <-m.disconnected:
			if m.isReconnect {
				//重新连接，前等待
				log.Printf("Reconnect after %d seconds... \n", m.ReconnectWait)
				time.Sleep(m.ReconnectWait * time.Second)
				m.reconnect <- true
			} else {
				//退出
				break
			}
		}
	}

	return nil
}

//IsConnected 检查商户节点是否连接
func (m *MerchantNode) IsConnected() error {

	if m.Node == nil {
		return ErrMerchantNodeDisconnected
	}

	if !m.Node.IsConnected() {
		return ErrMerchantNodeDisconnected
	}
	return nil
}

//Stop 停止运行
func (m *MerchantNode) Stop() {
	close(m.disconnected)
}

/********** 商户服务相关方法【主动】 **********/

//GetChargeAddressVersion 获取要订阅的地址版本信息
func (m *MerchantNode) GetChargeAddressVersion() error {

	var (
		err error
	)

	//检查是否连接
	err = m.IsConnected()
	if err != nil {
		return err
	}

	db, err := m.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	var subs []Subscription
	err = db.Find("type", SubscribeTypeCharge, &subs)
	if err != nil {
		return err
	}

	for _, sub := range subs {

		//| 参数名称 | 类型   | 是否可空 | 描述     |
		//|----------|--------|----------|----------|
		//| coin     | string | 是       | 币名     |
		//| walletID | string | 是       | 钱包ID   |

		params := struct {
			Coin     string `json:"coin"`
			WalletID string `json:"walletID"`
		}{sub.Coin, sub.WalletID}

		//获取订阅的地址版本
		GetChargeAddressVersion(m.Node, params,
			true,
			func(addressVer *AddressVersion) {

				if addressVer != nil {
					db.Save(addressVer)
				}

			})
	}

	var avs []AddressVersion
	err = db.All(&avs)
	if err != nil {
		return err
	}

	fmt.Printf("%v\n", avs)

	return nil
}

func (m *MerchantNode) GetChargeAddress() {

	//var (
	//	completed bool
	//)
	//
	//if merchantNode == nil {
	//	return
	//}
	//
	//for {
	//	//获取地址
	//	merchantNode.Call(
	//		"getChargeAddress",
	//		nil,
	//		true,
	//		func(resp owtp.Response) {
	//			if resp.Status == 0 {
	//				completed = true
	//			}
	//		})
	//
	//	if completed == true {
	//		break
	//	}
	//}
}
