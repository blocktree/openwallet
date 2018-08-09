## Walletnode
--------------

#### 管理 Fullnode 的启动，重启，关闭。

	import "github.com/blocktree/OpenWallet/walletnode"

	symbol := "bopo"
	wn := walletnode.NodeManagerStruct{}

	// 关闭
	if err := wn.StopNodeFlow(symbol); err != nil {
		log.Println(err)
	}

	// 开启
	if err := wn.StartNodeFlow(symbol); err != nil {
		log.Println(err)
	}

	// 重启
	if err := wn.RestartNodeFlow(symbol); err != nil {
		log.Println(err)
	}

--------------
