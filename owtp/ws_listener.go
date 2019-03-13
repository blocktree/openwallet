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

package owtp

import (
	"context"
	"fmt"
	"github.com/blocktree/openwallet/log"
	ws "github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"net"
	"net/http"
)

//Listener 监听接口定义
type Listener interface {
	Accept() (Peer, error)
	Close() error
	Addr() net.Addr
}

// Default gorilla upgrader
var upgrader = ws.Upgrader{}

//owtp监听器
type wsListener struct {
	net.Listener
	handler         PeerHandler
	closed          chan struct{}
	incoming        chan Peer
	laddr           string
	enableSignature bool
	cert            Certificate
}

//serve 监听服务
func (l *wsListener) serve() error {

	if l.Listener == nil {
		return errors.New("listener is not setup.")
	}

	defer close(l.closed)
	//http.ListenAndServe(l.laddr, l)
	http.Serve(l.Listener, l)

	return nil
}

//ServeHTTP 实现HTTP服务监听
func (l *wsListener) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//创建一个上下文通知，监控节点是否已经关闭
	ctx, cancel := context.WithCancel(context.Background())
	httpCtx := r.Context()

	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade websocket", 400)
		return
	}

	peer, err := NewWSClientWithHeader(r.Header, l.cert, c, l.handler, l.enableSignature, cancel)
	if err != nil {
		log.Error("NewWSClientWithHeader unexpected error:", err)
		http.Error(w, "authorization not passed", 401)
		return
	}
	// Just to make sure.
	//defer peer.Close()
	select {
	case l.incoming <- peer:
	case <-l.closed:
		//peer.Close()
		return
	//case <-cnCh:
	case <-httpCtx.Done():
		log.Error("http CloseNotify")
		return
	}

	// wait until conn gets closed, otherwise the handler closes it early
	select {
	case <-ctx.Done(): //收到节点关闭的通知
		//log.Debug("peer 1:", peer.PID(), "closed")
		return
	case <-l.closed:
		//log.Debug("peer 2:", peer.PID(), "closed")
		peer.close()
	//case <-cnCh:
	case <-httpCtx.Done():
		log.Error("http CloseNotify")
		return
	}

}

//Accept 接收新节点链接，线程阻塞
func (l *wsListener) Accept() (Peer, error) {
	select {
	case c, ok := <-l.incoming:
		if !ok {
			return nil, fmt.Errorf("listener is closed")
		}
		return c, nil
	case <-l.closed:
		return nil, fmt.Errorf("listener is closed")
	}
}

//WSListenAddr 创建websocket通信监听
func WSListenAddr(addr string, cert Certificate, enableSignature bool, handler PeerHandler) (*wsListener, error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	listener := wsListener{
		Listener:        l,
		laddr:           addr,
		cert:            cert,
		handler:         handler,
		incoming:        make(chan Peer),
		closed:          make(chan struct{}),
		enableSignature: enableSignature,
	}

	go listener.serve()

	return &listener, nil
}
