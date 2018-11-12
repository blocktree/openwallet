# AssetsAdapter

AssetsAdapter是区块链资产适配接口。assets-adapter/assets包下各个区块链对接口进行实现。
我们的开发经验总结出，适配区块链资产需要支持以下，则实现以下接口：

- 币种信息。关键信息有"标识符"symbol（框架内唯一标识），"小数位精度"Decimal，"ECC曲线类型"CurveType。
- 资产配置。每个资产适配器的配置存在差异，适配器实现加载配置，外部程序传入配置接口。
- 地址解析器。不同区块链存在地址编码不同的情况，资产适配器提供地址解析器工具，给应用去处理地址解析问题。
- 交易单解析器。提供标准的交易流程：创建交易单，签名交易单，验证交易单，广播交易单。资产适配器各自实现其交易协议。
- 区块链扫描器。提供标准的区块扫描器，调用设置观测者，可接受区块扫描提取结果。资产适配器各自实现区块提取算法。
- 智能合约解析器。用于区块链智能合约方面的扩展，目前支持获取Token余额。

```go

// AssetsAdapter 资产适配器接口
// 适配OpenWallet钱包体系的抽象接口
type AssetsAdapter interface {

	//币种信息
	SymbolInfo

	//配置
	AssetsConfig

	//GetAddressDecode 地址解析器
	GetAddressDecode() AddressDecoder

	//GetTransactionDecoder 交易单解析器
	GetTransactionDecoder() TransactionDecoder

	//GetBlockScanner 获取区块链扫描器
	GetBlockScanner() BlockScanner

	//GetSmartContractDecoder 获取智能合约解析器
	GetSmartContractDecoder() SmartContractDecoder
}


```