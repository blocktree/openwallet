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
	"github.com/blocktree/OpenWallet/timer"
	"log"
	"github.com/blocktree/OpenWallet/assets"
	"github.com/blocktree/OpenWallet/openwallet"
)

//GetChargeAddressVersion 获取要订阅的地址版本信息
func (m *MerchantNode) GetChargeAddressVersion() error {

	var (
		err  error
		subs = make([]*Subscription, 0)
	)

	//检查是否连接
	err = m.IsConnected()
	if err != nil {
		return err
	}

	//db, err := m.OpenDB()
	//if err != nil {
	//	return err
	//}

	m.mu.RLock()
	for _, s := range m.subscriptions {
		if s.Type == SubscribeTypeCharge {
			subs = append(subs, s)
		}
	}
	m.mu.RUnlock()

	//err = db.Find("type", SubscribeTypeCharge, &subs)
	//if err != nil {
	//	return err
	//}
	//
	//db.Close()

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
			func(addressVer *AddressVersion, status uint64, msg string) {

				if addressVer != nil {

					innerdb, err := m.OpenDB()
					if err != nil {
						return
					}
					defer innerdb.Close()
					var oldVersion AddressVersion
					err = innerdb.One("Key", addressVer.Key, &oldVersion)
					//if err != nil {
					//	return
					//}
					//log.Printf("old version = %d", oldVersion.Version)
					//log.Printf("new version = %d", addressVer.Version)
					if addressVer.Version > oldVersion.Version || err != nil {

						//TODO:加入到订阅地址通道
						m.getAddressesCh <- *addressVer

						//更新记录
						innerdb.Save(addressVer)
					}

				}

			})
	}

	return nil
}

//GetChargeAddress 获取地址
func (m *MerchantNode) getChargeAddress() error {

	var (
		err   error
		limit = uint64(20)
	)

	////检查是否连接
	//err = m.IsConnected()
	//if err != nil {
	//	return err
	//}

	//log.Printf("getChargeAddress running...\n")

	for {

		select {
		//接收通道发送的地址版本
		case v := <-m.getAddressesCh:

			getCount := uint64(0)

			//log.Printf("get address version: %v", v)
			for i := uint64(0); i < v.Total; i = i + limit {

				params := struct {
					Coin     string `json:"coin"`
					WalletID string `json:"walletID"`
					Offset   uint64 `json:"offset"`
					Limit    uint64 `json:"limit"`
				}{v.Coin, v.WalletID, i, limit}

				err = GetChargeAddress(m.Node, params,
					true,
					func(addrs []*openwallet.Address, status uint64, msg string) {

						if status == owtp.StatusSuccess {

							wallet, err := m.GetMerchantWalletByID(v.WalletID)
							if err != nil {
								return
							}

							//导入到每个币种的数据库
							mer := assets.GetMerchantAssets(v.Coin)
							mer.ImportMerchantAddress(wallet, addrs)

							getCount = getCount + limit
						}
					})
				if err != nil {
					log.Printf("GetChargeAddress unexpected error: %v", err)
					continue
				}

			}

			//检查地址条数是否完整
			if getCount != v.Total {
				// 扔到通道中，重新下载地址
				// 如果一直都获取不完整，或者对方统计的地址不对，就会使这个进程死循环
				//m.getAddressesCh <- v

				//删除这个不对的版本，重新在下一轮获取
				m.DeleteAddressVersion(&v)
			}

		}
	}

	return nil
}

//runSubscribeAddressTask 运行订阅地址任务
func (m *MerchantNode) runSubscribeAddressTask() {

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.subscribeAddressTask == nil {
		//启动钱包汇总程序
		task := timer.NewTask(PeriodOfTask, m.updateSubscribeAddress)
		m.subscribeAddressTask = task
		m.subscribeAddressTask.Start()

		//开启获取地址消费者
		go m.getChargeAddress()
	}
	log.Printf("Start Subscribe Address Task...\n")
	m.subscribeAddressTask.Restart()
}

//updateSubscribeAddress 更新地址
func (m *MerchantNode) updateSubscribeAddress() {

	var (
		err error
	)

	m.mu.RLock()
	m.mu.RUnlock()

	if len(m.subscriptions) == 0 {
		return
	}

	log.Printf("Update Subscribe Address...\n")
	//获取订阅地址的最新版本
	err = m.GetChargeAddressVersion()
	if err != nil {
		log.Printf("GetChargeAddressVersion unexpected error: %v", err)
	}
}
