package testgen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/rpc"
)

// firstBALBlock returns the earliest block that commits to an EIP-7928 block
// access list in its header.
func firstBALBlock(t *T) *types.Block {
	for i := 0; i <= int(t.chain.Head().NumberU64()); i++ {
		if b := t.chain.GetBlock(i); b.Header().BlockAccessListHash != nil {
			return b
		}
	}
	panic("chain has no block with a block access list")
}

// balAccountFields are the fields required on every AccountAccess object.
var balAccountFields = []string{"address", "storageChanges", "storageReads", "balanceChanges", "nonceChanges", "codeChanges"}

// checkBALJSON verifies the basic shape of an eth_getBlockAccessList response.
// Full schema validation is performed by speccheck.
func checkBALJSON(raw json.RawMessage) error {
	var accounts []map[string]json.RawMessage
	if err := json.Unmarshal(raw, &accounts); err != nil {
		return fmt.Errorf("response is not an array of account changes: %v", err)
	}
	if len(accounts) == 0 {
		return fmt.Errorf("block access list is empty")
	}
	for i, acct := range accounts {
		for _, field := range balAccountFields {
			if _, ok := acct[field]; !ok {
				return fmt.Errorf("account %d missing field %q", i, field)
			}
		}
		if len(acct) != len(balAccountFields) {
			return fmt.Errorf("account %d has unexpected extra fields", i)
		}
	}
	return nil
}

// checkErrorCode verifies that err is a JSON-RPC error with the given code.
func checkErrorCode(err error, code int) error {
	if err == nil {
		return fmt.Errorf("expected error with code %d, got none", code)
	}
	var rpcErr rpc.Error
	if !errors.As(err, &rpcErr) {
		return fmt.Errorf("expected rpc error with code %d, got: %v", code, err)
	}
	if rpcErr.ErrorCode() != code {
		return fmt.Errorf("wrong error code: got %d (%v), want %d", rpcErr.ErrorCode(), err, code)
	}
	return nil
}

// EthGetBlockAccessList stores a list of all tests against the method.
var EthGetBlockAccessList = MethodTests{
	"eth_getBlockAccessList",
	[]Test{
		{
			Name:  "get-by-number",
			About: "gets the block access list of the first Amsterdam block by number",
			Run: func(ctx context.Context, t *T) error {
				b := firstBALBlock(t)
				var result json.RawMessage
				if err := t.rpc.CallContext(ctx, &result, "eth_getBlockAccessList", hexutil.EncodeUint64(b.NumberU64())); err != nil {
					return err
				}
				return checkBALJSON(result)
			},
		},
		{
			Name:  "get-by-hash",
			About: "gets the block access list of the head block by hash",
			Run: func(ctx context.Context, t *T) error {
				var result json.RawMessage
				if err := t.rpc.CallContext(ctx, &result, "eth_getBlockAccessList", t.chain.Head().Hash()); err != nil {
					return err
				}
				return checkBALJSON(result)
			},
		},
		{
			Name:  "get-latest",
			About: "gets the block access list of the latest block",
			Run: func(ctx context.Context, t *T) error {
				var result json.RawMessage
				if err := t.rpc.CallContext(ctx, &result, "eth_getBlockAccessList", "latest"); err != nil {
					return err
				}
				return checkBALJSON(result)
			},
		},
		{
			Name:  "get-pre-amsterdam",
			About: "requests the block access list of a block before the Amsterdam fork",
			Run: func(ctx context.Context, t *T) error {
				var result json.RawMessage
				err := t.rpc.CallContext(ctx, &result, "eth_getBlockAccessList", "0x0")
				return checkErrorCode(err, -32001)
			},
		},
		{
			Name:  "get-notfound",
			About: "requests the block access list of a non-existent block",
			Run: func(ctx context.Context, t *T) error {
				var result json.RawMessage
				if err := t.rpc.CallContext(ctx, &result, "eth_getBlockAccessList", "0x2710"); err != nil {
					return err
				}
				if string(result) != "null" {
					return fmt.Errorf("expected null for non-existent block, got: %s", result)
				}
				return nil
			},
		},
	},
}

// DebugGetRawBlockAccessList stores a list of all tests against the method.
var DebugGetRawBlockAccessList = MethodTests{
	"debug_getRawBlockAccessList",
	[]Test{
		{
			Name:  "get-by-number",
			About: "gets the RLP-encoded block access list of the first Amsterdam block",
			Run: func(ctx context.Context, t *T) error {
				b := firstBALBlock(t)
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlockAccessList", hexutil.EncodeUint64(b.NumberU64())); err != nil {
					return err
				}
				if h := crypto.Keccak256Hash(got); h != *b.Header().BlockAccessListHash {
					return fmt.Errorf("access list hash mismatch: got %x, header commits to %x", h, *b.Header().BlockAccessListHash)
				}
				return nil
			},
		},
		{
			Name:  "get-by-hash",
			About: "gets the RLP-encoded block access list of the head block by hash",
			Run: func(ctx context.Context, t *T) error {
				head := t.chain.Head()
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlockAccessList", head.Hash()); err != nil {
					return err
				}
				if h := crypto.Keccak256Hash(got); h != *head.Header().BlockAccessListHash {
					return fmt.Errorf("access list hash mismatch: got %x, header commits to %x", h, *head.Header().BlockAccessListHash)
				}
				return nil
			},
		},
		{
			Name:  "get-pre-amsterdam",
			About: "requests the block access list of a block before the Amsterdam fork",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "debug_getRawBlockAccessList", "0x0")
				return checkErrorCode(err, -32001)
			},
		},
		{
			Name:  "get-notfound",
			About: "requests the block access list of a non-existent block",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "debug_getRawBlockAccessList", "0x2710")
				return checkErrorCode(err, -32001)
			},
		},
	},
}
