package rest

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
	"io"
	"io/ioutil"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync/atomic"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

//RpcClient for ontology rpc api
type RestClient struct {
	qid        uint64
	addr       string
	httpClient *http.Client
}

//NewRpcClient return RpcClient instance
func NewRestClient() *RestClient {
	return &RestClient{
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

//SetAddress set rest server address. Simple http://localhost:20334
func (this *RestClient) SetAddress(addr string) *RestClient {
	this.addr = addr
	return this
}

//SetHttpClient set rest client to RestClient. In most cases SetHttpClient is not necessary
func (this *RestClient) SetHttpClient(httpClient *http.Client) *RestClient {
	this.httpClient = httpClient
	return this
}

func (this *RestClient) GetVersion() (string, error) {
	reqPath := GET_VERSION
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return "", err
	}
	return utils.GetVersion(data)
}

func (this *RestClient) GetBlockByHash(hash common.Uint256) (*types.Block, error) {
	return this.GetBlockByHashWithHexString(hash.ToHexString())
}

func (this *RestClient) GetBlockByHashWithHexString(hash string) (*types.Block, error) {
	reqPath := GET_BLK_BY_HASH + hash
	reqValues := &url.Values{}
	reqValues.Add("raw", "1")
	data, err := this.sendRestGetRequest(reqPath, reqValues)
	if err != nil {
		return nil, err
	}
	return utils.GetBlock(data)
}

func (this *RestClient) GetBlockByHeight(height uint32) (*types.Block, error) {
	reqPath := fmt.Sprintf("%s%d", GET_BLK_BY_HEIGHT, height)
	reqValues := &url.Values{}
	reqValues.Add("raw", "1")
	data, err := this.sendRestGetRequest(reqPath, reqValues)
	if err != nil {
		return nil, err
	}
	return utils.GetBlock(data)
}

func (this *RestClient) GetCurrentBlockHeight() (uint32, error) {
	reqPath := GET_BLK_HEIGHT
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return 0, err
	}
	return utils.GetUint32(data)
}

func (this *RestClient) GetBlockHash(height uint32) (common.Uint256, error) {
	reqPath := fmt.Sprintf("%s%d", GET_BLK_HASH, height)
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return utils.GetUint256(data)
}

//GetRawTransaction return transaction by transaction hash
func (this *RestClient) GetRawTransaction(txHash common.Uint256) (*types.Transaction, error) {
	return this.GetRawTransactionWithHexString(txHash.ToHexString())
}

//GetRawTransaction return transaction by transaction hash in hex string code
func (this *RestClient) GetRawTransactionWithHexString(txHash string) (*types.Transaction, error) {
	reqPath := GET_TX + txHash
	reqValues := &url.Values{}
	reqValues.Add("raw", "1")
	data, err := this.sendRestGetRequest(reqPath, reqValues)
	if err != nil {
		return nil, err
	}
	return utils.GetTransaction(data)
}

//GetBalance return ont and ong balance of a ontology account
func (this *RestClient) GetBalance(addr common.Address) (*sdkcom.Balance, error) {
	ontBalance, err := this.PrepareInvokeNativeContractWithRes(nutils.OntContractAddress,
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
}

//GetBalance return ont and ong balance of a ontology account in base58 code address
func (this *RestClient) GetBalanceWithBase58(base58Addr string) (*sdkcom.Balance, error) {
	addr, err := common.AddressFromBase58(base58Addr)
	if err != nil {
		return nil, fmt.Errorf("AddressFromBase58 error:%s", err)
	}
	return this.GetBalance(addr)
}

func (this *RestClient) GetStorage(contractAddress common.Address, key []byte) ([]byte, error) {
	reqPath := GET_STORAGE + contractAddress.ToHexString() + "/" + hex.EncodeToString(key)
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return nil, err
	}
	return utils.GetStorage(data)
}

//GetSmartContractEvent return smart contract event execute by invoke transaction by hex string code
func (this *RestClient) GetSmartContractEventWithHexString(txHash string) (*sdkcom.SmartContactEvent, error) {
	reqPath := GET_SMTCOCE_EVTS + txHash
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetSmartContractEvent(data)
}

func (this *RestClient) GetSmartContractEventByBlock(blockHeight uint32) ([]*sdkcom.SmartContactEvent, error) {
	reqPath := fmt.Sprintf("%s%d", GET_SMTCOCE_EVT_TXS, blockHeight)
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return nil, fmt.Errorf("sendRpcRequest error:%s", err)
	}
	return utils.GetSmartContactEvents(data)
}

func (this *RestClient) GetSmartContract(contractAddress common.Address) (*payload.DeployCode, error) {
	return this.GetSmartContractWithHexString(contractAddress.ToHexString())
}

func (this *RestClient) GetSmartContractWithHexString(contractAddress string) (*payload.DeployCode, error) {
	reqPath := GET_CONTRACT_STATE + contractAddress
	reqValues := &url.Values{}
	reqValues.Add("raw", "1")
	data, err := this.sendRestGetRequest(reqPath, reqValues)
	if err != nil {
		return nil, err
	}
	return utils.GetSmartContract(data)
}

func (this *RestClient) GetMerkleProof(txHash common.Uint256) (*sdkcom.MerkleProof, error) {
	return this.GetMerkleProofWithHexString(txHash.ToHexString())
}

func (this RestClient) GetMerkleProofWithHexString(txHash string) (*sdkcom.MerkleProof, error) {
	reqPath := GET_MERKLE_PROOF + txHash
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return nil, err
	}
	return utils.GetMerkleProof(data)
}

//WaitForGenerateBlock Wait ontology generate block. Default wait 2 blocks.
//return timeout error when there is no block generate in some time.
func (this *RestClient) WaitForGenerateBlock(timeout time.Duration, blockCount ...uint32) (bool, error) {
	return utils.WaitForGenerateBlock(this.GetCurrentBlockHeight, timeout, blockCount...)
}

func (this *RestClient) GetGenerateBlockTime() (int, error) {
	reqPath := GET_GEN_BLK_TIME
	data, err := this.sendRestGetRequest(reqPath)
	if err != nil {
		return 0, err
	}
	return utils.GetInt(data)
}

//Transfer ONT of ONG
//for ONT amount is the raw value
//for ONG amount is the raw value * 10e9
func (this *RestClient) Transfer(gasPrice, gasLimit uint64,
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

func (this *RestClient) Allowance(asset string, from, to common.Address) (uint64, error) {
	type allowanceStruct struct {
		From common.Address
		To   common.Address
	}
	contractAddress, err := utils.GetAssetAddress(asset)
	if err != nil {
		return 0, err
	}
	allowance, err := this.PrepareInvokeNativeContractWithRes(contractAddress,
		sdkcom.VERSION_CONTRACT_ONT,
		sdkcom.NATIVE_ALLOWANCE,
		[]interface{}{&allowanceStruct{From: from, To: to}},
		sdkcom.NEOVM_TYPE_INTEGER)
	if err != nil {
		return 0, err
	}
	return allowance.(*big.Int).Uint64(), nil
}

func (this *RestClient) Approve(gasPrice, gasLimit uint64,
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

func (this *RestClient) TransferFrom(gasPrice, gasLimit uint64,
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

func (this *RestClient) UnboundONG(user common.Address) (uint64, error) {
	return this.Allowance("ong", nutils.OntContractAddress, user)
}

func (this *RestClient) WithdrawONG(gasPrice,
	gasLimit uint64,
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

func (this *RestClient) NewTransferTransaction(gasPrice,
	gasLimit uint64,
	asset string,
	from,
	to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewTransferTransaction(gasPrice, gasLimit, asset, from, to, amount)
}

func (this *RestClient) NewApproveTransaction(gasPrice,
	gasLimit uint64,
	asset string,
	from,
	to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewApproveTransaction(gasPrice, gasLimit, asset, from, to, amount)
}

func (this *RestClient) NewTransferFromTransaction(gasPrice,
	gasLimit uint64,
	asset string,
	sender,
	from,
	to common.Address,
	amount uint64) (*types.Transaction, error) {
	return utils.NewTransferFromTransaction(gasPrice, gasLimit, asset, sender, from, to, amount)
}

//DeploySmartContract Deploy smart contract to ontology
func (this *RestClient) DeploySmartContract(
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

func (this *RestClient) InvokeNativeContract(
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
func (this *RestClient) InvokeNeoVMContract(
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

func (this *RestClient) NewDeployCodeTransaction(
	gasPrice, gasLimit uint64,
	code []byte,
	needStorage bool,
	cname, cversion, cauthor, cemail, cdesc string) *types.Transaction {
	return sdkcom.NewDeployCodeTransaction(gasPrice, gasLimit, code, needStorage, cname, cversion, cauthor, cemail, cdesc)
}

//PrepareInvokeNeoVMContractWithRes Prepare invoke neovm contract, and return the value of result.
//Param returnType must be one of NeoVMReturnType, or array of NeoVMReturnType
func (this *RestClient) PrepareInvokeNeoVMContractWithRes(
	contractAddress common.Address,
	params []interface{},
	returnType interface{}) (interface{}, error) {
	preResult, err := this.PrepareInvokeNeoVMContract(contractAddress, params)
	if err != nil {
		return nil, err
	}
	v, err := utils.ParsePreExecResult(preResult.Result, returnType)
	if err != nil {
		return nil, fmt.Errorf("ParseNeoVMContractReturnType error:%s", err)
	}
	return v, nil
}

func (this *RestClient) PrepareInvokeNeoVMContract(contractAddress common.Address,
	params []interface{}) (*cstates.PreExecResult, error) {
	this.NewNeoVMSInvokeTransaction(0, 0, contractAddress, params)

	tx, err := this.NewNeoVMSInvokeTransaction(0, 0, contractAddress, params)
	if err != nil {
		return nil, fmt.Errorf("NewNeoVMSInvokeTransaction error:%s", err)
	}
	return this.PrepareInvokeContract(tx)
}

func (this *RestClient) PrepareInvokeNativeContract(
	contractAddress common.Address,
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
func (this *RestClient) PrepareInvokeNativeContractWithRes(
	contractAddress common.Address,
	version byte,
	method string,
	params []interface{}, returnType interface{}) (interface{}, error) {
	preResult, err := this.PrepareInvokeNativeContract(contractAddress, version, method, params)
	if err != nil {
		return nil, err
	}
	v, err := utils.ParsePreExecResult(preResult.Result, returnType)
	if err != nil {
		return nil, fmt.Errorf("ParseNeoVMContractReturnType error:%s", err)
	}
	return v, nil
}

func (this *RestClient) NewNativeInvokeTransaction(gasPrice,
	gasLimit uint64,
	cversion byte,
	contractAddress common.Address,
	method string,
	params []interface{},
) (*types.Transaction, error) {
	return utils.NewNativeInvokeTransaction(gasPrice, gasLimit, cversion, contractAddress, method, params)
}

func (this *RestClient) NewNeoVMSInvokeTransaction(
	gasPrice, gasLimit uint64,
	contractAddress common.Address,
	params []interface{},
) (*types.Transaction, error) {
	return utils.NewNeoVMSInvokeTransaction(gasPrice, gasLimit, contractAddress, params)
}

func (this *RestClient) SignToTransaction(tx *types.Transaction, signer *account.Account) error {
	return sdkcom.SignToTransaction(tx, signer)
}

func (this *RestClient) SendRawTransaction(tx *types.Transaction) (common.Uint256, error) {
	reqPath := POST_RAW_TX
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return common.Uint256{}, fmt.Errorf("Serialize error:%s", err)
	}
	data, err := this.sendRestPostRequest(buffer.Bytes(), reqPath)
	if err != nil {
		return common.UINT256_EMPTY, err
	}
	return utils.GetUint256(data)
}

//PrepareInvokeContract return the vm execute result of smart contract but not commit into ledger.
//It's useful for debugging smart contract.
func (this *RestClient) PrepareInvokeContract(tx *types.Transaction) (*cstates.PreExecResult, error) {
	var buffer bytes.Buffer
	err := tx.Serialize(&buffer)
	if err != nil {
		return nil, fmt.Errorf("Serialize error:%s", err)
	}
	reqPath := POST_RAW_TX
	reqValues := &url.Values{}
	reqValues.Add("preExec", "1")
	data, err := this.sendRestPostRequest(buffer.Bytes(), reqPath, reqValues)
	if err != nil {
		return nil, err
	}
	preResult := &cstates.PreExecResult{}
	err = json.Unmarshal(data, &preResult)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal PreExecResult:%s error:%s", data, err)
	}
	return preResult, nil
}

func (this *RestClient) getAddress() (string, error) {
	if this.addr == "" {
		return "", fmt.Errorf("cannot get address, please add adrress first")
	}
	return this.addr, nil
}

func (this *RestClient) getQid() string {
	return fmt.Sprintf("%d", atomic.AddUint64(&this.qid, 1))
}

func (this *RestClient) getRequestUrl(reqPath string, values ...*url.Values) (string, error) {
	addr, err := this.getAddress()
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(addr, "http") {
		addr = "http://" + addr
	}
	reqUrl, err := new(url.URL).Parse(addr)
	if err != nil {
		return "", fmt.Errorf("Parse address:%s error:%s", addr, err)
	}
	reqUrl.Path = reqPath
	if len(values) > 0 {
		reqUrl.RawQuery = values[0].Encode()
	}
	return reqUrl.String(), nil
}

func (this *RestClient) sendRestGetRequest(reqPath string, values ...*url.Values) ([]byte, error) {
	reqUrl, err := this.getRequestUrl(reqPath, values...)
	if err != nil {
		return nil, err
	}
	resp, err := this.httpClient.Get(reqUrl)
	if err != nil {
		return nil, fmt.Errorf("send http get request error:%s", err)
	}
	defer resp.Body.Close()
	return this.dealRestResponse(resp.Body)
}

func (this *RestClient) sendRestPostRequest(data []byte, reqPath string, values ...*url.Values) ([]byte, error) {
	reqUrl, err := this.getRequestUrl(reqPath, values...)
	if err != nil {
		return nil, err
	}
	restReq := &RestfulReq{
		Action:  ACTION_SEND_RAW_TRANSACTION,
		Version: REST_VERSION,
		Data:    hex.EncodeToString(data),
	}
	reqData, err := json.Marshal(restReq)
	if err != nil {
		return nil, fmt.Errorf("json.Marshal error:%s", err)
	}
	resp, err := this.httpClient.Post(reqUrl, "application/json", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("send http post request error:%s", err)
	}
	defer resp.Body.Close()
	return this.dealRestResponse(resp.Body)
}

func (this *RestClient) dealRestResponse(body io.Reader) ([]byte, error) {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read http body error:%s", err)
	}
	restRsp := &RestfulResp{}
	err = json.Unmarshal(data, restRsp)
	if err != nil {
		return nil, fmt.Errorf("json.Unmarshal RestfulResp:%s error:%s", body, err)
	}
	if restRsp.Error != 0 {
		return nil, fmt.Errorf("sendRestRequest error code:%d desc:%s result:%s", restRsp.Error, restRsp.Desc, restRsp.Result)
	}
	return restRsp.Result, nil
}
