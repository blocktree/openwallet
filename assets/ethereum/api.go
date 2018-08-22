package ethereum

import (
	"errors"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/blocktree/OpenWallet/common"
	"github.com/blocktree/OpenWallet/logger"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

type Client struct {
	BaseURL string
	Debug   bool
}

type Response struct {
	Id      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
}

const (
	ETH_GET_TOKEN_BALANCE_METHOD      = "0x70a08231"
	ETH_TRANSFER_TOKEN_BALANCE_METHOD = "0xa9059cbb"
)

const (
	SOLIDITY_TYPE_ADDRESS = "address"
	SOLIDITY_TYPE_UINT256 = "uint256"
	SOLIDITY_TYPE_UINT160 = "uint160"
)

type SolidityParam struct {
	ParamType  string
	ParamValue interface{}
}

func makeRepeatString(c string, count uint) string {
	cs := make([]string, 0)
	for i := 0; i < int(count); i++ {
		cs = append(cs, c)
	}
	return strings.Join(cs, "")
}

func makeTransactionData(methodId string, params []SolidityParam) (string, error) {

	data := methodId
	for i, _ := range params {
		var param string
		if params[i].ParamType == SOLIDITY_TYPE_ADDRESS {
			param = strings.ToLower(params[i].ParamValue.(string))
			if strings.Index(param, "0x") != -1 {
				param = common.Substr(param, 2, len(param))
			}

			if len(param) != 40 {
				return "", errors.New("length of address error.")
			}
			param = makeRepeatString("0", 24) + param
		} else if params[i].ParamType == SOLIDITY_TYPE_UINT256 {
			intParam := params[i].ParamValue.(*big.Int)
			param = intParam.Text(16)
			l := len(param)
			if l > 64 {
				return "", errors.New("integer overflow.")
			}

			param = makeRepeatString("0", uint(64-l)) + param
		} else {
			return "", errors.New("not support solidity type")
		}

		data += param
	}
	return data, nil
}

func ERC20GetAddressBalance(address string, contractAddr string) (*big.Int, error) {

	var funcParams []SolidityParam
	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_ADDRESS,
		ParamValue: address,
	})
	trans := make(map[string]interface{})
	data, err := makeTransactionData(ETH_GET_TOKEN_BALANCE_METHOD, funcParams)
	if err != nil {
		openwLogger.Log.Errorf("make transaction data failed, err = %v", err)
		return nil, err
	}

	trans["to"] = contractAddr
	trans["data"] = data
	params := []interface{}{
		trans,
		"latest",
	}
	result, err := client.Call("eth_call", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] erc20 balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] erc20 balance format failed, response is %v\n", address, result.Type))
		return big.NewInt(0), err
	}

	balance, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert addr[%v] erc20 balance format to bigint failed, response is %v, and err = %v\n", address, result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return balance, nil
}

func GetAddrBalance(address string) (*big.Int, error) {

	params := []interface{}{
		address,
		"latest",
	}
	result, err := client.Call("eth_getBalance", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] balance failed, err=%v\n", address, err))
		return big.NewInt(0), err
	}
	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get addr[%v] balance format failed, response is %v\n", address, result.Type))
		return big.NewInt(0), err
	}

	balance, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert addr[%v] balance format to bigint failed, response is %v, and err = %v\n", address, result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return balance, nil
}

func makeSimpleTransactionPara(fromAddr *Address, toAddr string, amount *big.Int, password string, fee *txFeeInfo) map[string]interface{} {
	paraMap := make(map[string]interface{})

	//use password to unlock the account
	paraMap["password"] = password
	//use the following attr to eth_sendTransaction
	paraMap["from"] = fromAddr.Address
	paraMap["to"] = toAddr
	paraMap["value"] = "0x" + amount.Text(16)
	paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	return paraMap
}

func makeSimpleTransGasEstimatedPara(fromAddr string, toAddr string, amount *big.Int) map[string]interface{} {
	paraMap := make(map[string]interface{})
	paraMap["from"] = fromAddr
	paraMap["to"] = toAddr
	paraMap["value"] = "0x" + amount.Text(16)
	return paraMap
}

func makeERC20TokenTransData(contractAddr string, toAddr string, amount *big.Int) (string, error) {
	var funcParams []SolidityParam
	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_ADDRESS,
		ParamValue: toAddr,
	})

	funcParams = append(funcParams, SolidityParam{
		ParamType:  SOLIDITY_TYPE_UINT256,
		ParamValue: amount,
	})

	data, err := makeTransactionData(ETH_TRANSFER_TOKEN_BALANCE_METHOD, funcParams)
	if err != nil {
		openwLogger.Log.Errorf("make transaction data failed, err = %v", err)
		return "", err
	}
	return data, nil
}

func makeERC20TokenTransGasEstimatePara(fromAddr string, contractAddr string, data string) map[string]interface{} {

	paraMap := make(map[string]interface{})

	//use password to unlock the account
	//use the following attr to eth_sendTransaction
	paraMap["from"] = fromAddr //fromAddr.Address
	paraMap["to"] = contractAddr
	//paraMap["value"] = "0x" + amount.Text(16)
	//paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	//paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	paraMap["data"] = data
	return paraMap
}

func ethGetGasEstimated(paraMap map[string]interface{}) (*big.Int, error) {
	trans := make(map[string]interface{})
	var temp interface{}
	var exist bool
	var fromAddr string
	var toAddr string

	if temp, exist = paraMap["from"]; !exist {
		openwLogger.Log.Errorf("from not found")
		return big.NewInt(0), errors.New("from not found")
	} else {
		fromAddr = temp.(string)
		trans["from"] = fromAddr
	}

	if temp, exist = paraMap["to"]; !exist {
		openwLogger.Log.Errorf("to not found")
		return big.NewInt(0), errors.New("to not found")
	} else {
		toAddr = temp.(string)
		trans["to"] = toAddr
	}

	if temp, exist = paraMap["value"]; exist {
		amount := temp.(string)
		trans["value"] = amount
	}

	if temp, exist = paraMap["data"]; exist {
		data := temp.(string)
		trans["data"] = data
	}

	params := []interface{}{
		trans,
	}

	result, err := client.Call("eth_estimateGas", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("get estimated gas limit from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return big.NewInt(0), err
	}

	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("get estimated gas from [%v] to [%v] failed, response is %v\n", fromAddr, toAddr, err))
		return big.NewInt(0), err
	}

	gasLimit, err := convertToBigInt(result.String(), 16)
	if err != nil {
		errInfo := fmt.Sprintf("convert estimated gas[%v] format to bigint failed, err = %v\n", result.String(), err)
		openwLogger.Log.Errorf(errInfo)
		return big.NewInt(0), errors.New(errInfo)
	}
	return gasLimit, nil
}

func makeERC20TokenTransactionPara(fromAddr *Address, contractAddr string, data string,
	password string, fee *txFeeInfo) map[string]interface{} {

	paraMap := make(map[string]interface{})

	//use password to unlock the account
	paraMap["password"] = password
	//use the following attr to eth_sendTransaction
	paraMap["from"] = fromAddr.Address
	paraMap["to"] = contractAddr
	//paraMap["value"] = "0x" + amount.Text(16)
	paraMap["gas"] = "0x" + fee.GasLimit.Text(16)
	paraMap["gasPrice"] = "0x" + fee.GasPrice.Text(16)
	paraMap["data"] = data
	return paraMap
}

func SendTransactionToAddr(param map[string]interface{}) (string, error) {
	//(addr *Address, to string, amount *big.Int, password string, fee *txFeeInfo) (string, error) {
	var exist bool
	var temp interface{}
	if temp, exist = param["from"]; !exist {
		openwLogger.Log.Errorf("from not found.")
		return "", errors.New("from not found.")
	}

	fromAddr := temp.(string)

	if temp, exist = param["password"]; !exist {
		openwLogger.Log.Errorf("password not found.")
		return "", errors.New("password not found.")
	}

	password := temp.(string)

	err := UnlockAddr(fromAddr, password, 300)
	if err != nil {
		openwLogger.Log.Errorf("unlock addr failed, err = %v", err)
		return "", err
	}

	txId, err := ethSendTransaction(param)
	if err != nil {
		openwLogger.Log.Errorf("ethSendTransaction failed, err = %v", err)
		return "", err
	}

	err = LockAddr(fromAddr)
	if err != nil {
		openwLogger.Log.Errorf("lock addr failed, err = %v", err)
		return txId, err
	}

	return txId, nil
}

func ethSendTransaction(paraMap map[string]interface{}) (string, error) {
	//(fromAddr string, toAddr string, amount *big.Int, fee *txFeeInfo) (string, error) {
	trans := make(map[string]interface{})
	var temp interface{}
	var exist bool
	var fromAddr string
	var toAddr string

	if temp, exist = paraMap["from"]; !exist {
		openwLogger.Log.Errorf("from not found")
		return "", errors.New("from not found")
	} else {
		fromAddr = temp.(string)
		trans["from"] = fromAddr
	}

	if temp, exist = paraMap["to"]; !exist {
		openwLogger.Log.Errorf("to not found")
		return "", errors.New("to not found")
	} else {
		toAddr = temp.(string)
		trans["to"] = toAddr
	}

	if temp, exist = paraMap["value"]; exist {
		amount := temp.(string)
		trans["value"] = amount
	}

	if temp, exist = paraMap["gas"]; exist {
		gasLimit := temp.(string)
		trans["gas"] = gasLimit
	}

	if temp, exist = paraMap["gasPrice"]; exist {
		gasPrice := temp.(string)
		trans["gasPrice"] = gasPrice
	}

	if temp, exist = paraMap["data"]; exist {
		data := temp.(string)
		trans["data"] = data
	}

	params := []interface{}{
		trans,
	}

	result, err := client.Call("eth_sendTransaction", 1, params)
	if err != nil {
		openwLogger.Log.Errorf(fmt.Sprintf("start transaction from [%v] to [%v] faield, err = %v \n", fromAddr, toAddr, err))
		return "", err
	}

	if result.Type != gjson.String {
		openwLogger.Log.Errorf(fmt.Sprintf("send transaction from [%v] to [%v] failed, response is %v\n", fromAddr, toAddr, err))
		return "", err
	}
	return result.String(), nil
}

func (c *Client) Call(method string, id int64, params []interface{}) (*gjson.Result, error) {
	authHeader := req.Header{
		"Accept": "application/json",
		//		"Authorization": "Basic " + c.AccessToken,
	}
	body := make(map[string]interface{}, 0)
	body["jsonrpc"] = "2.0"
	body["id"] = id
	body["method"] = method
	body["params"] = params

	if c.Debug {
		log.Println("Start Request API...")
	}

	r, err := req.Post(c.BaseURL, req.BodyJSON(&body), authHeader)

	if c.Debug {
		log.Println("Request API Completed")
	}

	if c.Debug {
		log.Printf("%+v\n", r)
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())
	err = isError(&resp)
	if err != nil {
		return nil, err
	}

	result := resp.Get("result")

	return &result, nil
}

//isError 是否报错
func isError(result *gjson.Result) error {
	var (
		err error
	)

	if !result.Get("error").IsObject() {

		if !result.Get("result").Exists() {
			return errors.New("Response is empty! ")
		}

		return nil
	}

	errInfo := fmt.Sprintf("[%d]%s",
		result.Get("error.code").Int(),
		result.Get("error.message").String())
	err = errors.New(errInfo)

	return err
}
