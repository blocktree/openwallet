package ontology

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/blocktree/OpenWallet/log"
	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

type Local struct {
	BaseURL     string
	AccessToken string
	Debug       bool
	client      *req.Req
}

func NewLocal(url string, debug bool) *Local {
	c := Local{
		BaseURL: url,
		//AccessToken: token,
		Debug: debug,
	}

	api := req.New()
	c.client = api

	return &c
}

func (l *Local) Call(path string, request interface{}, method string) (*gjson.Result, error) {
	if l.client == nil {
		return nil, errors.New("API url is not setup. ")
	}

	if l.Debug {
		log.Std.Debug("Start Request API...")
	}

	url := l.BaseURL + path

	r, err := l.client.Do(method, url, request)

	if l.Debug {
		log.Std.Debug("Request API Completed")
	}

	if l.Debug {
		log.Std.Debug("%+v", r)
	}

	err = l.isError(r)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	resp := gjson.ParseBytes(r.Bytes())

	return &resp, nil
}

func (l *Local) isError(resp *req.Resp) error {

	if resp == nil || resp.Response() == nil {
		return errors.New("Response is empty! ")
	}

	if resp.Response().StatusCode != http.StatusOK {
		return fmt.Errorf("%s", resp.String())
	}

	return nil
}

func (wm *WalletManager) getBlockHeightByLocal() (uint64, error) {

	path := "block/height"

	result, err := wm.LocalClient.Call(path, nil, "GET")
	if err != nil {
		return 0, err
	}

	height := result.Get("Result").Uint()

	return height, nil
}

func (wm *WalletManager) getBlockHashByLocal(height uint64) (string, error) {

	path := fmt.Sprintf("block/details/height/%d", height)

	result, err := wm.LocalClient.Call(path, nil, "GET")
	if err != nil {
		return "", err
	}

	return result.Get("Result").Get("Hash").String(), nil
}

func (wm *WalletManager) getBalanceByLocal(address string) (*AddrBalance, error) {
	path1 := fmt.Sprintf("balance/%s", address)

	result1, err := wm.LocalClient.Call(path1, nil, "GET")

	if err != nil {
		return nil, err
	}

	path2 := fmt.Sprintf("unboundong/%s", address)

	result2, err := wm.LocalClient.Call(path2, nil, "GET")

	if err != nil {
		return nil, err
	}

	balance := newAddrBalance([]*gjson.Result{result1, result2})

	if balance == nil {
		return nil, errors.New("Get balance by local failed!")
	}

	balance.Address = address

	return balance, nil
}
