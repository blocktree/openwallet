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

package tron

var (
	tw *WalletManager

	TOADDRESS            = "TSdXzXKSQ3RQzQ5Ge8TiYfMQEjofSVQ8ax"                               // account: t2
	OWNERADDRESS         = "TNQkiUv4qtDRKWDrKS628FTbDwxLMiqbAz"                               // account: t1
	PRIVATEKEY           = "6c5e3afd3d6c0394dc922d9bcaf98fd9c972aa226948b44e14a7e4b0566c69ca" // account: t1
	AMOUNT        = "0.1"

	TXRAW    = "0a7e0a023c462208c84cf406d3b89d2640ffbd85e0fa2c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a1541887661d2e0215851756b1e7933216064526badcd121541b6c1abf9fb31c9077dfb3c25469e6e943ffbfa7a18a08d06"
	TXSIGNED = "0a7e0a02d5842208a43b6160eaa543f840f8ffb4a9e12c5a67080112630a2d747970652e676f6f676c65617069732e636f6d2f70726f746f636f6c2e5472616e73666572436f6e747261637412320a154199fee02e1ee01189bc41a68e9069b7919ef2ad82121541e11973395042ba3c0b52b4cdf4e15ea77818f27518c0843d1241915254f565f20a327047e59327bf7f0b9e5600c09a321e6da45de4be4158f1f14bbf520ac1a0d71da10696ee80dfa5860086a6e470c96e7a275c4433616b8beb00"
)

func init() {

	tw = NewWalletManager()

	//tw.Config.ServerAPI = "http://127.0.0.1:28090"
	// tw.Config.ServerAPI = "http://127.0.0.1:18090"
	//tw.Config.ServerAPI = "http://192.168.2.194:18090"
	//tw.Config.ServerAPI = "http://192.168.2.194:28090"
	tw.Config.ServerAPI = "http://192.168.27.124:18090"
	// tw.Config.RpcUser = "walletUser"
	// tw.Config.RpcPassword = "walletPassword2017"
	// token := BasicAuth(tw.Config.RpcUser, tw.Config.RpcPassword)
	tw.WalletClient = NewClient(tw.Config.ServerAPI, "", true)
}
