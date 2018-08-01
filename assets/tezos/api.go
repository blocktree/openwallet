package tezos

import (
	"github.com/imroc/req"
	"log"
)
var (
	api      = req.New()
	header = req.Header{"Content-Type": "application/json"}
)


func callGetHeader() []byte {
	url := serverAPI + "/chains/main/blocks/head/header"
	param := make(req.QueryParam, 0)

	r, err := api.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func callGetCounter(pkh string) []byte {
	url := serverAPI + "/chains/main/blocks/head/context/contracts/" + pkh + "/counter"

	r, err := api.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	//因为结果为"number"\n ，所以去掉双引号和\n
	lenght := len(r.Bytes())
	return r.Bytes()[1:lenght-2]
}

func callGetManagerKey(pkh string) []byte {
	url := serverAPI + "/chains/main/blocks/head/context/contracts/" + pkh + "/manager_key"

	r, err := api.Get(url)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func callForgeOps(chain_id string, head_hash string, body interface{}) string {
	url := serverAPI + "/chains/" + chain_id + "/blocks/" + head_hash + "/helpers/forge/operations"
	param := make(req.Param, 0)

	log.Println(body)
	r, err := api.Post(url, param, header, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return ""
	}
	//因为结果为"hex"\n ，所以去掉双引号和\n
	lenght := len(r.Bytes())
	return string(r.Bytes()[1:lenght-2])
}

func callPreapplyOps(body interface{}) []byte{
	url := serverAPI + "/chains/main/blocks/head/helpers/preapply/operations"
	param := make(req.Param, 0)

	r, err := api.Post(url, param, header, req.BodyJSON(&body))
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

func callInjectOps(body string) []byte {
	url := serverAPI + "/injection/operation"
	param := make(req.Param, 0)

	r, err := api.Post(url, param, header, req.BodyJSON(&body))

	if err != nil {
		log.Println(err.Error())
		return nil
	}

	return r.Bytes()
}

func callGetbalance(addr string) []byte {
	url := serverAPI + "/chains/main/blocks/head/context/contracts/" + addr + "/balance"
	param := make(req.QueryParam, 0)

	r, err := api.Get(url, param)
	if err != nil {
		log.Println(err)
		return nil
	}

	return r.Bytes()
}

