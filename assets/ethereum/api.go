package ethereum

import (
	"errors"
	"fmt"
	"log"

	"github.com/imroc/req"
	"github.com/tidwall/gjson"
)

/*
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": [
        "0x50068fd632c1a6e6c5bd407b4ccf8861a589e776",
        "0x2a63b2203955b84fefe52baca3881b3614991b34"
    ]
}

*/

type Client struct {
	BaseURL string
	Debug   bool
}

type Response struct {
	Id      int         `json:"id"`
	Version string      `json:"jsonrpc"`
	Result  interface{} `json:"result"`
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
