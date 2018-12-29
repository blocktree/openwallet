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

package owtp

import (
	"github.com/tidwall/gjson"
	"net"
)

//DataPacket 数据包
type DataPacket struct {
	/*

		本协议传输数据，格式编码采用json。消息接收与发送，都遵循数据包规范定义字段内容。

		| 参数名 | 类型   | 示例             | 描述                                                                                |
		|--------|--------|------------------|-----------------------------------------------------------------------------------|
		| r      | uint8  | 1                | 传输类型，1：请求，2：响应                                                              |
		| m      | string | subscribe        | 方法名，对应接口方法定义                                                             |
		| n      | uint32 | 123              | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。   |
		| t      | uint32 | 1528520843       | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                       |
		| d      | Object | {"foo": "hello"} | 数据主体，请求内容或响应内容。接口方法说明中，主要说明这部分。                          |
		| s      | string | Qwse             | [可选]合并[r+m+n+t+d]进行sha256两次ECC签名并base58编码，用于校验数据的一致性和合法性 |
		| z      | string | 1b24ac           | [可选]协商校验值SA，开启协商密码后，有协商过程计算出来SA值，以保证节点通信的密钥一致   |
		| v      | string | 1                | 版本号                                                                              |

	*/

	Req       uint64      `json:"r"`
	Method    string      `json:"m"`
	Nonce     uint64      `json:"n" storm:"id"`
	Timestamp int64       `json:"t"`
	Data      interface{} `json:"d"`
	Signature string      `json:"s"`
	CheckCode string      `json:"z"`
	Version   string      `json:"v"`
	PublicKey string      `json:"a"`
}

//NewDataPacket 通过 gjson转为DataPacket
func NewDataPacket(json gjson.Result) *DataPacket {
	dp := &DataPacket{}
	dp.Req = json.Get("r").Uint()
	dp.Method = json.Get("m").String()
	dp.Nonce = json.Get("n").Uint()
	dp.Timestamp = json.Get("t").Int()
	dp.Data = json.Get("d").String()
	dp.Signature = json.Get("s").String()
	dp.CheckCode = json.Get("z").String()
	dp.Version = json.Get("v").String()
	dp.PublicKey = json.Get("a").String()
	return dp
}

type PeerInfo struct {
	ID     string
	Config ConnectConfig
}

type PeerAttribute map[string]interface{}

// Peer 节点
type Peer interface {

	/* 公开方法 */

	PID() string                  //节点ID
	IsHost() bool                 //是否主机，我方主动连接的节点
	IsConnected() bool            //是否已经连接
	LocalAddr() net.Addr          //本地节点地址
	RemoteAddr() net.Addr         //远程节点地址
	ConnectConfig() ConnectConfig // 返回配置信息

	/* 内部方法 */

	auth() Authorization
	setHandler(handler PeerHandler) error //设置节点的服务者
	openPipe() error                      //OpenPipe 打开通道
	send(data DataPacket) error           //发送请求
	close() error                         //关闭节点
}

// PeerHandler 节点监听器
type PeerHandler interface {
	OnPeerOpen(peer Peer)                                      //节点连接成功
	OnPeerClose(peer Peer, reason string)                      //节点关闭
	OnPeerNewDataPacketReceived(peer Peer, packet *DataPacket) //节点获取新数据包
	GetValueForPeer(peer Peer, key string) interface{}
	PutValueForPeer(peer Peer, key string, val interface{}) error
}
