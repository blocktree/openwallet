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

// owtp全称OpenWallet Transfer Protocol，OpenWallet的一种点对点的分布式私有通信协议。
package owtp

import (
	"github.com/blocktree/OpenWallet/logger"
	"github.com/pkg/errors"
	"sync"
	"time"
)

// 路由处理方法
type HandlerFunc func(ctx *Context)

//请求方法，回调响应结果
type RequestFunc func(resp Response)

//ServeMux 多路复用服务
type ServeMux struct {
	//读写锁
	mu sync.RWMutex
	//路由方法绑定
	m map[string]muxEntry
	//请求队列
	requestQueue map[uint64]requestEntry
}

type muxEntry struct {
	h      HandlerFunc
	method string
}

type requestEntry struct {
	sync     bool
	method   string
	h        RequestFunc
	respChan chan Response
}

//HandleFunc 路由处理器绑定
//@param method API方法名
//@param handler 处理方法入口
func (mux *ServeMux) HandleFunc(method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if method == "" {
		panic("OWTP: invalid pattern")
	}
	if handler == nil {
		panic("OWTP: nil handler")
	}
	if _, exist := mux.m[method]; exist {
		panic("OWTP: multiple registrations for " + method)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
	mux.m[method] = muxEntry{h: handler, method: method}

}

//AddRequest 添加请求到队列
//@param nonce 递增不可重复
//@param method API方法名
//@param reqFunc 异步请求的回调函数
//@param respChan 同步请求的响应通道
//@param sync 是否同步
func (mux *ServeMux) AddRequest(nonce uint64, method string, reqFunc RequestFunc, respChan chan Response, sync bool) error {

	mux.mu.Lock()
	defer mux.mu.Unlock()

	if _, exist := mux.requestQueue[nonce]; exist {
		return errors.New("OWTP: nonce exist. ")
	}

	mux.requestQueue[nonce] = requestEntry{sync, method, reqFunc, respChan}
	return nil
}

//ResetQueue 重置请求队列
func (mux *ServeMux) ResetQueue() {
	mux.requestQueue = make(map[uint64]requestEntry)
}

//ServeOWTP OWTP协议消息监听方法
func (mux *ServeMux) ServeOWTP(client *Client, ctx *Context) {

	switch ctx.Req {
	case WSRequest: //对方发送请求

		f, ok := mux.m[ctx.Method]
		if ok {
			f.h(ctx)
		} else {
			//找不到方法的处理
			client.Send(responsErrorPacket("can not find method", ctx.Method, errNotFoundMethod, ctx.nonce))
		}
	case WSResponse: //我方请求后，对方响应返回
		f := mux.requestQueue[ctx.nonce]
		if f.method == ctx.Method {

			if f.sync {
				f.respChan <- ctx.Resp
			} else {
				f.h(ctx.Resp)
			}

		} else {
			//响应与请求的方法不一致
			client.Send(responsErrorPacket("reponse method is not equal request", ctx.Method, errResponseMethodDiffer, ctx.nonce))
		}

	default:
		//未知类型的日志记录
		openwLogger.Log.Info("Unknown messages.")
	}
}

//responsErrorPacket 返回一个错误数据包
func responsErrorPacket(err, method string, status, nonce uint64) DataPacket {

	resp := Response{
		Status: status,
		Msg:    err,
		Result: nil,
	}

	//封装数据包
	packet := DataPacket{
		Method:    method,
		Req:       WSResponse,
		Nonce:     nonce,
		Timestamp: time.Now().Unix(),
		Data:      resp,
	}

	return packet
}
