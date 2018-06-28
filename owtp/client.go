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
	"encoding/json"
	"github.com/asdine/storm"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"log"
	"path/filepath"
	"time"
)

//局部常量
const (
	WriteWait      = 120 * time.Second //超时为6秒
	PongWait       = 60 * time.Second
	PingPeriod     = (PongWait * 9) / 10
	MaxMessageSize = 8 * 1024 // 最大消息缓存KB
	WSRequest      = 1        //wesocket请求标识
	WSResponse     = 2        //wesocket响应标识
)

var (
	//默认的缓存文件路径
	defaultCacheFile = filepath.Join(".", "cahce", "openw.db")
	//重放限制时长，数据包的时间戳超过后，这个间隔，可以重复nonce
	replayLimit = 7 * 24 * time.Hour
)

type Handler interface {
	ServeOWTP(client *Client, ctx *Context)
}

//Client 基于websocket的通信客户端
type Client struct {
	ws        *websocket.Conn
	handler   Handler
	send      chan []byte
	cacheFile string
}

//DataPacket 数据包
type DataPacket struct {

	/*

		本协议传输数据，格式编码采用json。消息接收与发送，都遵循数据包规范定义字段内容。

		| 参数名 | 类型   | 示例             | 描述                                                                              |
		|--------|--------|------------------|-----------------------------------------------------------------------------------|
		| r      | uint8  | 1                | 传输类型，1：请求，2：响应                                                            |
		| m      | string | subscribe        | 方法名，对应接口方法定义                                                           |
		| n      | uint32 | 123              | 请求序号。为了保证请求对应响应按序执行，并防御重放攻击，序号可以为随机数，但不可重复。 |
		| t      | uint32 | 1528520843       | 时间戳。限制请求在特定时间范围内有效，如10分钟。                                     |
		| d      | Object | {"foo": "hello"} | 数据主体，请求内容或响应内容。接口方法说明中，主要说明这部分。                        |
		| s      | string | Qwse==           | 合并[r+m+n+t+d]进行sha256两次签名并base64编码，用于校验数据的一致性和合法性       |

	*/

	Req       uint64      `json:"r"`
	Method    string      `json:"m"`
	Nonce     uint64      `json:"n" storm:"id"`
	Timestamp int64       `json:"t"`
	Data      interface{} `json:"d"`
	Signature string      `json:"s"`
}

type Response struct {
	Status uint64      `json:"status"`
	Msg    string      `json:"msg"`
	Result interface{} `json:"result"`
}

type Context struct {
	//传输类型，1：请求，2：响应
	Req int
	//请求序号
	nonce uint64
	//传入参数
	Params interface{}
	//方法
	Method string
	//响应
	Resp Response
}

//NewDataPacket 通过 gjson转为DataPacket
func NewDataPacket(json gjson.Result) *DataPacket {
	dp := &DataPacket{}
	dp.Req = json.Get("r").Uint()
	dp.Method = json.Get("m").String()
	dp.Nonce = json.Get("n").Uint()
	dp.Timestamp = json.Get("t").Int()
	dp.Data = json.Get("d").Value()
	dp.Signature = json.Get("s").String()
	return dp
}

// Dial connects a client to the given URL.
func Dial(url string, router Handler, cacheFile string) (*Client, error) {

	if router == nil {
		return nil, errors.New("router should not be nil!")
	}

	log.Printf("connecting to %s\n", url)
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}

	if len(cacheFile) > 0 {
		cacheFile = defaultCacheFile
	}

	client := &Client{
		ws:        c,
		handler:   router,
		cacheFile: cacheFile,
		send:      make(chan []byte, MaxMessageSize),
	}

	//发送通道
	go client.writePump()

	//监听消息
	go client.readPump()

	return client, nil
}

//Send 发送消息
func (c *Client) Send(data DataPacket) error {
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	c.send <- respBytes
	return nil
}

// WritePump 发送消息通道
func (c *Client) writePump() {

	ticker := time.NewTicker(PingPeriod) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
		c.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			//发送消息
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			log.Printf("write: %s\n", string(message))
			if err := c.write(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			//定时器的回调,发送心跳检查,
			err := c.write(websocket.PingMessage, []byte{})

			if err != nil {
				return //客户端不响应心跳就停止
			}

		}
	}
}

// write 输出数据
func (c *Client) write(mt int, message []byte) error {
	c.ws.SetWriteDeadline(time.Now().Add(WriteWait)) //设置发送的超时时间点
	return c.ws.WriteMessage(mt, message)
}

// ReadPump 监听消息
func (c *Client) readPump() {

	c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置客户端心跳响应的最后限期
	c.ws.SetPongHandler(func(string) error {
		c.ws.SetReadDeadline(time.Now().Add(PongWait)) //设置下一次心跳响应的最后限期
		return nil
	})
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		log.Printf("read: %s\n", string(message))

		packet := NewDataPacket(gjson.ParseBytes(message))

		//开一个goroutine处理消息
		go func(p DataPacket, client *Client) {

			errRun := client.checkNonceReplay(p)
			if errRun != nil {
				client.Send(responsErrorPacket(errRun.Error(), p.Method, errReplayAttack, p.Nonce))
				return
			}

			//TODO:验证签名

			//创建上下面指针，处理请求参数，及响应
			ctx := &Context{Params: p.Data, Method: p.Method}
			client.handler.ServeOWTP(c, ctx)

			//处理完后，推送响应结果给服务端
			p.Req = WSResponse

			client.Send(p)

		}(*packet, c)

	}
}

//verifySignature 钱包签名
func (c *Client) verifySignature(packet DataPacket) error {
	//TODO:
	return nil
}

//checkNonceReplay 检查nonce是否重放
func (c *Client) checkNonceReplay(packet DataPacket) error {

	if packet.Nonce == 0 || packet.Timestamp == 0 {
		//没有nonce直接跳过
		return errors.New("no nonce")
	}

	//检查是否重放
	db, err := storm.Open(c.cacheFile)
	if err != nil {
		return errors.New("client system cache error")
	}
	defer db.Close()

	var existPacket DataPacket
	db.One("Nonce", packet.Nonce, &existPacket)
	if &existPacket != nil {
		return errors.New("this is a replay attack")
	}

	return nil

}

//Close 关闭连接
func (c *Client) Close() {
	c.ws.Close()
	close(c.send)
}
