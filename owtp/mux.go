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
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/cache"
	"github.com/blocktree/OpenWallet/log"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"sync"
	"time"
)

const (
	WSRequest  = 1 //请求标识
	WSResponse = 2 //响应标识
)

var (
	//重放限制时长，数据包的时间戳超过后，这个间隔，可以重复nonce
	replayLimit = 2 * time.Hour
)

// 路由处理方法
type HandlerFunc func(ctx *Context)

//请求方法，回调响应结果
type RequestFunc func(resp Response)

//请求队列
type RequestQueue map[uint64]requestEntry

type muxEntry struct {
	h      HandlerFunc
	method string
}

type requestEntry struct {
	sync     bool
	method   string
	h        RequestFunc
	respChan chan Response
	time     int64
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
	//节点ID
	PID string
	//传输类型，1：请求，2：响应
	Req uint64
	//请求的远程IP
	RemoteAddress string
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
	//节点会话
	peerstore Peerstore
	//是否中断，Context.stop = true，将不再执行后面的绑定的业务
	stop bool
}

//NewContext
func NewContext(req, nonce uint64, pid, method string, inputs interface{}) *Context {
	ctx := Context{
		Req:    req,
		nonce:  nonce,
		PID:    pid,
		Method: method,
		inputs: inputs,
	}

	return &ctx
}

//Params 获取参数
func (ctx *Context) Params() gjson.Result {
	//如果param没有值，使用inputs初始化
	if !ctx.params.Exists() {
		//inbs, err := json.Marshal(ctx.inputs)
		inbs, ok := ctx.inputs.([]byte)
		if ok {
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

// ResponseStopRun 中断操作，Context.stop = true，将不再执行后面的绑定的业务
// 并完成Response处理
func (ctx *Context) ResponseStopRun(result interface{}, status uint64, msg string) {
	ctx.stop = true
	ctx.Response(result, status, msg)
}

// SetSession puts value into session.
func (ctx *Context) SetSession(name string, value interface{}) {
	ctx.peerstore.Put(ctx.PID, name, value)
}

// GetSession gets value from session.
func (ctx *Context) GetSession(name string) interface{} {
	return ctx.peerstore.Get(ctx.PID, name)
}

// DelSession removes value from session.
func (ctx *Context) DelSession(name string) {
	ctx.peerstore.Delete(ctx.PID, name)
}

// DestroySession cleans session data
func (ctx *Context) DestroySession() {
	ctx.peerstore.Destroy(ctx.PID)
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

//ServeMux 多路复用服务
type ServeMux struct {
	//读写锁
	mu sync.RWMutex
	//路由方法绑定
	m map[string]muxEntry
	//超时时间
	timeout time.Duration
	//是否启动了请求超时检查
	startRequestTimeoutCheck bool
	//节点的请求队列
	peerRequest map[string]RequestQueue
	//请求缓存
	peerRequestCache cache.Cache
	//请求nonce的市场限制
	requestNonceLimit time.Duration
}

func NewServeMux(timeoutSEC int) *ServeMux {
	//6小时清理一次内存中的请求nonce
	cache, err := cache.NewCache("memory", `{"interval":21600}`)
	if err != nil {
		log.Error("NewServeMux unexpected err:", err)
	}

	serveMux := ServeMux{
		timeout:           time.Duration(timeoutSEC) * time.Second,
		peerRequest:       make(map[string]RequestQueue),
		m:                 make(map[string]muxEntry),
		peerRequestCache:  cache,
		requestNonceLimit: replayLimit,
	}
	return &serveMux
}

//HandleFunc 路由处理器绑定
//@param method API方法名
//@param handler 处理方法入口
func (mux *ServeMux) HandleFunc(method string, handler HandlerFunc) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if method == "" {
		log.Error("OWTP: invalid pattern")
	}
	if handler == nil {
		log.Error("OWTP: nil handler")
	}
	if _, exist := mux.m[method]; exist {
		log.Error("OWTP: multiple registrations for " + method)
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

	mux.peerRequest[pid] = requestQueue

	return nil
}

//RemoveRequest 移除请求
func (mux *ServeMux) RemoveRequest(pid string, nonce uint64) error {

	requestQueue := mux.peerRequest[pid]

	if requestQueue == nil {
		return nil
	}

	mux.mu.Lock()
	defer mux.mu.Unlock()

	delete(requestQueue, nonce)
	mux.peerRequest[pid] = requestQueue

	return nil
}

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

		//重复攻击检查
		if !mux.checkNonceReplay(ctx) {
			log.Error("nonce duplicate: ", ctx)
		} else {
			f, ok := mux.m[ctx.Method]
			if ok {
				//执行准备处理方法
				if prepareFunc, exist := mux.m[PrepareMethod]; exist {
					prepareFunc.h(ctx)
				}

				if !ctx.stop {
					//执行路由方法
					f.h(ctx)
				}

				if !ctx.stop {
					//执行结束处理方法
					if finishFunc, exist := mux.m[FinishMethod]; exist {
						finishFunc.h(ctx)
					}
				}

				//添加已完成的请求
				if mux.peerRequestCache != nil {
					mux.peerRequestCache.Put(
						fmt.Sprintf("%s_%d", pid, ctx.nonce),
						ctx.Method,
						mux.requestNonceLimit)
				}

			} else {
				//找不到方法的处理
				ctx.Resp = responseError("can not find method", ErrNotFoundMethod)
			}
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

//checkNonceReplay 检查nonce是否重放
func (mux *ServeMux) checkNonceReplay(ctx *Context) bool {

	//检查
	status, errMsg := mux.checkNonceReplayReason(ctx.PID, ctx.nonce)

	if status != StatusSuccess {
		resp := Response{
			Status: status,
			Msg:    errMsg,
			Result: nil,
		}
		ctx.Resp = resp
		return false
	}

	return true

}

//checkNonceReplayReason 检查是否重放攻击
func (mux *ServeMux) checkNonceReplayReason(pid string, nonce uint64) (uint64, string) {

	if nonce == 0 {
		//没有nonce直接跳过
		return ErrReplayAttack, "no nonce"
	}

	//检查是否重放
	if mux.peerRequestCache.IsExist(fmt.Sprintf("%s_%d", pid, nonce)) {
		return ErrReplayAttack, "this is a replay attack"
	}

	return StatusSuccess, ""
}
