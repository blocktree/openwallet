# AssetsAdapter

AssetsAdapter是区块链资产适配接口。assets-adapter/assets包下各个区块链对接口进行实现。
我们的开发经验总结出，适配区块链资产需要支持以下，则实现以下接口：

- 币种信息。关键信息有"标识符"symbol（框架内唯一标识），"小数位精度"Decimal，"ECC曲线类型"CurveType。
- 资产配置。每个资产适配器的配置存在差异，适配器实现加载配置，外部程序传入配置接口。
- 地址解析器。不同区块链存在地址编码不同的情况，资产适配器提供地址解析器工具，给应用去处理地址解析问题。
- 交易单解析器。
- 区块链扫描器。


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