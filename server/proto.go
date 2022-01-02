package server

import (
	"encoding/json"
)

//矿机的请求参数
type JSONRpcReq struct {
	Id     json.RawMessage `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

//矿机的请求参数
type MinerReq struct {
	JSONRpcReq
	//矿机提交的参数中,矿机标识
	Worker string `json:"worker"`
}

type ErrorReply struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

//矿池反馈的请求参数
type PoolResp struct {
	Id      *json.RawMessage       `json:"id"`
	Jsonrpc string                 `json:"jsonrpc"`
	Result  *json.RawMessage       `json:"result"`
	Error   map[string]interface{} `json:"error"`
}

//反馈给矿机的请求参数
type MinerResp struct {
	Id      json.RawMessage `json:"id"`
	Jsonrpc string          `json:"jsonrpc"`
	Result  interface{}     `json:"result"`
	Error   interface{}     `json:"error,omitempty"`
}