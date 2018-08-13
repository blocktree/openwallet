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
	"github.com/pkg/errors"
	"sync"
	"time"
	"fmt"
	"github.com/blocktree/OpenWallet/log"
)

// 路由处理方法
type HandlerFunc func(ctx *Context)

//请求方法，回调响应结果
type RequestFunc func(resp Response)

//请求队列
type RequestQueue map[uint64]requestEntry

//ServeMux 多路复用服务
type ServeMux struct {
	//读写锁
	mu sync.RWMutex
	//路由方法绑定
	m map[string]muxEntry
	//请求队列
	//requestQueue map[uint64]requestEntry
	//超时时间
	timeout time.Duration
	//是否启动了请求超时检查
	startRequestTimeoutCheck bool
	//节点的请求队列
	peerRequest map[string]RequestQueue
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
	time int64
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
func (mux *ServeMux) AddRequest(pid string, nonce uint64, time int64, method string, reqFunc RequestFunc, respChan chan Response, sync bool) error {

	mux.mu.Lock()
	defer mux.mu.Unlock()

	if !mux.startRequestTimeoutCheck {
		go mux.timeoutRequestHandle()
	}

	requestQueue := mux.peerRequest[pid]

	if requestQueue == nil {
		requestQueue = make(map[uint64]requestEntry)
		mux.peerRequest[pid] = requestQueue
	}

	if _, exist := requestQueue[nonce]; exist {
		return errors.New("OWTP: nonce exist. ")
	}

	requestQueue[nonce] = requestEntry{sync, method, reqFunc, respChan, time}
	return nil
}

//ResetQueue 重置请求队列
//func (mux *ServeMux) ResetQueue() {
//	mux.mu.Lock()
//	defer mux.mu.Unlock()
//
//	if mux.requestQueue == nil {
//		mux.requestQueue = make(map[uint64]requestEntry)
//	}
//
//	//处理所有未完成的请求，返回连接断开的异常
//	for n, r := range mux.requestQueue {
//		resp := responseError("network disconnected", ErrNetworkDisconnected)
//		if r.sync {
//			r.respChan <- resp
//		} else {
//			r.h(resp)
//		}
//		delete(mux.requestQueue, n)
//	}
//
//}


//ResetRequestQueue 重置请求队列
func (mux *ServeMux) ResetRequestQueue(pid string) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	requestQueue := mux.peerRequest[pid]

	if requestQueue == nil {
		return
	}

	//处理所有未完成的请求，返回连接断开的异常
	for n, r := range requestQueue {
		resp := responseError("network disconnected", ErrNetworkDisconnected)
		if r.sync {
			r.respChan <- resp
		} else {
			r.h(resp)
		}
		delete(requestQueue, n)
	}

	mux.peerRequest[pid] = requestQueue
}


//ServeOWTP OWTP协议消息监听方法
func (mux *ServeMux) ServeOWTP(pid string, ctx *Context) {

	switch ctx.Req {
	case WSRequest: //对方发送请求

		f, ok := mux.m[ctx.Method]
		if ok {
			f.h(ctx)
		} else {
			//找不到方法的处理
			ctx.Resp = responseError("can not find method", ErrNotFoundMethod)
		}
	case WSResponse: //我方请求后，对方响应返回
		mux.mu.Lock()
		defer mux.mu.Unlock()
		requestQueue := mux.peerRequest[pid]
		if requestQueue == nil {
			log.Error("peer:", pid, "requestQueue is nil.")
			return
		}
		f := requestQueue[ctx.nonce]
		if f.method == ctx.Method {
			//log.Printf("f: %v", f)
			if f.sync {
				f.respChan <- ctx.Resp
			} else {
				f.h(ctx.Resp)
			}
			delete(requestQueue, ctx.nonce)
		} else {
			//响应与请求的方法不一致
			//ctx.Resp = responseError("reponse method is not equal request", ErrResponseMethodDiffer)
			log.Error("reponse method is not equal request.")
		}

	default:
		//未知类型的日志记录
		log.Error("Unknown messages.")
	}
}

// timeoutRequestHandle 超时请求检查
func (mux *ServeMux) timeoutRequestHandle() {
	if mux.timeout == 0 {
		mux.timeout = 120
	}
	ticker := time.NewTicker(10 * time.Second) //发送心跳间隔事件要<等待时间
	defer func() {
		ticker.Stop()
	}()

	mux.startRequestTimeoutCheck = true

	for {
		select {
		case <-ticker.C:
			//log.Printf("check request timeout \n")
			mux.mu.Lock()
			//定时器的回调，处理超时请求
			for _, requestQueue := range mux.peerRequest {

				for n, r := range requestQueue {

					currentServerTime := time.Now()

					//计算客户端过期时间
					requestTimestamp := time.Unix(r.time, 0)
					expiredTime := requestTimestamp.Add(mux.timeout)

					//log.Printf("requestTimestamp = %s \n", requestTimestamp.String())
					//log.Printf("currentServerTime = %s \n", currentServerTime.String())
					//log.Printf("expiredTime = %s \n", expiredTime.String())

					if currentServerTime.Unix() > expiredTime.Unix() {
						//log.Printf("request expired time")
						//返回超时响应
						errInfo := fmt.Sprintf("request timeout over %s", mux.timeout.String())
						resp := responseError(errInfo, ErrRequestTimeout)
						if r.sync {
							//log.Error("resp =", resp)
							r.respChan <- resp
						} else {
							r.h(resp)
						}
						delete(requestQueue, n)
					}

				}

			}

			mux.mu.Unlock()
		}
	}
}

