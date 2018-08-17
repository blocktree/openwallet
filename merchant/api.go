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
	"github.com/blocktree/OpenWallet/log"
	"github.com/blocktree/OpenWallet/openwallet"
	"github.com/blocktree/OpenWallet/owtp"
)

//CallGetChargeAddressVersion 获取要订阅的地址版本信息
func GetChargeAddressVersion(
	node *owtp.OWTPNode,
	nodeID string,
	params interface{},
	sync bool,
	callback func(addressVer *AddressVersion, status uint64, msg string)) error {

	//| 参数名称 | 类型   | 是否可空 | 描述     |
	//|----------|--------|----------|----------|
	//| coin     | string | 是       | 币名     |
	//| walletID | string | 是       | 钱包ID   |

	method := "getChargeAddressVersion"
	log.Info("Merchant Call:", method)

	//获取订阅的地址版本
	err := node.Call(
		nodeID,
		method,
		params,
		sync,
		func(resp owtp.Response) {
			log.Info(method, ":", "Response:", resp.JsonData())
			result := resp.JsonData()
			if result.Exists() {
				addressVersion := NewAddressVersion(result)
				callback(addressVersion, resp.Status, resp.Msg)
			} else {
				callback(nil, resp.Status, resp.Msg)
			}

		})

	return err
}

//GetChargeAddress 获取订阅的地址表
func GetChargeAddress(
	node *owtp.OWTPNode,
	nodeID string,
	params interface{},
	sync bool,
	callback func(addrs []*openwallet.Address, status uint64, msg string)) error {

	//| 参数名称 | 类型   | 是否可空 | 描述     |
	//|----------|--------|----------|----------|
	//| coin     | string | 是       | 币名     |
	//| walletID | string | 是       | 钱包ID   |
	//| offset   | int    | 是       | 偏移量   |
	//| limit    | int    | 是       | 读取条数 |

	method := "getChargeAddress"
	log.Info("Merchant Call:", method)
	//获取订阅的地址版本
	err := node.Call(
		nodeID,
		method,
		params,
		sync,
		func(resp owtp.Response) {
			log.Info(method, ":", "Response:", resp.JsonData())
			addrs := make([]*openwallet.Address, 0)
			result := resp.JsonData()
			if result.Exists() {
				arrs := result.Get("addresses").Array()
				for _, ad := range arrs {
					a := openwallet.NewAddress(ad)
					addrs = append(addrs, a)
				}
			}

			callback(addrs, resp.Status, resp.Msg)

		})

	return err
}

//SubmitRechargeTransaction 提交充值到账记录
func SubmitRechargeTransaction(
	node *owtp.OWTPNode,
	nodeID string,
	params interface{},
	sync bool,
	callback func(confirms []uint64, status uint64, msg string)) error {
	method := "submitRechargeTransaction"
	log.Info("Merchant Call:", method)

	//获取订阅的地址版本
	err := node.Call(
		nodeID,
		method,
		params,
		sync,
		func(resp owtp.Response) {
			log.Info(method, ":", "Response:", resp.JsonData())
			confirms := make([]uint64, 0)
			result := resp.JsonData()
			if result.Exists() {
				arrs := result.Get("confirms").Array()
				for _, c := range arrs {
					confirms = append(confirms, c.Uint())
				}
			}
			callback(confirms, resp.Status, resp.Msg)

		})

	return err
}
