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
	"errors"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"github.com/tidwall/gjson"
	"net/http"
	"time"
	"github.com/blocktree/OpenWallet/log"
)

//局部常量
const (
	WriteWait       = 60 * time.Second //超时为6秒
	PongWait        = 30 * time.Second
	PingPeriod      = (PongWait * 9) / 10
	WSRequest       = 1 //wesocket请求标识
	WSResponse      = 2 //wesocket响应标识
	MaxMessageSize  = 1 * 1024
)

var (
	Debug = false
)

type Handler interface {
	ServeOWTP(client *Client, ctx *Context)
}

//Authorization 授权
type Authorization interface {
	//ConnectAuth 连接授权，处理连接授权参数，返回完整的授权连接
	ConnectAuth(url string) string
	//GenerateSignature 生成签名，并把签名加入到DataPacket中
	AddAuth(data *DataPacket) bool
	//VerifySignature 校验签名，若验证错误，可更新错误信息到DataPacket中
	VerifyAuth(data *DataPacket) bool
	//EncryptData 加密数据
	EncryptData(data []byte) ([]byte, error)
	//DecryptData 解密数据
	DecryptData(data []byte) ([]byte, error)
	//EnableAuth 开启授权
	EnableAuth() bool

	//Header 协议头
	//Header() map[string]string
	////VerifyHeader 校验授权
	//VerifyHeader(header map[string]string) bool
}

//Client 基于websocket的通信客户端
type Client struct {
	auth    Authorization
	ws      *websocket.Conn
	handler Handler
	send    chan []byte
	close   chan struct{}
	isHost  bool
	ReadBufferSize  int
	WriteBufferSize int
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

type Param struct {
	rawValue interface{}
}

type Context struct {
	//传输类型，1：请求，2：响应
	Req uint64
	//请求序号
	nonce uint64
	//参数内部
	params gjson.Result
	//方法
	Method string
	//响应
	Resp Response
	//传入参数，map的结构
	inputs interface{}
}

//NewContext
func NewContext(req, nonce uint64, method string, inputs interface{}) *Context {
	ctx := Context{
		Req:    req,
		nonce:  nonce,
		Method: method,
		inputs: inputs,
	}

	return &ctx
}

//Params 获取参数
func (ctx *Context) Params() gjson.Result {
	//如果param没有值，使用inputs初始化
	if !ctx.params.Exists() {
		inbs, err := json.Marshal(ctx.inputs)
		if err == nil {
			ctx.params = gjson.ParseBytes(inbs)
		}
	}
	return ctx.params
}

func (ctx *Context) Response(result interface{}, status uint64, msg string) {

	resp := Response{
		Status: status,
		Msg:    msg,
		Result: result,
	}

	ctx.Resp = resp
}

//JsonData the result of Response encode gjson
func (resp *Response) JsonData() gjson.Result {
	var jsondata gjson.Result
	inbs, err := json.Marshal(resp.Result)
	if err == nil {
		jsondata = gjson.ParseBytes(inbs)
	}

	return jsondata
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
func Dial(url string, router Handler, auth Authorization) (*Client, error) {

	if router == nil {
		return nil, errors.New("router should not be nil!")
	}

	//处理连接授权
	authURL := url
	if auth != nil && auth.EnableAuth() {
		authURL = auth.ConnectAuth(url)
	}
	log.Debug("Connecting URL:", authURL)



	client := &Client{
		handler: router,
		send:    make(chan []byte, MaxMessageSize),
		auth:    auth,
		close:   make(chan struct{}),
		ReadBufferSize: 1024 * 1024,
		WriteBufferSize: 1024 * 1024,
	}

	dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		ReadBufferSize: client.ReadBufferSize,
		WriteBufferSize: client.WriteBufferSize,
	}

	c, _, err := dialer.Dial(authURL, nil)
	if err != nil {
		return nil, err
	}

	client.ws = c

	//发送通道
	go client.writePump()

	//监听消息
	go client.readPump()

	return client, nil
}

//SetCloseHandler 设置关闭连接时的回调
func (c *Client) SetCloseHandler(h func(code int, text string) error) {
	c.ws.SetCloseHandler(h)
}

//Send 发送消息
func (c *Client) Send(data DataPacket) error {

	//添加授权
	if c.auth != nil && c.auth.EnableAuth() {
		if !c.auth.AddAuth(&data) {
			return errors.New("OWTP: authorization failed")
		}
	}
	//log.Printf("Send: %v\n", data)
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	//log.Printf("Send: %s\n", string(respBytes))
	c.send <- respBytes
	return nil
}

// WritePump 发送消息通道
func (c *Client) writePump() {

	ticker := time.NewTicker(PingPeriod) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
		c.Close()
		log.Debug("writePump end")
	}()
	for {
		select {
		case message, ok := <-c.send:
			//发送消息
			if !ok {
				c.write(websocket.CloseMessage, []byte{})
				return
			}
			if Debug {
				log.Debug("Send: ", string(message))
			}
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
	defer func() {
		//c.Close()
		log.Debug("readPump end")
	}()

	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			log.Error("Read unexpected error: ", err)
			//close(c.send) //读取通道异常，关闭读通道
			break
		}
		if Debug {
			log.Debug("Read: ", string(message))
		}

		packet := NewDataPacket(gjson.ParseBytes(message))

		//开一个goroutine处理消息
		go func(p DataPacket, client *Client) {

			//验证授权
			if client.auth != nil && client.auth.EnableAuth() {
				if !c.auth.VerifyAuth(&p) {
					log.Error("auth failed: ", p)
					client.Send(p) //发送验证失败结果
					return
				}
			}

			if p.Req == WSRequest {

				//创建上下面指针，处理请求参数
				ctx := Context{Req: p.Req, nonce: p.Nonce, inputs: p.Data, Method: p.Method}

				client.handler.ServeOWTP(client, &ctx)

				//处理完请求，推送响应结果给服务端
				p.Req = WSResponse
				p.Data = ctx.Resp
				client.Send(p)
			} else if p.Req == WSResponse {

				//创建上下面指针，处理响应
				var resp Response
				runErr := mapstructure.Decode(p.Data, &resp)
				if runErr != nil {
					log.Error("Response decode error: ", runErr)
					return
				}

				ctx := Context{Req: p.Req, nonce: p.Nonce, inputs: nil, Method: p.Method, Resp: resp}

				client.handler.ServeOWTP(client, &ctx)

			}
		}(*packet, c)

	}
}

//Close 关闭连接
func (c *Client) Close() {
	log.Debug("client close\n")
	c.ws.Close()
	//close(c.send)
	c.close <- struct{}{}
}
