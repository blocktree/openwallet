# OWTPNode

## 如何使用

### 创建服务端节点

```go

    //创建一个全局的会话管理器，详细资料参考sesssion包的README.md
    globalSessions, _ = owtp.NewSessionManager("memory", &session.ManagerConfig{
		Gclifetime: 10,
	})
	go globalSessions.GC()
    
    //节点配置，address：监听或访问的地址，connectType：监听链接方式，enableSignature：是否开启授权签名。
    config := make(map[string]string)
	config["address"] = httpURL
	config["connectType"] = owtp.HTTP
	config["enableSignature"] = "1"
	
    //随机生成私钥，创建证书
    cert, err := owtp.NewCertificate(owtp.RandomPrivateKey(), "")
	if err != nil {
		return nil
	}
    
    //创建服务器主机节点
	httpHost = owtp.NewOWTPNode(cert, 0, 0)
	
	//设置节点Peerstore指向一个全局的会话管理
	httpHost.SetPeerstore(globalSessions)
	
	//绑定方法
	httpHost.HandleFunc("getInfo", func (ctx *owtp.Context) {
		//记录数据到session会话
        ctx.SetSession("username", "john")
        ctx.Resp = owtp.Response{
            Status: 0,
            Msg:    "success",
            Result: map[string]interface{}{
                "username": "john",
            },
        }
    })
	
	//绑定准备处理方法
	httpHost.HandlePrepareFunc(func(ctx *owtp.Context) {
		log.Notice("prepare")
		//如果该需要提前结束请求，调用ctx.ResponseStopRun
		//ctx.ResponseStopRun(nil, StatusSuccess, "success")
	})
	
	//绑定结束处理方法
	httpHost.HandleFinishFunc(func(ctx *owtp.Context) {
		//从session会话中读取数据
		username := ctx.GetSession("username")
		log.Notice("username:", username)
		log.Notice("finish")
	})
	
	//启动监听
	httpHost.Listen(config)

```

### 创建客户端节点


```go

    //节点配置，address：访问服务端地址，connectType：监听链接方式，enableSignature：是否开启授权签名。
    config := make(map[string]string)
	config["address"] = httpURL
	config["connectType"] = owtp.HTTP
	config["enableSignature"] = "1"
	
	//随机生成一个带证书的节点
	httpClient := owtp.RandomOWTPNode()
	err := httpClient.Connect(httpHostNodeID, config)
	if err != nil {
		t.Errorf("Connect unexcepted error: %v", err)
		return
	}
	
	//开启协商密码，在通信过程中数据包D字段部分将用密码加密
	//err = httpClient.KeyAgreement(httpHostNodeID, "aes")
	//if err != nil {
	//	t.Errorf("KeyAgreement unexcepted error: %v", err)
	//	return
	//}

	params := map[string]interface{}{
		"name": "chance",
		"age":  18,
	}

    //向服务端发起请求
	err = httpClient.Call(httpHostNodeID, "getInfo", params, false, func(resp owtp.Response) {

		result := resp.JsonData()
		symbols := result.Get("symbols")

		fmt.Printf("symbols: %v\n", symbols)
	})

	if err != nil {
		t.Errorf("unexcepted error: %v", err)
		return
	}
	
```