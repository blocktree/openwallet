# OWTP Protocal

OWTP协议全称OpenWallet Transfer Protocol，是一种基于点对点的分布式通信协议。
OWTP协议能够让开发者专注于解决应用业务实现，无需关心底层的网络连接实现。
通过简单的配置就能够让两端建立不同的网络连接方式，路由可复用，业务无需重写。

## 框架特点

- 支持多种网络连接协议：http，websocket，mq等。
- 支持多种网络传输数据格式：目前只有JSON，未来支持Protobuf。
- 内置SM2协商密码机制，无需https，也可实现加密通信。
- 内置数字签名，防重放，防中途篡改数据。
- 支持多种session缓存方案。
- 多种网络连接协议复用统一的路由配置。
 
## 如何使用


### 创建节点

```go

    //1. 使用配置文件创建
    cert, err := NewCertificate(RandomPrivateKey())
    if err != nil {
        return nil
    }
    
    config := NodeConfig{
        Cert: cert,    //配置节点证书
        TimeoutSEC: 60 //请求超时时间
    }
    
    host := NewNode(config)
    
    //2. 随机创建一个带证书的节点
    host := RandomOWTPNode()
	
```

### 可选配置Session

```go

    //创建一个全局的会话管理器，详细资料参考sesssion包的README.md
    globalSessions, _ = owtp.NewSessionManager("memory", &session.ManagerConfig{
		Gclifetime: 10,
	})
	go globalSessions.GC()
	
	//设置节点Peerstore指向一个全局的会话管理
	host.SetPeerstore(globalSessions)

```

### 节点作为服务端使用

```go

    //外置的业务方法
     func getInfo (ctx *Context) {
        //记录当前连接的信息到session，用于上下文操作
        ctx.SetSession("username", "kkk")
    
        ctx.Resp = Response{
            Status: 0,
            Msg:    "success",
            Result: map[string]interface{}{
                "getInfo": "hi boy",
            },
        }
    }
```

```go
    
    
    //配置路由的业务方法
    host.HandleFunc("getInfo", getInfo)

    //配置处理业务前的准备过程【可选】
	host.HandlePrepareFunc(func(ctx *Context) {
		
	})

    //配置处理业务后的结束过程【可选】
	host.HandleFinishFunc(func(ctx *Context) {
		
		//断开节点连接，长连接才响应，HTTP为短连接，不会响应
        host.ClosePeer(peer.ID)
	})

	//开启端口监听HTTP连接请求
	host.Listen(
		ConnectConfig{
			Address:     ":9432",
			ConnectType: HTTP,
			EnableSignature: true, //开启数字签名
		})

	//开启端口监听websocket连接请求
	host.Listen(
		ConnectConfig{
			Address:     ":9433",
			ConnectType: Websocket,
		})
    
    //更多复杂的连接配置可查看ConnectConfig类

    //监听长连接打开，处理后续业务（HTTP为短连接，不支持）
	host.SetOpenHandler(func(n *OWTPNode, peer PeerInfo) {
		log.Infof("peer[%s] connected", peer.ID)
		log.Infof("peer[%+v] config", peer.Config)
	})

    //监听长连接断开，处理后续业务（HTTP为短连接，不支持）
	wsHost.SetCloseHandler(func(n *OWTPNode, peer PeerInfo) {
		
	})

```

### 节点作为客户端使用


```go

    //随机创建带证书的客户端
    client := RandomOWTPNode()
    
    //配置路由的业务方法
    client.HandleFunc("getInfo", getInfo)
    
    //通过HTTP连接服务端
    err := client.Connect("testhost", ConnectConfig{
        Address:     ":9432",
        ConnectType: HTTP,
        EnableSignature: true, //开启数字签名
    })
    
    /*
    //或通过Websocket连接服务端
    err := client.Connect("testhost", ConnectConfig{
            Address:     ":9433",
            ConnectType: Websocket,
        })
    */

    if err != nil {
        return
    }

    //向已连接的testhost主机，开启协商密码，加密方式AES
    err = client.KeyAgreement("testhost", "aes")
    if err != nil {
        return
    }

    params := map[string]interface{}{
        "name": "chance",
        "age":  18,
    }

    //向已连接的testhost主机，发起业务请求
    //参数1：主机ID，参数2：路由的方法名，参数3：业务参数，参数4：是否同步线程，参数5：响应结果处理
    //参数4 sync = true，程序会等待响应结果处理完，才走程序下一步处理。
    err = client.Call("testhost", "getInfo", params, true, func(resp Response) {

        result := resp.JsonData()
        symbols := result.Get("getInfo")
        fmt.Printf("getInfo: %v\n", symbols)
    })

    if err != nil {
        return
    }
	
```