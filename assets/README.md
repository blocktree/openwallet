# assets 区块链资产适配指引

assets包下的各个子模块是实现区块链资产适配器。

## 资产适配器使用说明

```go

// 注册适配器
assets.RegAssets(bitcoin.Symbol, bitcoin.NewWalletManager())

// 注册适配器，并加载配置文件
absFile := filepath.Join(configFilePath, symbol+".ini")
c, _ := config.NewConfig("ini", absFile)
assets.RegAssets(bitcoin.Symbol, bitcoin.NewWalletManager(), c)

// 获取资产适配器对象
adapter := assets.GetAssets(symbol)
if adapter == nil {
    return nil, fmt.Errorf("assets: %s is not support", symbol)
}

```