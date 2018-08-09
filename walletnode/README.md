# Walletnode


## Fullnode 管理接口

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

## 远程备份，恢复文件接口(上传，下载)

开发中，可以先使用本地复制，待完成后升级为“远程复制”方式即可（生产环境需要）。

// 从钱包备份
```
symbol := "bopo"
src := filepath.Join(walletDataPath, "wallet.dat")  // 源文件，指向全节点中的路径，需全路径和文件名
dst := filepath.Join(newBackupDir, "wallet.dat")	// 目标文件，指向本地路径，也需指定文件名（支持相对路径）

// 开始备份
wn := walletnode.NodeManagerStruct{}
if err := wn.CopyFromContainer(symbol, src, dst); err != nil {
	log.Println(err)
}
```

// 恢复到钱包
```
symbol := "bopo"
src := datFile			// 本地文件，可用相对路径
dst := walletDataPath		// 远程fullnode目录，无需指定文件名，只需要声明恢复到远程钱包的那个目录即可

// 开始恢复
wn := walletnode.NodeManagerStruct{}
if err := wn.CopyToContainer(symbol, src, dst); err != nil {
	log.Println(err)
}


```