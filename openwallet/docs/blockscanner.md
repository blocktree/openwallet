# BlockScanner

区块扫描器接口，由资产适配包实现区块扫描方法，通知已openwallet钱包体系设计的数据到应用层。

## 区块链交易单模型

几乎所有的区块链交易单结构，或采用utxo模型，或采用account模型。（DAG有向无环图未研究）

### UTXO模型

这类模型的区块链不能直接查询到余额的，每笔交易单都以多个输入（vins）和多个输出（vouts）组成。
输入合计数量必须大于或等于输出合计数量。矿工费 = 总输入 - 总输出。
每次构建新交易单，总是依赖地址未使用过的输出，我们称为unspent transaction output（utxo）。
所以要知道地址的余额，就是累计地址所有的utxo中的数量。

大部分区块链采用这种模型的：bitcoin系，bytom，qtum等等。

### account模型

采用这类模型的，可以直接在区块链中查到地址的余额。与utxo模型不同是，交易单只能有一对输入输出。

采用这种模型的：ethereum，nas等等。

## 区块扫描器使用流程

- 获取某个适配的区块扫描器接口。
- 通过调用SetBlockScanAddressFunc(scanAddressFunc BlockScanAddressFunc)方法，设置区块扫描器，查找地址算法。
- 通过AddObserver(obj BlockScanNotificationObject)添加监听者。
- 调用者实现BlockScanAddressFunc方法，返回传入的地址是否存在，及地址关联的sourceKey值。
- BlockScanNotificationObject要实现接收数据的接口方法，并保存。
- 执行Run()启动扫描器运行。

## 区块扫描任务实现

- 区块任务会放在定时器中定时执行，定时间隔要少于出块时间，不然永远都扫不到链的最大高度。
- 扫描器以本地已扫高度为开始，先一直扫描到最大高度，再扫描内存池交易，最后扫描失败记录，已结束一次任务。
- 每完成一个区块扫描，通知观察者新的区块头数据，并使用本地数据库文件记录区块头数据。
- 每次任务都会查找是否有新区块可扫。
- 如果扫描到新区块时，发现就上一区块分叉，则回退扫描，并通知观察者有区块分叉，直到没有分叉的区块为止。

## 区块中的交易单提取实现

- 由于区块中的交易单数组是否互相独立的，为了提高效率，我们可以采用生产消费者并发模型，并行提取多张交易单的数据。
- openwallet钱包体系，以资产账户作为基本操作单位。一个账户可能有多个地址，所以存在多个地址指向相同的sourceKey。sourceKey作为索引收集一张交易单中与地址相关的提取结果。
- 提取结果最终以N个TxInput，M个TxOutPut，1个Transaction，通知给BlockScanNotificationObject观察者。

## BlockScanNotificationObject接口说明

实现BlockScanNotificationObject接口。

- BlockScanNotify获取新区块。
- BlockExtractDataNotify交易单提取结果

```go

    //BlockScanNotify 新区块扫描完成通知
	BlockScanNotify(header *BlockHeader) error

	//BlockExtractDataNotify 区块提取结果通知
	BlockExtractDataNotify(sourceKey string, data *TxExtractData) error

```

### TxExtractData

blockscanner会从每笔交易单中，根据订阅地址提取出来，并以地址绑定的sourceKey汇集在一个提取结果中。
sourceKey其实就是账户的唯一标识符。

- []*TxInput，utxo模型的可以有多个输入，account模型只有1个。
- []*TxOutPut，utxo模型的可以有多个输出，account模型只有1个。
- *Transaction，openwallet的钱包模型是以账户为单位的，账户存在多个地址可能，所以账户在该交易单的转账数量是以[]*TxOutPut的总数量 - []*TxInput的总数量计算出来。

### account模型类区块扫描注意事项

- 手续费是记录在交易单上，为了给上层应用计算完整，需要为手续费建1个TxInput。
- 智能合约类的交易单存在失败状态。失败状态下，只扣了手续费，而from和to就不需要建TxOutPut，TxInput。
- 扫描合约转账。要提取2张交易单，1张是代币的转账，1张是主链币消耗手续费单。








