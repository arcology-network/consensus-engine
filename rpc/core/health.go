package core

import (
	ctypes "github.com/arcology-network/consensus-engine/rpc/core/types"
	rpctypes "github.com/arcology-network/consensus-engine/rpc/jsonrpc/types"
)

// Health gets node health. Returns empty result (200 OK) on success, no
// response - in case of an error.
// More: https://docs.tendermint.com/master/rpc/#/Info/health
func Health(ctx *rpctypes.Context) (*ctypes.ResultHealth, error) {
	return &ctypes.ResultHealth{}, nil
}
