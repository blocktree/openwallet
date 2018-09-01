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
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
	"net"
	"sync"
	"time"
	"github.com/streadway/amqp"
)


//MQClient 基于mq的通信客户端
type MQClient struct {
	auth            Authorization
	conn            *amqp.Connection
	channel         *amqp.Channel
	handler         PeerHandler
	send            chan []byte
	isHost          bool
	ReadBufferSize  int
	WriteBufferSize int
	pid             string
	isConnect       bool
	mu              sync.RWMutex //读写锁
	closeOnce       sync.Once
	done            func()
	config          map[string]string //节点配置
}

// Dial connects a client to the given URL.
func MQDial(pid, url string, handler PeerHandler) (*MQClient, error) {

	if handler == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	//处理连接授权
	//authURL := url
	//if auth != nil && auth.EnableAuth() {
	//	authURL = auth.ConnectAuth(url)
	//}
	log.Debug("Connecting URL:", url)

	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	channel, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	client, err := NewMQClient(pid, conn, channel, handler, nil, nil)
	if err != nil {
		return nil, err
	}

	client.isConnect = true
	client.isHost = true //我方主动连接
	client.handler.OnPeerOpen(client)

	return client, nil
}

func NewMQClient(pid string, conn *amqp.Connection, channel *amqp.Channel, hander PeerHandler, auth Authorization, done func()) (*MQClient, error) {

	if hander == nil {
		return nil, errors.New("hander should not be nil! ")
	}

	client := &MQClient{
		pid:  pid,
		conn: conn,
		channel:channel,
		send: make(chan []byte, MaxMessageSize),
		auth: auth,
		done: done,
	}

	client.isConnect = true
	client.SetHandler(hander)

	return client, nil
}

func (c *MQClient) PID() string {
	return c.pid
}

func (c *MQClient) Auth() Authorization {

	return c.auth
}

func (c *MQClient) SetHandler(handler PeerHandler) error {
	c.handler = handler
	return nil
}

func (c *MQClient) IsHost() bool {
	return c.isHost
}

func (c *MQClient) IsConnected() bool {
	return c.isConnect
}

func (c *MQClient) GetConfig() map[string]string {
	return c.config
}


//Close 关闭连接
func (c *MQClient) Close() error {
	var err error

	//保证节点只关闭一次
	c.closeOnce.Do(func() {

		if !c.isConnect {
			//log.Debug("end close")
			return
		}

		//调用关闭函数通知上级
		if c.done != nil {
			c.done()
			// Be nice to GC
			c.done = nil
		}

		err = c.conn.Close()
		c.isConnect = false
		c.handler.OnPeerClose(c, "client close")
	})
	return err
}

//LocalAddr 本地节点地址
func (c *MQClient) LocalAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	return c.conn.LocalAddr()
}

//RemoteAddr 远程节点地址
func (c *MQClient) RemoteAddr() net.Addr {
	if c.conn == nil {
		return nil
	}
	addr := &MqAddr{
		NetWork:c.config["address"],
	}
	return addr
}

//Send 发送消息
func (c *MQClient) Send(data DataPacket) error {

	////添加授权
	//if c.auth != nil && c.auth.EnableAuth() {
	//	if !c.auth.GenerateSignature(&data) {
	//		return errors.New("OWTP: authorization failed")
	//	}
	//}
	//log.Emergency("Send DataPacket:", data)
	respBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}

	if c.auth != nil && c.auth.EnableAuth() {
		respBytes, err = c.auth.EncryptData(respBytes)
		if err != nil {
			return errors.New("OWTP: EncryptData failed")
		}
	}

	//log.Printf("Send: %s\n", string(respBytes))
	c.send <- respBytes
	return nil
}

//OpenPipe 打开通道
func (c *MQClient) OpenPipe() error {

	if !c.IsConnected() {
		return fmt.Errorf("client is not connect")
	}

	//发送通道
	go c.writePump()

	//监听消息
	go c.readPump()

	return nil
}

// WritePump 发送消息通道
func (c *MQClient) writePump() {

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
func (c *MQClient) write(mt int, message []byte) error {
	if c.channel == nil {
		return new(amqp.Error)
	}
	exchange := c.config["exchange"]
	queueName := c.config["queueName"]
	fmt.Println("queueName:",queueName,",exchange",exchange)
	err := c.channel.Publish(exchange, queueName, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(message),
	})
	return err
}

// ReadPump 监听消息
func (c *MQClient) readPump() {
	if c.channel == nil {
		return
	}
	queueName := c.config["receiveQueueName"]

	fmt.Println("queueName:",queueName)
	msgs, err := c.channel.Consume(queueName, "", true, false, false, false, nil)

	if err!=nil{

	}

	forever := make(chan bool)

	go func() {
		//fmt.Println(*msgs)
		for d := range msgs {
			packet := NewDataPacket(gjson.ParseBytes(d.Body))

			//开一个goroutine处理消息
			go c.handler.OnPeerNewDataPacketReceived(c, packet)
		}
	}()
	<-forever

}
