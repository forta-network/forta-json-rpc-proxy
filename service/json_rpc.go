package service

import (
	"encoding/json"
)

var (
	jsonNull       = json.RawMessage("null")
	jsonRpcVersion = "2.0"
)

type jsonrpcErrorMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
}

type jsonError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

var (
	methodParseError    []byte
	methodNotFoundError []byte
)

func init() {
	var err error

	// The errors are from go-ethereum rpc error codes.

	methodParseError, err = json.Marshal(&jsonrpcErrorMessage{
		Version: jsonRpcVersion,
		ID:      jsonNull,
		Error: &jsonError{
			Code:    -32600,
			Message: "failed to parse json-rpc method",
		},
	})
	if err != nil {
		panic(err)
	}

	methodNotFoundError, err = json.Marshal(&jsonrpcErrorMessage{
		Version: jsonRpcVersion,
		ID:      jsonNull,
		Error: &jsonError{
			Code:    -32601,
			Message: "method not available",
		},
	})
	if err != nil {
		panic(err)
	}
}
