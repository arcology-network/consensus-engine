package monaco

import (
	"time"

	"github.com/arcology-network/consensus-engine/types"
)

type BlockStore interface {
	Base() int64
	Height() int64
	Size() int64

	LoadBaseMeta() *types.BlockMeta
	LoadBlockMeta(height int64) *types.BlockMeta
	LoadBlock(height int64) *types.Block

	SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit)
	SaveBlockAsync(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit)

	PruneBlocks(height int64) (uint64, error)

	LoadBlockByHash(hash []byte) *types.Block
	LoadBlockPart(height int64, index int) *types.Part

	LoadBlockCommit(height int64) *types.Commit
	LoadSeenCommit(height int64) *types.Commit
	SaveSeenCommit(height int64, seenCommit *types.Commit) error
}

type BackendProxy interface {
	Reap(maxBytes int64, maxGas int64, height int64) (txs [][]byte, hashes [][]byte)
	AddToMempool(txs [][]byte, src string)
	ApplyTxsSync(height int64, coinbase []byte, timestamp time.Time, hashes [][]byte) []byte
	GetLocalTxsChan() chan [][]byte
	GetTxsOnBlock(height uint64) ([][]byte, error)
	CreateBlockStore() BlockStore
	CreateStateStore() interface{}
	UpdateMaxPeerHeight(height uint64)
	SwitchToConsensus()
}
