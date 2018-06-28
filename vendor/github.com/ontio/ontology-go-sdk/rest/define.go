package rest

import (
	"encoding/json"
)

const (
	GET_GEN_BLK_TIME      = "/api/v1/node/generateblocktime"
	GET_CONN_COUNT        = "/api/v1/node/connectioncount"
	GET_BLK_TXS_BY_HEIGHT = "/api/v1/block/transactions/height/"
	GET_BLK_BY_HEIGHT     = "/api/v1/block/details/height/"
	GET_BLK_BY_HASH       = "/api/v1/block/details/hash/"
	GET_BLK_HEIGHT        = "/api/v1/block/height"
	GET_BLK_HASH          = "/api/v1/block/hash/"
	GET_TX                = "/api/v1/transaction/"
	GET_STORAGE           = "/api/v1/storage/"
	GET_BALANCE           = "/api/v1/balance/"
	GET_CONTRACT_STATE    = "/api/v1/contract/"
	GET_SMTCOCE_EVT_TXS   = "/api/v1/smartcode/event/transactions/"
	GET_SMTCOCE_EVTS      = "/api/v1/smartcode/event/txhash/"
	GET_BLK_HGT_BY_TXHASH = "/api/v1/block/height/txhash/"
	GET_MERKLE_PROOF      = "/api/v1/merkleproof/"
	GET_GAS_PRICE         = "/api/v1/gasprice"
	GET_ALLOWANCE         = "/api/v1/allowance/"
	GET_UNBOUNDONG        = "/api/v1/unboundong/"
	GET_MEMPOOL_TXCOUNT   = "/api/v1/mempool/txcount"
	GET_MEMPOOL_TXSTATE   = "/api/v1/mempool/txstate/"
	GET_VERSION           = "/api/v1/version"
	POST_RAW_TX           = "/api/v1/transaction"
)

const (
	ACTION_SEND_RAW_TRANSACTION = "sendrawtransaction"
)

const REST_VERSION = "1.0.0"

type RestfulReq struct {
	Action  string
	Version string
	Type    int
	Data    string
}

type RestfulResp struct {
	Action  string          `json:"action"`
	Result  json.RawMessage `json:"result"`
	Error   int64           `json:"error"`
	Desc    string          `json:"desc"`
	Version string          `json:"version"`
}
