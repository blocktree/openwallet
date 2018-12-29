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
	"fmt"
	"github.com/blocktree/OpenWallet/log"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

//owtp监听器
type httpListener struct {
	net.Listener
	handler         PeerHandler
	laddr           string
	peerstore       Peerstore //节点存储器
	enableSignature bool
}

//serve 监听服务
func (l *httpListener) serve() error {

	if l.Listener == nil {
		return errors.New("listener is not setup.")
	}

	http.Serve(l.Listener, l)

	return nil
}

//ServeHTTP 实现HTTP服务监听
func (l *httpListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//建立节点
	peer, err := NewHTTPClientWithHeader(w, r, l.handler, l.enableSignature)
	if err != nil {
		log.Error("NewClient unexpected error:", err)
		http.Error(w, "authorization not passed", 400)
		return
	}

	//HTTP是短连接，接收到数据，节点马上处理，无需像websocket那样管理连接
	err = peer.HandleRequest()
	if err != nil {
		log.Error("HandleRequest unexpected error:", err)
		http.Error(w, err.Error(), 400)
		return
	}
}

//Accept 接收新节点链接，线程阻塞
func (l *httpListener) Accept() (Peer, error) {
	return nil, fmt.Errorf("http do not implement")
}

//ListenAddr 创建OWTP协议通信监听
func HttpListenAddr(addr string, enableSignature bool, handler PeerHandler) (*httpListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener := httpListener{
		Listener:        l,
		laddr:           addr,
		handler:         handler,
		enableSignature: enableSignature,
	}

	go listener.serve()

	return &listener, nil
}
