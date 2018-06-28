/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */

//RPC client for ontology
package rpc

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	sdkcom "github.com/ontio/ontology-go-sdk/common"
	"github.com/ontio/ontology-go-sdk/utils"
	"github.com/ontio/ontology/account"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology/core/payload"
	"github.com/ontio/ontology/core/types"
	"github.com/ontio/ontology/smartcontract/service/native/ont"
	nutils "github.com/ontio/ontology/smartcontract/service/native/utils"
	cstates "github.com/ontio/ontology/smartcontract/states"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"sync/atomic"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//RpcClient for ontology rpc api
type RpcClient struct {
	qid        uint64
	addr       string
	httpClient *http.Client
}

//NewRpcClient return RpcClient instance
func NewRpcClient() *RpcClient {
	return &RpcClient{
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConnsPerHost:   5,
				DisableKeepAlives:     false, //enable keepalive
				IdleConnTimeout:       time.Second * 300,
				ResponseHeaderTimeout: time.Second * 300,
			},
			Timeout: time.Second * 300, //timeout for http response
		},
	}
}

//SetAddress set rpc server address. Simple http://localhost:20336
func (this *RpcClient) SetAddress(addr string) *RpcClient {
	this.addr = addr
	return this
}

//SetHttpClient set http client to RpcClient. In most cases SetHttpClient is not necessary
func (this *RpcClient) SetHttpClient(httpClient *http.Client) *RpcClient {
	this.httpClient = httpClient
	return this
}

//GetVersion return the version of ontology
func (this *RpcClient) GetVersion() (string, error) {
	data, err := this.sendRpcRequest(RPC_GET_VERSION, []interface{}{})
	if err != nil {
		return "", fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetVersion(data)
}

//GetBlockByHash return block with specified block hash
func (this *RpcClient) GetBlockByHash(hash common.Uint256) (*types.Block, error) {
	return this.GetBlockByHashWithHexString(hash.ToHexString())
}

//GetBlockByHash return block with specified block hash in hex string code
func (this *RpcClient) GetBlockByHashWithHexString(hash string) (*types.Block, error) {
	data, err := this.sendRpcRequest(RPC_GET_BLOCK, []interface{}{hash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetBlock(data)
}

//GetBlockByHeight return block by specified block height
func (this *RpcClient) GetBlockByHeight(height uint32) (*types.Block, error) {
	data, err := this.sendRpcRequest(RPC_GET_BLOCK, []interface{}{height})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetBlock(data)
}

//GetBlockCount return the total block count of ontology
func (this *RpcClient) GetBlockCount() (uint32, error) {
	data, err := this.sendRpcRequest(RPC_GET_BLOCK_COUNT, []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetUint32(data)
}

//GetCurrentBlockHash return the current block hash of ontology
func (this *RpcClient) GetCurrentBlockHash() (common.Uint256, error) {
	data, err := this.sendRpcRequest(RPC_GET_CURRENT_BLOCK_HASH, []interface{}{})
	if err != nil {
		return common.Uint256{}, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetUint256(data)
}

//GetBlockHash return block hash by block height
func (this *RpcClient) GetBlockHash(height uint32) (common.Uint256, error) {
	data, err := this.sendRpcRequest(RPC_GET_BLOCK_HASH, []interface{}{height})
	if err != nil {
		return common.Uint256{}, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetUint256(data)
}

//GetBalance return ont and ong balance of a ontology account
func (this *RpcClient) GetBalance(addr common.Address) (*sdkcom.Balance, error) {
	ontBalance, err := this.PrepareInvokeNativeContractWithRes(
		nutils.OntContractAddress,
		sdkcom.VERSION_CONTRACT_ONT,
		ont.BALANCEOF_NAME,
		[]interface{}{addr[:]},
		sdkcom.NEOVM_TYPE_INTEGER)
	if err != nil {
		return nil, fmt.Errorf("Get ONT balance of error:%s", err)
	}
	ongBlance, err := this.PrepareInvokeNativeContractWithRes(
		nutils.OngContractAddress,
		sdkcom.VERSION_CONTRACT_ONG,
		ont.BALANCEOF_NAME,
		[]interface{}{addr[:]},
		sdkcom.NEOVM_TYPE_INTEGER)
	if err != nil {
		return nil, fmt.Errorf("Get ONG balance of error:%s", err)
	}
	return &sdkcom.Balance{
		Ont: ontBalance.(*big.Int).Uint64(),
		Ong: ongBlance.(*big.Int).Uint64(),
	}, nil

	return this.GetBalanceWithBase58(addr.ToBase58())
}

//GetBalance return ont and ong balance of a ontology account in base58 code address
func (this *RpcClient) GetBalanceWithBase58(base58Addr string) (*sdkcom.Balance, error) {
	addr, err := common.AddressFromBase58(base58Addr)
	if err != nil {
		return nil, fmt.Errorf("AddressFromBase58 error:%s", err)
	}
	return this.GetBalance(addr)
}

//GetStorage return smart contract storage item.
//addr is smart contact address
//key is the key of value in smart contract
func (this *RpcClient) GetStorage(contractAddress common.Address, key []byte) ([]byte, error) {
	data, err := this.sendRpcRequest(RPC_GET_STORAGE, []interface{}{contractAddress.ToHexString(), hex.EncodeToString(key)})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetStorage(data)
}

//GetSmartContractEvent return smart contract event execute by invoke transaction.
func (this *RpcClient) GetSmartContractEvent(txHash common.Uint256) (*sdkcom.SmartContactEvent, error) {
	return this.GetSmartContractEventWithHexString(txHash.ToHexString())
}

//GetSmartContractEvent return smart contract event execute by invoke transaction by hex string code
func (this *RpcClient) GetSmartContractEventWithHexString(txHash string) (*sdkcom.SmartContactEvent, error) {
	data, err := this.sendRpcRequest(RPC_GET_SMART_CONTRACT_EVENT, []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetSmartContractEvent(data)
}

func (this *RpcClient) GetSmartContractEventByBlock(blockHeight uint32) ([]*sdkcom.SmartContactEvent, error) {
	data, err := this.sendRpcRequest(RPC_GET_SMART_CONTRACT_EVENT, []interface{}{blockHeight})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetSmartContactEvents(data)
}

//GetRawTransaction return transaction by transaction hash
func (this *RpcClient) GetRawTransaction(txHash common.Uint256) (*types.Transaction, error) {
	return this.GetRawTransactionWithHexString(txHash.ToHexString())
}

//GetRawTransaction return transaction by transaction hash in hex string code
func (this *RpcClient) GetRawTransactionWithHexString(txHash string) (*types.Transaction, error) {
	data, err := this.sendRpcRequest(RPC_GET_TRANSACTION, []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetTransaction(data)
}

//GetSmartContract return smart contract deployed in ontology by specified smart contract address
func (this *RpcClient) GetSmartContract(contractAddress common.Address) (*payload.DeployCode, error) {
	return this.GetSmartContractWithHexString(contractAddress.ToHexString())
}

func (this *RpcClient) GetSmartContractWithHexString(contractAddress string) (*payload.DeployCode, error) {
	data, err := this.sendRpcRequest(RPC_GET_SMART_CONTRACT, []interface{}{contractAddress})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetSmartContract(data)
}

func (this *RpcClient) GetGenerateBlockTime() (int, error) {
	data, err := this.sendRpcRequest(RPC_GET_GENERATE_BLOCK_TIME, []interface{}{})
	if err != nil {
		return 0, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetInt(data)
}

//GetMerkleProof return the merkle proof whether tx is exist in ledger
func (this *RpcClient) GetMerkleProof(txHash common.Uint256) (*sdkcom.MerkleProof, error) {
	return this.GetMerkleProofWithHexString(txHash.ToHexString())
}

//GetMerkleProof return the merkle proof whether tx is exist in ledger. Param txHash is in hex string code
func (this *RpcClient) GetMerkleProofWithHexString(txHash string) (*sdkcom.MerkleProof, error) {
	data, err := this.sendRpcRequest(RPC_GET_MERKLE_PROOF, []interface{}{txHash})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetMerkleProof(data)
}

//WaitForGenerateBlock Wait ontology generate block. Default wait 2 blocks.
//return timeout error when there is no block generate in some time.
func (this *RpcClient) WaitForGenerateBlock(timeout time.Duration, blockCount ...uint32) (bool, error) {
	return utils.WaitForGenerateBlock(this.GetBlockCount, timeout, blockCount...)
}

//Transfer ONT of ONG
//for ONT amount is the raw value
//for ONG amount is the raw value * 10e9
func (this *RpcClient) Transfer(gasPrice,
	gasLimit uint64,
	asset string,
	from *account.Account,
	to common.Address,
	amount uint64) (common.Uint256, error) {
	tx, err := this.NewTransferTransaction(gasPrice, gasLimit, asset, from.Address, to, amount)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	err = this.SignToTransaction(tx, from)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return this.SendRawTransaction(tx)
}

func (this *RpcClient) Allowance(asset string, from, to common.Address) (uint64, error) {
	type allowanceStruct struct {
		From common.Address
		To   common.Address
	}
	contractAddress, err := utils.GetAssetAddress(asset)
	if err != nil {
		return 0, err
	}
	allowance, err := this.PrepareInvokeNativeContractWithRes(
		contractAddress,
		sdkcom.VERSION_CONTRACT_ONT,
		sdkcom.NATIVE_ALLOWANCE,
		[]interface{}{&allowanceStruct{From: from, To: to}},
		sdkcom.NEOVM_TYPE_INTEGER)
	if err != nil {
		return 0, err
	}
	return allowance.(*big.Int).Uint64(), nil
}

func (this *RpcClient) Approve(gasPrice, gasLimit uint64,
	asset string,
	from *account.Account,
	to common.Address,
	amount uint64) (common.Uint256, error) {
	tx, err := this.NewApproveTransaction(gasPrice, gasLimit, asset, from.Address, to, amount)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	err = this.SignToTransaction(tx, from)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return this.SendRawTransaction(tx)
}

func (this *RpcClient) TransferFrom(gasPrice, gasLimit uint64,
	asset string,
	sender *account.Account,
	from, to common.Address,
	amount uint64) (common.Uint256, error) {
	tx, err := this.NewTransferFromTransaction(gasPrice, gasLimit, asset, sender.Address, from, to, amount)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	err = this.SignToTransaction(tx, sender)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return this.SendRawTransaction(tx)
}

func (this *RpcClient) UnboundONG(user common.Address) (uint64, error) {
	return this.Allowance("ong", nutils.OntContractAddress, user)
}

func (this *RpcClient) WithdrawONG(gasPrice, gasLimit uint64,
	user *account.Account,
	withdrawAmount ...uint64) (common.Uint256, error) {
	var amount uint64
	var err error
	if len(withdrawAmount) > 0 {
		amount = withdrawAmount[0]
	}
	if amount == 0 {
		amount, err = this.UnboundONG(user.Address)
		if err != nil {
			return common.UINT256_EMPTY, fmt.Errorf("Get UnboundONG error:%s", err)
		}
	}
	if amount == 0 {
		return common.UINT256_EMPTY, nil
	}
	return this.TransferFrom(gasPrice, gasLimit, "ong", user, nutils.OntContractAddress, user.Address, amount)
}

func (this *RpcClient) NewTransferTransaction(gasPrice, gasLimit uint64,
	asset string,
	from, to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewTransferTransaction(gasPrice, gasLimit, asset, from, to, amount)
}

func (this *RpcClient) NewApproveTransaction(gasPrice, gasLimit uint64,
	asset string, from, to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewApproveTransaction(gasPrice, gasLimit, asset, from, to, amount)
}

func (this *RpcClient) NewTransferFromTransaction(gasPrice, gasLimit uint64,
	asset string,
	sender, from, to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewTransferFromTransaction(gasPrice, gasLimit, asset, sender, from, to, amount)
}

//DeploySmartContract Deploy smart contract to ontology
func (this *RpcClient) DeploySmartContract(
	gasPrice,
	gasLimit uint64,
	singer *account.Account,
	needStorage bool,
	code,
	cname,
	cversion,
	cauthor,
	cemail,
	cdesc string) (common.Uint256, error) {

	invokeCode, err := hex.DecodeString(code)
	if err != nil {
		return common.UINT256_EMPTY, fmt.Errorf("code hex decode error:%s", err)
	}
	tx := this.NewDeployCodeTransaction(gasPrice, gasLimit, invokeCode, needStorage, cname, cversion, cauthor, cemail, cdesc)
	err = this.SignToTransaction(tx, singer)
	if err != nil {
		return common.Uint256{}, err
	}
	txHash, err := this.SendRawTransaction(tx)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("SendRawTransaction error:%s", err)
	}
	return txHash, nil
}

func (this *RpcClient) InvokeNativeContract(
	gasPrice,
	gasLimit uint64,
	singer *account.Account,
	cversion byte,
	contractAddress common.Address,
	method string,
	params []interface{},
) (common.Uint256, error) {
	tx, err := this.NewNativeInvokeTransaction(gasPrice, gasLimit, cversion, contractAddress, method, params)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	err = this.SignToTransaction(tx, singer)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return this.SendRawTransaction(tx)
}

//Invoke neo vm smart contract.
func (this *RpcClient) InvokeNeoVMContract(
	gasPrice,
	gasLimit uint64,
	signer *account.Account,
	contractAddress common.Address,
	params []interface{}) (common.Uint256, error) {

	tx, err := this.NewNeoVMSInvokeTransaction(gasPrice, gasLimit, contractAddress, params)
	if err != nil {
		return common.UINT256_EMPTY, fmt.Errorf("NewNeoVMSInvokeTransaction error:%s", err)
	}
	err = this.SignToTransaction(tx, signer)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return this.SendRawTransaction(tx)
}

func (this *RpcClient) NewDeployCodeTransaction(
	gasPrice, gasLimit uint64,
	code []byte,
	needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {
	return sdkcom.NewDeployCodeTransaction(gasPrice, gasLimit, code, needStorage, cname, cversion, cauthor, cemail, cdesc)
}

func (this *RpcClient) NewNativeInvokeTransaction(gasPrice,
	gasLimit uint64,
	cversion byte,
	contractAddress common.Address,
	method string,
	params []interface{},
) (*types.Transaction, error) {
	return utils.NewNativeInvokeTransaction(gasPrice, gasLimit, cversion, contractAddress, method, params)
}

func (this *RpcClient) NewNeoVMSInvokeTransaction(
	gasPrice, gasLimit uint64,
	contractAddress common.Address,
	params []interface{},
) (*types.Transaction, error) {
	return utils.NewNeoVMSInvokeTransaction(gasPrice, gasLimit, contractAddress, params)
}

//PrepareInvokeNeoVMContractWithRes Prepare invoke neovm contract, and return the value of result.
//Param returnType must be one of NeoVMReturnType, or array of NeoVMReturnType
func (this *RpcClient) PrepareInvokeNeoVMContractWithRes(contractAddress common.Address,
	params []interface{},
	returnType interface{}) (interface{}, error) {
	preResult, err := this.PrepareInvokeNeoVMContract(contractAddress, params)
	if err != nil {
		return nil, err
	}
	if preResult.State == 0 {
		return nil, fmt.Errorf("prepare inoke failed")
	}
	v, err := utils.ParsePreExecResult(preResult.Result, returnType)
	if err != nil {
		return nil, fmt.Errorf("ParseNeoVMContractReturnType error:%s", err)
	}
	return v, nil
}

func (this *RpcClient) PrepareInvokeNeoVMContract(contractAddress common.Address,
	params []interface{}) (*cstates.PreExecResult, error) {
	this.NewNeoVMSInvokeTransaction(0, 0, contractAddress, params)

	tx, err := this.NewNeoVMSInvokeTransaction(0, 0, contractAddress, params)
	if err != nil {
		return nil, fmt.Errorf("NewNeoVMSInvokeTransaction error:%s", err)
	}
	return this.PrepareInvokeContract(tx)
}

func (this *RpcClient) PrepareInvokeNativeContract(contractAddress common.Address,
	version byte,
	method string,
	params []interface{}) (*cstates.PreExecResult, error) {
	tx, err := this.NewNativeInvokeTransaction(0, 0, version, contractAddress, method, params)
	if err != nil {
		return nil, fmt.Errorf("NewNeoVMSInvokeTransaction error:%s", err)
	}
	return this.PrepareInvokeContract(tx)
}

//PrepareInvokeNativeContractWithRes Prepare invoke native contract, and return the value of result.
//Param returnType must be one of NeoVMReturnType, or array of NeoVMReturnType
func (this *RpcClient) PrepareInvokeNativeContractWithRes(contractAddress common.Address,
	version byte,
	method string,
	params []interface{}, returnType interface{}) (interface{}, error) {
	preResult, err := this.PrepareInvokeNativeContract(contractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	if preResult.State == 0 {
		return nil, fmt.Errorf("prepare inoke failed")
	}
	v, err := utils.ParsePreExecResult(preResult.Result, returnType)
	if err != nil {
		return nil, fmt.Errorf("ParseNeoVMContractReturnType error:%s", err)
	}
	return v, nil
}

//PrepareInvokeContract return the vm execute result of smart contract but not commit into ledger.
//It's useful for debugging smart contract.
func (this *RpcClient) PrepareInvokeContract(tx *types.Transaction) (*cstates.PreExecResult, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := this.sendRpcRequest(RPC_SEND_TRANSACTION, []interface{}{txData, 1})
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func (this *RpcClient) SignToTransaction(tx *types.Transaction, signer *account.Account) error {
	return sdkcom.SignToTransaction(tx, signer)
}

//SendRawTransaction send a transaction to ontology network, and return hash of the transaction
func (this *RpcClient) SendRawTransaction(tx *types.Transaction) (common.Uint256, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("Serialize error:%s", err)
	}
	txData := hex.EncodeToString(buffer.Bytes())
	data, err := this.sendRpcRequest(RPC_SEND_TRANSACTION, []interface{}{txData})
	if err != nil {
		return common.Uint256{}, err
	}
	return utils.GetUint256(data)
}

func (this *RpcClient) getQid() string {
	return fmt.Sprintf("%d", atomic.AddUint64(&this.qid, 1))
}

//sendRpcRequest send Rpc request to ontology
func (this *RpcClient) sendRpcRequest(method string, params []interface{}) ([]byte, error) {
	rpcReq := &JsonRpcRequest{
		Version: JSON_RPC_VERSION,
		Id:      this.getQid(),
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(rpcReq)
	if err != nil {
		return nil, fmt.Errorf("JsonRpcRequest json.Marsha error:%s", err)
	}
	resp, err := this.httpClient.Post(this.addr, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("http post request:%s error:%s", data, err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read rpc response body error:%s", err)
	}
	rpcRsp := &JsonRpcResponse{}
	err = json.Unmarshal(body, rpcRsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal JsonRpcResponse:%s error:%s", body, err)
	}
	if rpcRsp.Error != 0 {
		return nil, fmt.Errorf("JsonRpcResponse error code:%d desc:%s result:%s", rpcRsp.Error, rpcRsp.Desc, rpcRsp.Result)
	}
	return rpcRsp.Result, nil
}

//SendEmergencyGovReq return error
func (this *RpcClient) SendEmergencyGovReq(block []byte) error {
	blockString := hex.EncodeToString(block)
	_, err := this.sendRpcRequest(SEND_EMERGENCY_GOV_REQ, []interface{}{blockString})
	if err != nil {
		return fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return nil
}

//GetGetBlockRoot return common.Uint256
func (this *RpcClient) GetBlockRootWithNewTxRoot(txRoot common.Uint256) (common.Uint256, error) {
	hashString := hex.EncodeToString(txRoot.ToArray())
	data, err := this.sendRpcRequest(GET_BLOCK_ROOT_WITH_NEW_TX_ROOT, []interface{}{hashString})
	if err != nil {
		return common.Uint256{}, err
	}
	return utils.GetUint256(data)
}
